import sys

class Report:
    cluster_files = {
        "pxctl status" : "px-status.out",
        "Bootstrap entries" : "px-boostrap-list.out",
        "PX volumes" : 'px-volumes.out'
        }
    node_files = {
        "lsblk" : "lsblk.out",
        "PX version": "px-version.out",
        "System restarts" : "last.out",
        "Node name" : "uname.out"
        }

    part2 = """<div data-role="collapsible">
      <h1>Node1</h1>
      <div data-role="collapsible">
      """

    def __init__(self, read_dir):
        self.read_dir = read_dir


    def get_command(self, title, content):
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

    def get_information_section(self):
        content = ""
        command_collapse = ""
        command_collapse += self.get_command("pxctl status", "pxctl status output")
        command_collapse += self.get_command("Bootstrap entries", " bootstrap output")
        content += self.get_single_node("Cluster", command_collapse)
        command_collapse = ""
        command_collapse += self.get_command("lsblk", "lsblk output")
        command_collapse += self.get_command("blkid", "lsblk output")
        content += self.get_single_node("Node1", command_collapse)
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
            command_collapse += self.get_command(key, must_fix[key])
        content += self.get_single_node("Must-fix", command_collapse)
        command_collapse = ""
        reco = self.get_recommended_fix()
        for key in reco.keys():
            command_collapse += self.get_command(key, reco[key])
        content += self.get_single_node("Recommended fixes", command_collapse)
        return self.get_section("Finterprints", content)

    def get_page(self):
        page_start = """<!DOCTYPE html>
    <html>
    <head>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://code.jquery.com/mobile/1.4.5/jquery.mobile-1.4.5.min.css">
    <script src="https://code.jquery.com/jquery-1.11.3.min.js"></script>
    <script src="https://code.jquery.com/mobile/1.4.5/jquery.mobile-1.4.5.min.js"></script>
    </head>
    <body>

    <div data-role="page" id="pageone">
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
    print(parser.get_page())



if __name__ == "__main__":
    main()
