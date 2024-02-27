package raft

import (
	"bytes"
	"encoding/gob"
	"log"
)

//type Log struct {
//	//snap      *Snapshot //
//	//unstable  *Unstable
//	storage   *StorageIFace
//	committed int
//	applying  int
//	applied   int
//
//	logger *DefaultLogger
//
//	applyingEntsPaused bool
//}

type Entry struct {
	Term int
	Data interface{}
}

type CommitEntry struct {
	Term  int
	Index int
	Data  interface{}
}

// ToBytes 将 Entry 转化为二进制流
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
