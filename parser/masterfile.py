#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Sat Sep 24 12:06:12 2022

@author: nrevanna
"""

import hashlib
import re

class LogLine:
    def __init__(self, pattern, column_names, logfile, component, status):
        self.pattern = pattern
        self.column_names = column_names
        self.logfile = logfile
        self.componenet = component
        self.status = status
        self.df = {}
        
    def get_table_name(self):
        p = hashlib.sha256(self.pattern.encode("UTF-8")).hexdigest()
        return "table_" + p + ".csv"

    

class MasterFile:
    STATUS_FAIL = "fail"
    STATUS_PASS = "pass"

    COMPONENT_GLOBAL = "global"
    COMPONENT_MSG = "msg"
    # log files 
    PX_LOG = "docker.out"
    
    patterns = [
        LogLine("Started px with pid (\d+)", ["pid"], PX_LOG, COMPONENT_GLOBAL, STATUS_PASS),
        LogLine("""failed to setup internal kvdb: ([^"]+)""", ["error_msg"], PX_LOG, COMPONENT_MSG, STATUS_FAIL)
        ]