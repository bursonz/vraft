package vraft

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Entry struct {
	Term int

	Data interface{}
	Size int // how many bytes per Entry
	//Type  EntryType
}

//type LogEntry struct {
//	Data interface{}
//	Term int
//	Size int
//}

type CommitEntry struct {
	Term  int
	Index int
	Data  interface{}
	Size  int // how many bytes per Entry
}

// ToBytes 将 Message 转化为二进制流
func (e *Entry) ToBytes() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(e)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		return nil
	}

	//fmt.Print(buf.Bytes())
	return buf.Bytes()
}

func (e *Entry) FromBytes(b []byte) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(e)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	//fmt.Print(m)

}
