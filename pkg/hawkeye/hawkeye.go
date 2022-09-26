package hawkeye

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	mountSetupEvent = "mount setup"
)

// Event severities
const (
	Info    = "info"
	Warning = "warning"
	Error   = "error"
)

// Event sources
const (
	Node   = "node"
	Pod    = "pod"
	Volume = "volume"
)

type EventSource struct {
	Name string `json:"objectName"`
	Kind string `json:"objectType"`
}

type Event struct {
	Name         string        `json:"eventName"`
	StartTime    uint64        `json:"start"`
	EndTime      uint64        `json:"finish"`
	StartUTC     string        `json:"startUTC"`
	EndUTC       string        `json:"finishUTC"`
	Severity     string        `json:"eventSeverity"`
	EventSources []EventSource `json:"objects"`
}

type record interface {
	getNodeName() string
	getPodUID() string
	getPVName() string
	getTimestamp() uint64
}

type commonRec struct {
	timestamp uint64
	nodeName  string
}

type nodePublishVolumeRec struct {
	*commonRec
	volumeID string
	pvName   string
	podUID   string
}

type mountDeviceSetupSucceededRec struct {
	*commonRec
	pvName       string
	podUID       string
	podNamespace string
	podName      string
}

type mountVolumeSetupFailedRec struct {
	*commonRec
	pvName   string
	volumeID string
	podName  string
	podUID   string
}

type valuesToRec func(vals []string) (record, error)

func (c *commonRec) getNodeName() string {
	return c.nodeName
}

func (c *commonRec) getTimestamp() uint64 {
	return c.timestamp
}

func NewNodePublishVolumeRec(vals []string) (record, error) {
	// timestamp,node_name,vol_id,target_path
	// 1662415900.0,ip-10-13-112-170.pwx.dev.purestorage.com,920849628428829313,/var/lib/kubelet/pods/5a21d20f-cacd-43fe-be3e-194c34c673cd/volumes/kubernetes.io~csi/pvc-0d81053d-6952-404d-b213-2adea
	expected := 4
	if len(vals) != expected {
		return nil, fmt.Errorf("NewNodePublishVolumeRec: wrong number of values %d, expected %d", len(vals), expected)
	}
	targetPathRe := regexp.MustCompile(`^/var/lib/kubelet/pods/(.*)/volumes/kubernetes.io~csi/(.*)$`)
	matches := targetPathRe.FindStringSubmatch(vals[3])
	if matches == nil {
		return nil, fmt.Errorf("NewNodePublishVolumeRec: could not parse targetPath %v", vals[3])
	}

	// strip /mount from the pvName if needed (TODO: fix this in a better way)
	pvName := strings.TrimSuffix(matches[2], "/mount\"")

	return &nodePublishVolumeRec{
		commonRec: &commonRec{
			timestamp: parseTimestamp(vals[0]),
			nodeName:  vals[1],
		},
		volumeID: vals[2],
		podUID:   matches[1],
		pvName:   pvName,
	}, nil
}

func (p *nodePublishVolumeRec) getPodUID() string {
	return p.podUID
}

func (p *nodePublishVolumeRec) getPVName() string {
	return p.pvName
}

func parseTimestamp(val string) uint64 {
	// 1662415900.0
	val = strings.Split(val, ".")[0] // TODO: handle millis
	ret, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func NewMountVolumeSetupFailedRec(vals []string) (record, error) {
	// timestamp,node_name,pv_name,pod_name,UID
	// 1662408494.0,ip-10-13-112-170.pwx.dev.purestorage.com,pvc-05b19920-4646-470a-a4c8-19aa5247fe89,postgres-646f6f7487-w48fp,6552d0e5-606f-41d1-bf7f-478d7e7e60c1
	expected := 5
	if len(vals) != expected {
		return nil, fmt.Errorf("NewMountVolumeSetupFailedRec: wrong number of values %d, expected %d", len(vals), expected)
	}

	return &mountVolumeSetupFailedRec{
		commonRec: &commonRec{
			timestamp: parseTimestamp(vals[0]),
			nodeName:  vals[1],
		},
		pvName:  vals[2],
		podName: vals[3],
		podUID:  vals[4],
	}, nil
}

func (r *mountVolumeSetupFailedRec) getPodUID() string {
	return r.podUID
}

func (r *mountVolumeSetupFailedRec) getPVName() string {
	return r.pvName
}

func NewMountDeviceSetupSucceededRec(vals []string) (record, error) {
	// timestamp,node_name,pv_name,UID,device_path,pod_name
	// 1662412120.0,ip-10-13-112-170.pwx.dev.purestorage.com,pvc-a945884a-95ae-4b67-8ef1-ba3a7776de37,6441dfed-9989-4d74-abfd-0e5d3ae66995,/var/lib/kubelet/plugins/kubernetes.io/csi/pxd.portworx.com/fb7a950c5cc4988077f9656465ac58fc546cc96dd2792fedd3bbc32e0193ee19/globalmount,nginx-sharedv4-setupteardown-0-09-05-14h07m47s/nginx-6b5d97d5cb-vfp6l
	expected := 6
	if len(vals) != expected {
		return nil, fmt.Errorf("NewMountDeviceSetupSucceededRec: wrong number of values %d, expected %d", len(vals), expected)
	}

	podFullNameParts := strings.Split(vals[5], "/")
	if len(podFullNameParts) != 2 {
		return nil, fmt.Errorf("NewMountDeviceSetupSucceededRec: failed to split pod full name %s", vals[5])
	}

	return &mountDeviceSetupSucceededRec{
		commonRec: &commonRec{
			timestamp: parseTimestamp(vals[0]),
			nodeName:  vals[1],
		},
		pvName:       vals[2],
		podUID:       vals[3],
		podNamespace: podFullNameParts[0],
		podName:      podFullNameParts[1],
	}, nil
}

func (r *mountDeviceSetupSucceededRec) getPodUID() string {
	return r.podUID
}

func (r *mountDeviceSetupSucceededRec) getPVName() string {
	return r.pvName
}

func sourceMatches(left, right []EventSource) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		// kind should match. But just in case.
		if left[i].Kind != right[i].Kind || left[i].Name != right[i].Name {
			return false
		}
	}
	return true
}

func sourceLess(left, right record) bool {
	if left.getNodeName() < right.getNodeName() {
		return true
	}
	if left.getPodUID() < right.getPodUID() {
		return true
	}
	return left.getPVName() < right.getPVName()
}

func recLess(left, right record) bool {
	if sourceLess(left, right) {
		return true
	}
	return left.getTimestamp() < right.getTimestamp()
}

func getSources(rec record) []EventSource {
	return []EventSource{
		{
			Name: rec.getNodeName(),
			Kind: Node,
		},
		{
			Name: rec.getPodUID(), // TODO: use mapper to get pod name
			Kind: Pod,
		},
		{
			Name: rec.getPVName(),
			Kind: Volume,
		},
	}
}

func GetEvents(path string, focusObj string) ([]*Event, error) {
	startRecs := []record{}
	finishRecs := []record{}

	// NodePublishVolume strings
	fPath := filepath.Join(path, "table_855d5500c4d4597a688b4ddbb81272e20fc4630c0d8727ec82ca7849dd387b12.csv")
	startRecs, err := getRecs(fPath, focusObj, NewNodePublishVolumeRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("NodePublishVolume recs=%v\n", startRecs)

	// MountVolume.MountDevice succeeded strings
	fPath = filepath.Join(path, "table_1fe7a28925e8967658c591af15413e4c54171179e613ef9b463daa06cc94dc61.csv")
	ret, err := getRecs(fPath, focusObj, NewMountDeviceSetupSucceededRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("MountDeviceSetupSucceeded recs=%v\n", ret)
	finishRecs = append(finishRecs, ret...)

	// MountVolume.SetUp failed strings
	fPath = filepath.Join(path, "table_d3043b4ae27f9ba9811e5c0c4af7b11a01983fbe23c221341eea698bbcfba922.csv")
	ret, err = getRecs(fPath, focusObj, NewMountVolumeSetupFailedRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("MountVolumeSetupFailed recs=%v\n", ret)
	finishRecs = append(finishRecs, ret...)

	sort.SliceStable(startRecs, func(i, j int) bool {
		return recLess(startRecs[i], startRecs[j])
	})

	sort.SliceStable(finishRecs, func(i, j int) bool {
		return recLess(finishRecs[i], finishRecs[j])
	})

	// for _, rec := range recs {
	// 	fmt.Printf("rec: %T:%v:%v:%v:%v\n\n\n", rec, rec.getNodeName(), rec.getPodUID(), rec.getPVName(), rec.getTimestamp())
	// }

	events := []*Event{}

	var startIdx, finishIdx int
	for startIdx < len(startRecs) && finishIdx < len(finishRecs) {
		startRec := startRecs[startIdx]
		finishRec := finishRecs[finishIdx]

		startSources := getSources(startRec)
		finishSources := getSources(finishRec)

		if startRec.getTimestamp() > finishRec.getTimestamp() {
			// finish without a start
			event := &Event{
				Name:         mountSetupEvent,
				EventSources: finishSources,
				EndTime:      finishRec.getTimestamp(),
			}
			event.Severity = Error
			events = append(events, event)
			finishIdx++
			continue
		} else if !sourceMatches(startSources, finishSources) {
			if sourceLess(startRec, finishRec) {
				// start without finish
				event := &Event{
					Name:         mountSetupEvent,
					EventSources: startSources,
					StartTime:    startRec.getTimestamp(),
				}
				event.Severity = Error
				events = append(events, event)
				startIdx++
				continue
			} else {
				// finish without a start
				event := &Event{
					Name:         mountSetupEvent,
					EventSources: finishSources,
					EndTime:      finishRec.getTimestamp(),
				}
				event.Severity = Error
				events = append(events, event)
				finishIdx++
				continue
			}
		}

		// start and finish recs match
		sev := Info
		if _, ok := finishRec.(*mountVolumeSetupFailedRec); ok {
			sev = Error
		}
		event := &Event{
			Name:         mountSetupEvent,
			EventSources: startSources,
			StartTime:    startRec.getTimestamp(),
			EndTime:      finishRec.getTimestamp(),
			Severity:     sev,
		}
		events = append(events, event)
		startIdx++
		finishIdx++
	}
	for ; startIdx < len(startRecs); startIdx++ {
		// start without finish
		rec := startRecs[startIdx]
		event := &Event{
			Name:         mountSetupEvent,
			EventSources: getSources(rec),
			StartTime:    rec.getTimestamp(),
		}
		event.Severity = Error
		events = append(events, event)
	}
	for ; finishIdx < len(finishRecs); finishIdx++ {
		// finish without start
		rec := finishRecs[finishIdx]
		event := &Event{
			Name:         mountSetupEvent,
			EventSources: getSources(rec),
			EndTime:      rec.getTimestamp(),
		}
		event.Severity = Error
		events = append(events, event)
	}
	for _, event := range events {
		event.StartUTC = time.Unix(int64(event.StartTime), 0).UTC().String()
		event.EndUTC = time.Unix(int64(event.EndTime), 0).UTC().String()
	}
	return events, nil
}

func getRecs(fpath string, focusObj string, recFunc valuesToRec) ([]record, error) {
	f, err := os.Open(fpath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	recs := []record{}
	csvReader := csv.NewReader(f)
	first := true
	for {
		vals, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if first {
			first = false
			continue
		}
		rec, err := recFunc(vals)
		if err != nil {
			log.Fatal(err)
		}
		if focusObj != "" && rec.getPodUID() != focusObj { // TODO: support more types for focusObj
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}
