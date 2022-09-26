# -*- coding: utf-8 -*-

import masterfile
import os
import re
import pandas as pd

TIMESTAMP = "timestamp"
NODE_NAME = "node_name"
REST = "rest"
ROOT_DIR = "/Users/nrevanna/Downloads/CEE-540"
DB_DIR = ROOT_DIR + "/database"

class Parser:
    def __init__(self):
        if not os.path.exists(DB_DIR):
            os.mkdir(DB_DIR, 0o777)
        master = masterfile.MasterFile()
        parse_files = []
        for path, subdirs, files in os.walk(ROOT_DIR):
            for name in files:
                if name.endswith("docker.out"):
                    filename = os.path.join(path, name)
                    parse_files.append(filename)
        count = 0
        for logfile in parse_files:
            for line in open(logfile, "r"):
                for idx, logline_obj in enumerate(master.patterns):
                    m = re.findall(logline_obj.pattern, line)
                    if len(m) != 0:
                        logline_obj = found_pattern(line, logline_obj)
                        master.patterns[idx] = logline_obj
        #Creating db files
        for logline_obj in master.patterns:
            filename = logline_obj.get_table_name()
            df = pd.DataFrame(logline_obj.df, columns=list(logline_obj.df.keys()))
            df.to_csv(os.path.join(DB_DIR, filename),index = False, header=True)
        print(count)

def add_to_dict(df, col, val):
    lst = df.get(col, [])
    lst.append(val)
    df[col] = lst
    return df

def get_log_section(logline, section):
    if 

def found_pattern(line, logline_obj):
    grp = re.match(r"""(?P<timestamp>\w{3} \d{2} \d{2}:\d{2}:\d{2}) (?P<node_name>\S+) (?P<portworx_process>\S+) (?P<rest>.*)""",line)            
    df = logline_obj.df
    timestamp = grp.group(TIMESTAMP)
    node_name = grp.group(NODE_NAME)
    rest = grp.group(REST)
    
    df = add_to_dict(df, TIMESTAMP, timestamp)
    df = add_to_dict(df, NODE_NAME, node_name)
    df = add_to_dict(df, REST, rest)
    
    
    
    logline_obj.df = df
    return logline_obj
    
    

        
def main():
    cd 
    

if __name__ == "__main__":
    main()
