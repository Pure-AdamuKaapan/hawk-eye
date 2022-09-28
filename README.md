# Hawk-eye

### Problems hawk-eye is addressing
- We waste several precious minutes in each customer escalation call trying to gather the same information about the cluster and nodes.
- Many a times customers are not proficient with pxctl commands and we end up dictating them letter by letter which wastes lot of time and end up asking for control
- While we make progress in our RCA of a customer problem, we wish to have more information about the environment and will have to go back and forth through support channels to get that information.
- While diags are useful tools, it takes an expert to tells us the correct logs to grep for. 
- Even for experts it's a long process to grep through multiple log files across different nodes and try to correlate what could be happeing.
- There is lot of tribal knowledge about best practices. It is told to the customer only when they hit an issue.

Hawk-Eye is a platform that attempts to solve the above problems and more. 

### Hawk-Eye Report:
The main output of hawk-eye is a Hawk-Eye Report. It consists of 3 sections
- Information
- Timeline
- Fingerprinting

### Hawk-Eye Intelligence:


### Hawk-Eye Live:

# Building and Running

## Parse diags to build the database
```
$ python3 parser/parser.py test_data/PWX-26783
```

## Build the go binary used to generate events
```
$ go build
```

## Generate events for a focus object (in this case a pod), the second param is the directory where the database was created
```
$ ./hawk-eye events --focus=0454503f-4399-46fc-ac26-7ada4ecaaa70 test_data/PWX-26783/database | jq '.'  > events.out
```

## Pass events through the grapher to generate a Gantt chart
```
$ TBD
```
