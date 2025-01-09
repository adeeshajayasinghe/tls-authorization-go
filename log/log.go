package log

import (
	"fmt"
	spec "go-authorization/spec"
	"sync"
)

type Log struct {
	mu sync.Mutex
	records []*spec.Record
}

func NewLog() *Log {
	return &Log{}
}

func (c *Log) Produce(req *spec.ProduceRequest) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record := req.Record
	record.Offset = uint64(len(c.records))
	c.records = append(c.records, record)
	return record.Offset, nil
}

func (c *Log) Consume(req *spec.ConsumeRequest) (*spec.Record, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if req.Offset >= uint64(len(c.records)) {
		return nil, fmt.Errorf("offset out of range")
	}
	record := c.records[req.Offset]
	return record, nil
}
