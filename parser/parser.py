# -*- coding: utf-8 -*-

import masterfile
import os
import sys

class Parser:
    def __init__(self, root_dir):
        self.ROOT_DIR = root_dir
        self.DB_DIR = os.path.join(self.ROOT_DIR, "database")
        if not os.path.exists(self.DB_DIR):
            os.mkdir(self.DB_DIR, 0o777)
        self.master = masterfile.MasterFile(self.DB_DIR)
    
    def start(self):
        files_to_parse = []
        for path, subdirs, files in os.walk(self.ROOT_DIR):
            for name in files:
                if name.endswith(self.master.PX_LOG) or name.endswith(self.master.KUBECTL_LOG):
                    filename = os.path.join(path, name)
                    files_to_parse.append(filename)
        print(files_to_parse)
        for logfile in files_to_parse:
            for line in open(logfile, "r"):
                self.master.check_if_exists(line)
        self.master.save_db_files()        

        
def main():
    if len(sys.argv) != 2:
        print("Error: Specify the directory to parse")
        exit(1)
    root_dir = sys.argv[1]
    parser = Parser(root_dir)
    parser.start()
    
    

if __name__ == "__main__":
    main()
