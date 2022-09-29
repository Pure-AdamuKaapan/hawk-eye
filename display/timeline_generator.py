import json
import time
import plotly.express as px  
import datetime
import pandas as pd
import plotly.graph_objs as go

class Events:
    def __init__(self, node, start, finish, eventName, eventSeverity, eventObject):
        self.node = node
        self.start = start
        self.finish = finish
        self.eventName=eventName
        self.eventSeverity=eventSeverity
        self.eventObject=eventObject
    def __init__(self):
        self.node = ""
        self.start = ""
        self.finish = ""
        self.eventName= ""
        self.eventSeverity= ""
        self.eventObject= ""
    def to_dict(self):
        start = datetime.datetime.fromtimestamp(int(self.start)).strftime('%Y-%m-%d %H:%M:%S')
        end = datetime.datetime.fromtimestamp(int(self.finish)).strftime('%Y-%m-%d %H:%M:%S')
        return {
            'node': self.node,
            'start': start ,
            'finish': end,
            'eventName': self.eventName,
            'eventSeverity': self.eventSeverity,
            'Event Source': self.eventObject
        }

# Opening JSON file
f = open('events.out')
  
# Returns JSON object as a dictionary
data = json.load(f)
for event in data:
    print(event)
  
# Create an event list to be used as dataframe
eventList = []

totalNodes = set()
for events in data:
    e = Events()        
    e.start=events['start']
    e.finish=events['finish']
    e.eventName=events['eventName']
    e.eventSeverity=events['eventSeverity']
    
    objects = events['objects']
    #objectList = []
    objectString = ""
    for obj in objects:
        # Update node type to separate graphs
        if obj['objectType'] == "node":
            e.node = obj['objectFullName']
            totalNodes.add(e.node )
        else:
            #objectList.append(obj['objectType'] + ":" + obj['objectName'])
            objectString +=  " " + obj['objectType'] + ":" + obj['objectName']
    e.eventObject =  objectString
    eventList.append(e.to_dict())

print(eventList)

df = pd.DataFrame(eventList)
df.head(5)

for name, group in df.groupby('node'):
    heading = "Cluster Node:" + name
    group.head()
    fig= px.timeline(group, x_start="start", x_end="finish", y="Event Source", color="eventSeverity", text="eventName")
    fig.update_layout(title_text=heading,
                  title_font_size=30)
    fig.update_xaxes(showgrid=True)

    fig.update_traces(textposition='inside')
    fig.update_traces(textposition='inside', hovertemplate = "EventSeverity:")

    #fig.update_yaxes(autorange="reversed") # otherwise tasks are listed from the bottom up
    fig.show()

fig.write_html("./sharedv4.html")
fig.write_html("./sharedv4.png")
f.close()

