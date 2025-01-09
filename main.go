package main

import (
	"fmt"
	log "go-authorization/log"
	spec "go-authorization/spec"
)

func main() {
	logServer := log.NewLog()
	record := spec.Record{Value: []byte("hello world")}
	off, err := logServer.Produce(&spec.ProduceRequest{Record: &record})
	if err != nil {
		fmt.Print(err)
	}

	result, err := logServer.Consume(&spec.ConsumeRequest{Offset: off}) 
	if err != nil {
		fmt.Print(err)
	}
	if (record.Offset == result.Offset) {
		fmt.Print("Success")
	} else {
		fmt.Print("Failure")
	}

}