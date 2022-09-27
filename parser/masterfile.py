#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Sep 24 12:06:12 2022

@author: nrevanna
"""

import hashlib
import re
import time
import datetime
import pandas as pd
import os

class LogLine:
    TIMESTAMP = "timestamp"
    NODE_NAME = "node_name"    
    REST = "rest"
    
    def __init__(self, pattern, column_names, logfile):
        self.pattern = pattern
        self.column_names = column_names
        self.logfile = logfile
        self.df = {}
        self.all_sections = {}
    
        
    def get_table_name(self):
        p = hashlib.sha256(self.pattern.encode("UTF-8")).hexdigest()
        return "table_" + p + ".csv"

    def found_a_pattern(self, line):
        grp = re.match(r"""(?P<timestamp>\w{3} \d{2} \d{2}:\d{2}:\d{2}) (?P<node_name>\S+) (?P<portworx_process>\S+) (?P<rest>.*)""",line)            
        timestamp = grp.group(self.TIMESTAMP)
        node_name = grp.group(self.NODE_NAME)
        rest = grp.group(self.REST)
        timestamp = time.mktime(datetime.datetime.strptime("2022 "+timestamp, "%Y %b %d %H:%M:%S").timetuple())
        self.add_to_dict(self.TIMESTAMP, timestamp)
        self.add_to_dict(self.NODE_NAME, node_name)
        f = re.finditer(self.pattern, rest)
        group_lst = [m.groupdict() for m in f]
        for each in group_lst:   
            for key in each.keys():
                self.add_to_dict(key, each[key])  



    def add_to_dict(self, col, val):
        lst = self.df.get(col, [])
        lst.append(val)
        self.df[col] = lst
            
    

class MasterFile:
    # log files 
    PX_LOG = "docker.out"
    KUBECTL_LOG = "kubelet.out"
    
    patterns = [
        #Ignore column_names for now. Names embedded in the regex
        #Cheatsheet
        # if the original string contains " -> \\\"
        # if the original string contains \" -> \\\\\"
        LogLine("Started px with pid (?P<pid>\d+)", ["pid"], PX_LOG),
        LogLine("""failed to setup internal kvdb: (?P<error_msg>[^"]+)""", ["error_msg"], PX_LOG),
        #Sep 05 15:11:40 ip-10-13-112-170.pwx.dev.purestorage.com portworx[1300]: time="2022-09-05T15:11:40Z" level=info msg="csi.NodePublishVolume request received. VolumeID: 920849628428829313, TargetPath: /var/lib/kubelet/pods/5a21d20f-cacd-43fe-be3e-194c34c673cd/volumes/kubernetes.io~csi/pvc-0d81053d-6952-404d-b213-2adea82bc609/mount" component=csi-driver correlation-id=b150fe93-d7ca-4293-9e55-9bc4fef2adf3 origin=csi-driver
        LogLine("""csi.NodePublishVolume request received. VolumeID: (?P<vol_id>\d+), TargetPath: (?P<target_path>\S+)""", ["vol_id", "target_path"], PX_LOG),
        #################################### kubelet.out #####################################
        #Sep 05 14:08:40 ip-10-13-112-170.pwx.dev.purestorage.com k3s[6839]: I0905 14:08:40.126274   6839 operation_generator.go:658] "MountVolume.MountDevice succeeded for volume \"pvc-251d77bd-f5ac-4c82-9aca-f767058167e4\" (UniqueName: \"kubernetes.io/csi/pxd.portworx.com^555410506377584416\") pod \"nginx-6b5d97d5cb-vfp6l\" (UID: \"6441dfed-9989-4d74-abfd-0e5d3ae66995\") device mount path \"/var/lib/kubelet/plugins/   kubernetes.io/csi/pxd.portworx.com/fb7a950c5cc4988077f9656465ac58fc546cc96dd2792fedd3bbc32e0193ee19/globalmount\"" pod="nginx-sharedv4-setupteardown-0-09-05-14h07m47s/nginx-6b5d97d5cb-vfp6l"
        LogLine("""MountVolume.MountDevice succeeded for volume (\S+) .*UID\: \\\\\"(?P<UID>[\w\-]+).* device mount path \\\\\"(?P<device_path>\S+)\\\\\"\\\".* pod=\\\"(?P<pod_name>\S+)\\\"""", [], KUBECTL_LOG),
        #LogLine('operationExecutor.VerifyControllerAttachedVolume started for volume.*UniqueName:.*\\\\\"(?P<unique_name>\S+)\\\\\".*pod.*\\\\\"(?P<pod_name>\S+)\\\\\".*\(UID: \\\\\"(?P<UID>\S+)\\\\\".*pod=\\\"(?P<pod_full_name>\S+)\\\"',[], KUBECTL_LOG),
        LogLine('operationExecutor.VerifyControllerAttachedVolume started for volume.*UniqueName:.*\\\\\"(?P<unique_name>\S+)\\\\\".*pod.*\\\\\"(?P<pod_name>\S+)\\\\\".*\(UID: \\\\\"(?P<UID>\S+)\\\\\".*pod=\\\"(?P<pod_full_name>\S+)\\\"',[], KUBECTL_LOG),

        # Clouddrive Logs
        #Sep 27 03:40:36 ip-192-168-85-22.ec2.internal portworx[28442]: time="2022-09-27T03:40:36Z" level=warning msg="Failed to find locally attached drive set: drive set not found" file="clouddrive.go:141" component=porx/px/config/identity
        LogLine("""Failed to find locally attached drive set: drive set not found""", [], PX_LOG),
        #Sep 27 03:40:37 ip-192-168-85-22.ec2.internal portworx[28442]: time="2022-09-27T03:40:37Z" level=warning msg="Unable to start as a storage node: Limit for maximum storage nodes (1) in the zone (us-east-1c) reached" file="cloud_drive.go:1633" component=porx/storage/hal/provider
        LogLine("""Unable to start as a storage node: Limit for maximum storage nodes""", [], PX_LOG),
        #Sep 27 03:40:37 ip-192-168-11-60.ec2.internal portworx[28922]: time="2022-09-27T03:40:37Z" level=info msg="Created drive vol-028b1ebe6572d7322" file="aws_storage.go:310" component=porx/storage/hal/provider/aws
        LogLine("""Created drive (?P<drive_id>\S+)\\\"""", ["drive_id"], PX_LOG),
        #Sep 27 03:40:42 ip-192-168-11-60.ec2.internal portworx[28922]: time="2022-09-27T03:40:42Z" level=info msg="Successfully attached the Drive Set" file="cloud_drive.go:2463" component=porx/storage/hal/provider
        LogLine("""Successfully attached the Drive Set""", [], PX_LOG),
        #Sep 27 03:40:38 ip-192-168-85-22.ec2.internal portworx[28442]: time="2022-09-27T03:40:38Z" level=info msg="Cloud driver provider indicated that node cannot contribute storage as: cannot create more drives as max count limit for drive sets reached.. Starting node as storage less." file="clouddrive.go:209" component=porx/px/config/identity
        LogLine("""Cloud driver provider indicated that node cannot contribute storage as: cannot create more drives as max count limit for drive sets reached.. Starting node as storage less.""", [], PX_LOG),

        # KVDB Logs
        #Sep 27 06:08:35 ip-192-168-11-60.ec2.internal portworx[4245]: time="2022-09-27T06:08:35Z" level=info msg="Bootstraping internal kvdb service." file="kvstore.go:156" component=porx/px/kvstore fn=kv-store.New id=16a99325-4507-431b-b3cc-f01f7f819f96
        LogLine("""Bootstraping internal kvdb service.""", [], PX_LOG),
        #Sep 27 16:12:09 ip-192-168-93-84.ec2.internal portworx[6902]: time="2022-09-27T16:12:09Z" level=info msg="initialized internal kvdb" file="boot.go:772" component=porx/px/boot fields.func=init func= package=boot
        LogLine("""initialized internal kvdb""", [], PX_LOG),
        #Sep 27 06:08:39 ip-192-168-85-22.ec2.internal portworx[22149]: time="2022-09-27T06:08:39Z" level=info msg="Kvdb rule storage-rule instructed to not proceed with provisioning: Storage less node detected. Node cannot act as a kvdb node. Waiting for other nodes to start kvdb. To start kvdb on this storage less node, label this node as a metadata node and restart Portworx." file="kvprovision.go:339" component=porx/px/kvstore fn=kvdb-provisioner.CanProvisionKvdb id=7250b84e-bd3b-4a64-9b48-20b1233efaa2
        LogLine("""Kvdb rule storage-rule instructed to not proceed with provisioning: : Storage less node detected""", [], PX_LOG),
        #Sep 27 06:09:15 ip-192-168-85-22.ec2.internal portworx[22149]: time="2022-09-27T06:09:15Z" level=info msg="Kvdb operating at maximum capacity. Not starting kvdb on this node." file="kvlistener.go:176" component=porx/px/kvstore fn=kv-listener.JoinComplete id=7250b84e-bd3b-4a64-9b48-20b1233efaa2
        LogLine("""Kvdb operating at maximum capacity. Not starting kvdb on this node.""", [], PX_LOG),
        #Sep 27 06:08:36 ip-192-168-11-60.ec2.internal portworx[4245]: time="2022-09-27T06:08:36Z" level=info msg="Mounting kvdb device /dev/nvme2n1 at /var/.px_kvdb" file="util.go:41" component=porx/px/kvstore/datadir
        LogLine(""" Mounting kvdb device""", [], PX_LOG),
    ]
    def __init__(self, db_dir):
        self.DB_DIR = db_dir
        
    def check_if_exists(self, line):
        for idx, logline_obj in enumerate(self.patterns):
            m = re.findall(logline_obj.pattern, line)
            '''
            if "operationExecutor.VerifyControllerAttachedVolume" in line and "operationExecutor.VerifyControllerAttachedVolume" in logline_obj.pattern:
                print(line)
                print(logline_obj.pattern)
                print(m)
                print("_____")
            '''
            if len(m) != 0:
                self.patterns[idx].found_a_pattern(line)

    
    def save_db_files(self):
        #Creating db files
        index = LogLine("", [], "")
        for logline_obj in self.patterns:
            #print(logline_obj.pattern)
            if len(logline_obj.df) == 0:
                continue
            filename = logline_obj.get_table_name()
            df = pd.DataFrame(logline_obj.df, columns=list(logline_obj.df.keys()))
            df.to_csv(os.path.join(self.DB_DIR, filename),index = False, header=True)
            index.add_to_dict("logfile", logline_obj.logfile)
            index.add_to_dict("pattern", logline_obj.pattern)
            index.add_to_dict("filename", filename)

        df = pd.DataFrame(index.df, columns=list(index.df.keys()))
        df.to_csv(os.path.join(self.DB_DIR, "index.csv"), index = False, header=True)
