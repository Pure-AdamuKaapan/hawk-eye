import sys
import os
import shutil

class Report:

    part2 = """<div data-role="collapsible">
      <h1>Node1</h1>
      <div data-role="collapsible">
      """

    def __init__(self, read_dir):
        self.read_dir = read_dir


    def get_command_html(self, title, content):
        return """<div data-role="collapsible">
        <h1>""" + title + """</h1>
        <pre>""" + content + """</pre>
      </div>"""

    def get_single_node(self, nodename, nested_collapse):
        top = """    <div data-role="collapsible">
      <h1>""" + nodename + '</h1>'
        bottom = """
        </div>
        """
        return top + nested_collapse + bottom

    def get_section(self, title, content):
        section_start = """<div data-role="header">
        <h1>"""
        header_end = """</h1>
      </div>

      <div data-role="main" class="ui-content">"""
        section_end = "</div>"
        return section_start + title + header_end + content  + section_end

    def get_timeline_graph(self):
        #TODO
        return """<pre>
            ||||||||||||||||||||||||||||||||||
            ||||||||||||||||||||||||||||||||||
            \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\
                \\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\\
                    \\\\\\\\\\\\\\\\\\\\\\\\\\\\
            |||||||||||||||||||||||||||||||||||
            |||||||||||||||||||||||||||||||||||
                        </pre>
    """

    def get_must_fix(self):
        #TODO
        return {
            "Must fix 1": "More details",
            "Must fix 2": "More details"}
    def get_recommended_fix(self):
        #TODO
        return {
            "Recommendation 1": "More details",
            "Recommendation 2": "More details"}

    def get_cluster_command_output(self):
        cluster_files = {
            "pxctl status" : "px-status.out",
            "Bootstrap entries" : "px-boostrap-list.out",
            #"PX volumes" : 'px-volumes.out'
            }
        files_to_parse = {}
        for path, subdirs, files in os.walk(self.read_dir):
            for name in files:
                for key, value in cluster_files.items():
                    if name.endswith(value):
                        filename = os.path.join(path, name)
                        files_to_parse[key] = filename
        result = ""
        for key, value in files_to_parse.items():
            content = open(value, "r").read()
            result += self.get_command_html(key, content)
        return result

    def get_nodes_command_output(self):
        node_files = {
            "Node name" : "uname.out",
            "PX version": "px-version.out",
            "lsblk" : "lsblk.out",
            "System restarts" : "last.out",
            }
        result = {}
        hostnames = {}
        for path, subdirs, files in os.walk(self.read_dir):
            for name in files:
                for key, value in node_files.items():
                    if name.endswith(value):
                        filename = os.path.join(path, name)
                        f = open(filename, "r")
                        content = f.read()
                        f.close()
                        html = self.get_command_html(key, content)
                        dict_content = result.get(path, "")
                        dict_content += html
                        result[path] = dict_content
                        if name.endswith("uname.out"):
                            s = content.split()
                            hostnames[path] = s[1]

        output = []
        for key, value in result.items():
            output.append([hostnames[key], value])
        return output

    def get_information_section(self):
        content = ""
        command_collapse = self.get_cluster_command_output()
        content += self.get_single_node("Cluster", command_collapse)
        command_collapse =""
        cmds = self.get_nodes_command_output()
        for each in cmds:
             content += self.get_single_node(each[0], each[1])
        return self.get_section("Information", content)

    def get_timeline_section(self):
        content = ""
        content += self.get_single_node("Volume Timeline", self.get_timeline_graph())
        content += self.get_single_node("Another Timeline", "")
        return self.get_section("Timeline", content)

    def get_fngerprint_section(self):
        content = ""
        command_collapse = ""
        must_fix = self.get_must_fix()
        for key in must_fix.keys():
            command_collapse += self.get_command_html(key, must_fix[key])
        content += self.get_single_node("Know Issues", command_collapse)
        command_collapse = ""
        reco = self.get_recommended_fix()
        for key in reco.keys():
            command_collapse += self.get_command_html(key, reco[key])
        content += self.get_single_node("Recommended fixes", command_collapse)
        return self.get_section("Fingerprints", content)

    def get_page(self):
        page_start = """<!DOCTYPE html>
    <html>
    <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://code.jquery.com/mobile/1.4.5/jquery.mobile-1.4.5.min.css">
    <script src="https://code.jquery.com/jquery-1.11.3.min.js"></script>
    <script src="https://code.jquery.com/mobile/1.4.5/jquery.mobile-1.4.5.min.js"></script>
        <style>
      .container {
        display: flex;
        align-items: right;
        justify-content: right
      }
      img {
        max-width: 100%
      }
      .image {
        flex-basis: 70%;
        order: 1;
      }
      .text {
        color: #89CFF0;
        padding-right: 20px;
        font: italic 30px "Fira Sans", serif;
      }
    </style>

    </head>
    <body>



    <div data-role="page" id="pageone">
    <div class="container">
      <div class="text">
        <h1>Hawk-Eye</h1>
      </div>
      <div class="image">
        <img src="logo.jpeg">
      </div>
    </div>
    """
        page_end = """</div>

</body>
</html>"""

        return page_start + self.get_information_section() + self.get_timeline_section() + self.get_fngerprint_section() + page_end



def main():
    if len(sys.argv) != 2:
        print("Error: Specify the directory to parse")
        exit(1)
    root_dir = sys.argv[1]
    parser = Report(root_dir)
    html = parser.get_page()
    f = open(os.path.join(root_dir, "index.html"), "w")
    f.write(html)
    f.close()
    src = "./logo.jpeg"
    shutil.copyfile(src, os.path.join(root_dir, "logo.jpeg"))




if __name__ == "__main__":
    main()
