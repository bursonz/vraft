package bak

import (
	"bytes"
	"encoding/gob"
	"log"
)

// persistToStorage 持久化所有的 Raft 状态到 r.storage 中
// 非线程安全,需要上锁
func (r *Raft) persistToStorage() {
	var termData bytes.Buffer
	if err := gob.NewEncoder(&termData).Encode(r.currentTerm); err != nil {
		log.Fatal(err)
	}
	r.storage.Set("currentTerm", termData.Bytes())

	var votedData bytes.Buffer
	if err := gob.NewEncoder(&votedData).Encode(r.votedFor); err != nil {
		log.Fatal(err)
	}
	r.storage.Set("votedFor", votedData.Bytes())

	var logData bytes.Buffer
	if err := gob.NewEncoder(&logData).Encode(r.log); err != nil {
		log.Fatal(err)
	}
	r.storage.Set("log", logData.Bytes())
}

// restoreFromStorage 从持久化中恢复 Raft 状态
// 非线程安全,需要上锁
func (r *Raft) restoreFromStorage() {
	if termData, found := r.storage.Get("currentTerm"); found {
		d := gob.NewDecoder(bytes.NewBuffer(termData))
		if err := d.Decode(&r.currentTerm); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("currentTerm not found in storage")
	}
	if votedData, found := r.storage.Get("votedFor"); found {
		d := gob.NewDecoder(bytes.NewBuffer(votedData))
		if err := d.Decode(&r.votedFor); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("votedFor not found in storage")
	}
	if logData, found := r.storage.Get("log"); found {
		d := gob.NewDecoder(bytes.NewBuffer(logData))
		if err := d.Decode(&r.log); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("log not found in storage")
	}
}
