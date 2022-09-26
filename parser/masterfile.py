#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Sep 24 12:06:12 2022

@author: nrevanna
"""

import hashlib
import re

class LogLine:
    def __init__(self, pattern, column_names, logfile, section, status):
        self.pattern = pattern
        self.column_names = column_names
        self.logfile = logfile
        self.section = section
        self.status = status
        self.df = {}
        self.all_sections = {}
        
    def get_table_name(self):
        p = hashlib.sha256(self.pattern.encode("UTF-8")).hexdigest()
        return "table_" + p + ".csv"

    

class MasterFile:
    STATUS_FAIL = "fail"
    STATUS_PASS = "pass"

    SECTION_GLOBAL = "global"
    SECTION_MSG = "msg"
    # log files 
    PX_LOG = "docker.out"
    
    patterns = [
        #Ignore column_names and section for now
        LogLine("Started px with pid (?P<pid>\d+)", ["pid"], PX_LOG, SECTION_GLOBAL, STATUS_PASS),
        LogLine("""failed to setup internal kvdb: (?P<error_msg>[^"]+)""", ["error_msg"], PX_LOG, SECTION_MSG, STATUS_FAIL),
        LogLine("""csi.NodePublishVolume request received. VolumeID: (?P<vol_id>\d+), TargetPath: (?P<target_path>\S+)""", ["vol_id", "target_path"], PX_LOG, SECTION_MSG, STATUS_PASS)
        
    ]