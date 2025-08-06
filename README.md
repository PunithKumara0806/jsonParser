# Json Parser written in go

Its a toy program that Decodes & Encodes to json format similar to json library in golang. 
```console
go run jsonParser.go
```
The main function has a example with incorrect json.
More work needs to be done to make it a library.

TODO :
 -  use tags from struct to infer json name ( currently only field names are used)
 -  expose error handling & external api to call Encode & Decode functions
