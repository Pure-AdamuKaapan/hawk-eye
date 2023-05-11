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
	mountSetupEvent = "Mount Setup"
	pxDownEvent     = "PX Down"
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
	Name      string `json:"objectFullName"`
	Kind      string `json:"objectType"`
	ShortName string `json:"objectName"`
}

type Event struct {
	Name         string         `json:"eventName"`
	StartTime    uint64         `json:"start"`
	EndTime      uint64         `json:"finish"`
	StartUTC     string         `json:"startUTC"`
	EndUTC       string         `json:"finishUTC"`
	Severity     string         `json:"eventSeverity"`
	EventSources []*EventSource `json:"objects"`
}

type record interface {
	getNodeName() string
	getTimestamp() uint64
}

type volRecord interface {
	record
	getPodUID() string
	getPVName() string
}

type commonRec struct {
	timestamp uint64
	nodeName  string
}

type pxDaemonExitedRec struct {
	*commonRec
	exitCode  int
	exitDescr string
}

type pxReadyRec struct {
	*commonRec
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

func NewPXDaemonExitedRec(vals []string) (record, error) {
	// timestamp,node_name,service_status,exit_code,exit_descr
	// 1662389588,ip-10-13-112-170.pwx.dev.purestorage.com,stopped,0,)
	// 1662390578,ip-10-13-112-170.pwx.dev.purestorage.com,exited,9,; not expected)
	expected := 5
	if len(vals) != expected {
		return nil, fmt.Errorf("NewPXDaemonExitedRec: wrong number of values %d, expected %d", len(vals), expected)
	}
	exitCode, err := strconv.Atoi(vals[2])
	if err != nil {
		// TODO: handle error
		exitCode = -1
	}
	return &pxDaemonExitedRec{
		commonRec: &commonRec{
			timestamp: parseTimestamp(vals[0]),
			nodeName:  vals[1],
		},
		exitCode:  exitCode,
		exitDescr: vals[3],
	}, nil
}

func NewPXReadyRec(vals []string) (record, error) {
	// timestamp,node_name,node_id
	// 1662386697,ip-10-13-112-170.pwx.dev.purestorage.com,0cf91bb9-d4da-4058-b6ee-cd08f26f9aff
	expected := 3
	if len(vals) != expected {
		return nil, fmt.Errorf("NewPXReadyRec: wrong number of values %d, expected %d", len(vals), expected)
	}
	return &pxDaemonExitedRec{
		commonRec: &commonRec{
			timestamp: parseTimestamp(vals[0]),
			nodeName:  vals[1],
		},
	}, nil
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

func sourceMatches(left, right []*EventSource) bool {
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

func sourceLessThan(left, right volRecord) bool {
	if left.getNodeName() < right.getNodeName() {
		return true
	}
	if left.getPodUID() < right.getPodUID() {
		return true
	}
	return left.getPVName() < right.getPVName()
}

func recLess(left, right record) bool {
	if left.getNodeName() < right.getNodeName() {
		return true
	}
	return left.getTimestamp() < right.getTimestamp()
}

func volRecLess(left, right volRecord) bool {
	if sourceLessThan(left, right) {
		return true
	}
	return left.getTimestamp() < right.getTimestamp()
}

func getRecSources(rec record) []*EventSource {
	node := &EventSource{
		Name: rec.getNodeName(),
		Kind: Node,
	}
	return []*EventSource{node}
}

func getVolRecSources(rec volRecord) []*EventSource {
	node := &EventSource{
		Name: rec.getNodeName(),
		Kind: Node,
	}
	pod := &EventSource{
		Name: rec.getPodUID(), // TODO: use mapper to get pod name
		Kind: Pod,
	}
	volume := &EventSource{
		Name: rec.getPVName(),
		Kind: Volume,
	}
	return []*EventSource{node, pod, volume}
}

func GetEvents(path string, focusObj string) ([]*Event, error) {
	var events []*Event

	ret, err := getPXDownEvents(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get PX down events: %w", err)
	}
	events = append(events, ret...)

	ret, err = getMountSetupEvents(path, focusObj)
	if err != nil {
		return nil, fmt.Errorf("failed to get mount setup events: %w", err)
	}
	events = append(events, ret...)

	// post process
	for _, event := range events {
		event.StartUTC = time.Unix(int64(event.StartTime), 0).UTC().String()
		event.EndUTC = time.Unix(int64(event.EndTime), 0).UTC().String()

		// TODO: temp for demo to make events show up on the timeline
		if event.StartTime == 0 {
			event.StartTime = event.EndTime - 1
		} else if event.EndTime == 0 {
			event.EndTime = event.StartTime + 1
		} else if event.StartTime == event.EndTime {
			event.EndTime++
		}

		// shorten source names
		for _, source := range event.EventSources {
			if source.Kind == Pod {
				source.ShortName = truncateString(source.Name, 8)
			} else if source.Kind == Volume {
				source.ShortName = truncateString(source.Name, 12)
			} else {
				source.ShortName = source.Name
			}
		}

		// add portworx as a source for PX Down events
		if event.Name == pxDownEvent {
			event.EventSources = append(event.EventSources, &EventSource{
				Name:      "Portworx",
				Kind:      "Portworx",
				ShortName: "Portworx",
			})
		}
	}
	return events, nil
}

func getPXDownEvents(path string) ([]*Event, error) {
	exitedRecs := []record{}
	readyRecs := []record{}

	// PX daemon exited strings
	fPath := filepath.Join(path, "table_PX_Daemon_Exited.csv")
	exitedRecs, err := getRecs(fPath, "", NewPXDaemonExitedRec)
	if err != nil {
		return nil, err
	}

	// PX daemon ready strings
	fPath = filepath.Join(path, "table_PX_Daemon_Ready.csv")
	readyRecs, err = getRecs(fPath, "", NewPXReadyRec)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(exitedRecs, func(i, j int) bool {
		return recLess(exitedRecs[i], exitedRecs[j])
	})

	sort.SliceStable(readyRecs, func(i, j int) bool {
		return recLess(readyRecs[i], readyRecs[j])
	})

	events := []*Event{}

	var exitedIdx, readyIdx int
	for exitedIdx < len(exitedRecs) && readyIdx < len(readyRecs) {
		exitedRec := exitedRecs[exitedIdx]
		readyRec := readyRecs[readyIdx]

		exitedSources := getRecSources(exitedRec)
		readySources := getRecSources(readyRec)

		if exitedRec.getTimestamp() > readyRec.getTimestamp() {
			// ready record without the preceding exit rec
			event := &Event{
				Name:         pxDownEvent,
				EventSources: readySources,
				EndTime:      readyRec.getTimestamp(),
			}
			event.Severity = Error
			events = append(events, event)
			readyIdx++
			continue
		} else if exitedRec.getNodeName() != readyRec.getNodeName() {
			if exitedRec.getNodeName() < readyRec.getNodeName() {
				// exited without the subsequent ready for the same node
				event := &Event{
					Name:         pxDownEvent,
					EventSources: exitedSources,
					StartTime:    exitedRec.getTimestamp(),
				}
				event.Severity = Error
				events = append(events, event)
				exitedIdx++
				continue
			} else {
				// ready without previous exited for the same node
				event := &Event{
					Name:         pxDownEvent,
					EventSources: readySources,
					EndTime:      readyRec.getTimestamp(),
				}
				event.Severity = Error
				events = append(events, event)
				readyIdx++
				continue
			}
		}

		// exited and ready recs match (have the same node and ready is after exited)
		event := &Event{
			Name:         pxDownEvent,
			EventSources: exitedSources,
			StartTime:    exitedRec.getTimestamp(),
			EndTime:      readyRec.getTimestamp(),
			Severity:     Error,
		}
		events = append(events, event)
		exitedIdx++
		readyIdx++
	}
	for ; exitedIdx < len(exitedRecs); exitedIdx++ {
		// exited without ready
		rec := exitedRecs[exitedIdx]
		event := &Event{
			Name:         pxDownEvent,
			EventSources: getRecSources(rec),
			StartTime:    rec.getTimestamp(),
			Severity:     Error,
		}
		events = append(events, event)
	}
	for ; readyIdx < len(readyRecs); readyIdx++ {
		// ready without exited
		rec := readyRecs[readyIdx]
		event := &Event{
			Name:         pxDownEvent,
			EventSources: getRecSources(rec),
			EndTime:      rec.getTimestamp(),
			Severity:     Error,
		}
		events = append(events, event)
	}
	return events, nil
}

func getMountSetupEvents(path string, focusObj string) ([]*Event, error) {
	startRecs := []record{}
	finishRecs := []record{}

	// NodePublishVolume strings
	fPath := filepath.Join(path, "table_NodePublishVolume_Request.csv")
	startRecs, err := getRecs(fPath, focusObj, NewNodePublishVolumeRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("NodePublishVolume recs=%v\n", startRecs)

	// MountVolume.MountDevice succeeded strings
	fPath = filepath.Join(path, "table_MountDevice_Succeeded.csv")
	ret, err := getRecs(fPath, focusObj, NewMountDeviceSetupSucceededRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("MountDeviceSetupSucceeded recs=%v\n", ret)
	finishRecs = append(finishRecs, ret...)

	// MountVolume.SetUp failed strings
	fPath = filepath.Join(path, "table_MountVolume_Failed.csv")
	ret, err = getRecs(fPath, focusObj, NewMountVolumeSetupFailedRec)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("MountVolumeSetupFailed recs=%v\n", ret)
	finishRecs = append(finishRecs, ret...)

	sort.SliceStable(startRecs, func(i, j int) bool {
		return volRecLess(startRecs[i].(volRecord), startRecs[j].(volRecord))
	})

	sort.SliceStable(finishRecs, func(i, j int) bool {
		return volRecLess(finishRecs[i].(volRecord), finishRecs[j].(volRecord))
	})

	// for _, rec := range recs {
	// 	fmt.Printf("rec: %T:%v:%v:%v:%v\n\n\n", rec, rec.getNodeName(), rec.getPodUID(), rec.getPVName(), rec.getTimestamp())
	// }

	events := []*Event{}

	var startIdx, finishIdx int
	for startIdx < len(startRecs) && finishIdx < len(finishRecs) {
		startRec := startRecs[startIdx].(volRecord)
		finishRec := finishRecs[finishIdx].(volRecord)

		startSources := getVolRecSources(startRec)
		finishSources := getVolRecSources(finishRec)

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
			if sourceLessThan(startRec, finishRec) {
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
		rec := startRecs[startIdx].(volRecord)
		event := &Event{
			Name:         mountSetupEvent,
			EventSources: getVolRecSources(rec),
			StartTime:    rec.getTimestamp(),
		}
		event.Severity = Error
		events = append(events, event)
	}
	for ; finishIdx < len(finishRecs); finishIdx++ {
		// finish without start
		rec := finishRecs[finishIdx].(volRecord)
		event := &Event{
			Name:         mountSetupEvent,
			EventSources: getVolRecSources(rec),
			EndTime:      rec.getTimestamp(),
		}
		event.Severity = Error
		events = append(events, event)
	}
	return events, nil
}

func truncateString(str string, maxLen int) string {
	r := []rune(str)
	trunc := r[:maxLen]
	return string(trunc) + "..."
}

func getRecs(fpath string, focusObj string, recFunc valuesToRec) ([]record, error) {
	f, err := os.Open(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
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
		if focusObj != "" {
			volRec, ok := rec.(volRecord)
			if ok && volRec.getPodUID() != focusObj { // TODO: support more types for focusObj
				continue
			}
		}
		recs = append(recs, rec)
	}
	return recs, nil
}
