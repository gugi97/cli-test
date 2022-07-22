# cli-test (Assessment Test)

This is a tool to transform a Log file into a JSON or Plain Text and can specify the output file location.

## How to use it
Clone Project, and run main cli-test.exe in Command Prompt, this Example for Windows User :

To display flag information (Helper)
```
$ cli-test.exe -h
```

To convert file from log file at exactly the same dir location of input file, this example use file from samplelog directory
```
$ cli-test.exe  samplelog/ibm.log -t json
```
or this command to convert to plain text
```
$ cli-test.exe  samplelog/ibm.log -t text
```

To specify the output file location can use the -o flag
```
$ cli-test.exe  samplelog/ibm.log -o <OutputFilePath>/ibm.txt
```
```
$ cli-test.exe  samplelog/ibm.log -t json -o <OutputFilePath>/ibm.json
```

## Learning Go
I made this tool for assessment test, feel free to use the source code as a way of learning more about how to use this programming language. Especially for creating CLI tools.
