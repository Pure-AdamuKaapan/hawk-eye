# -*- coding: utf-8 -*-

import masterfile
import os
import re
import time
import datetime
import pandas as pd

TIMESTAMP = "timestamp"
NODE_NAME = "node_name"
REST = "rest"
ROOT_DIR = "/Users/nrevanna/hawkeye/test_data/sim"
DB_DIR = ROOT_DIR + "/database"

class Parser:
    def __init__(self):
        if not os.path.exists(DB_DIR):
            os.mkdir(DB_DIR, 0o777)
        self.master = masterfile.MasterFile()
    
    def start(self):
        parse_files = []
        for path, subdirs, files in os.walk(ROOT_DIR):
            for name in files:
                if name.endswith("docker.out"):
                    filename = os.path.join(path, name)
                    parse_files.append(filename)
        count = 0
        for logfile in parse_files:
            for line in open(logfile, "r"):
                for idx, logline_obj in enumerate(self.master.patterns):
                    m = re.findall(logline_obj.pattern, line)
                    if len(m) != 0:
                        logline_obj = self.found_pattern(line, logline_obj)
                        self.master.patterns[idx] = logline_obj
        
        #Creating db files
        index = {}
        for logline_obj in self.master.patterns:
            filename = logline_obj.get_table_name()
            df = pd.DataFrame(logline_obj.df, columns=list(logline_obj.df.keys()))
            df.to_csv(os.path.join(DB_DIR, filename),index = False, header=True)
            index = self.add_to_dict(index, "logfile", logline_obj.logfile)
            index = self.add_to_dict(index, "pattern", logline_obj.pattern)
            index = self.add_to_dict(index, "filename", filename)

        df = pd.DataFrame(index, columns=list(index.keys()))
        df.to_csv(os.path.join(DB_DIR, "index.csv"), index = False, header=True)

    def add_to_dict(self, df, col, val):
        lst = df.get(col, [])
        lst.append(val)
        df[col] = lst
        return df


    def found_pattern(self, line, logline_obj):
        grp = re.match(r"""(?P<timestamp>\w{3} \d{2} \d{2}:\d{2}:\d{2}) (?P<node_name>\S+) (?P<portworx_process>\S+) (?P<rest>.*)""",line)            
        df = logline_obj.df
        timestamp = grp.group(TIMESTAMP)
        node_name = grp.group(NODE_NAME)
        rest = grp.group(REST)
        timestamp = time.mktime(datetime.datetime.strptime("2022 "+timestamp, "%Y %b %d %H:%M:%S").timetuple())
        df = self.add_to_dict(df, TIMESTAMP, timestamp)
        df = self.add_to_dict(df, NODE_NAME, node_name)
        f = re.finditer(logline_obj.pattern, rest)
        group_lst = [m.groupdict() for m in f]
        for each in group_lst:   
            for key in each.keys():
                df = self.add_to_dict(df, key, each[key])  
        logline_obj.df = df
        return logline_obj
        
        

        
def main():
    parser = Parser()
    parser.start()
    
    

if __name__ == "__main__":
    main()
