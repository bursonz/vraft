package vraft

import (
	"bytes"
	"encoding/gob"
)

//func EncodeDataToBytes(data interface{}) (error, []byte) {
//	var dataBuffer bytes.Buffer
//	if err := gob.NewEncoder(&dataBuffer).Encode(data); err != nil {
//		return err, nil
//	}
//	return nil, dataBuffer.Bytes()
//}
//
//func DecodeDataFromBytes(dataBytes []byte) (error, interface{}) {
//	var dataBuffer bytes.Buffer
//	var data interface{}
//	d := gob.NewDecoder(bytes.NewBuffer(dataBytes))
//	if err := d.Decode(&r.log); err != nil {
//		log.Fatal(err)
//	}
//	return nil, dataBuffer.Bytes()
//}

// Thread Unsafe
func (r *Raft) restoreFromStorage() {
	if termData, found := r.s.Get("currentTerm"); found {
		d := gob.NewDecoder(bytes.NewBuffer(termData))
		if err := d.Decode(&r.term); err != nil {
			r.l.Errorf("%v", err)
		}
	} else {
		r.l.Errorf("currentTerm not found in storage")
	}
	if votedData, found := r.s.Get("votedFor"); found {
		d := gob.NewDecoder(bytes.NewBuffer(votedData))
		if err := d.Decode(&r.vote); err != nil {
			r.l.Errorf("%v", err)
		}
	} else {
		r.l.Errorf("votedFor not found in storage")
	}
	if logData, found := r.s.Get("log"); found {
		d := gob.NewDecoder(bytes.NewBuffer(logData))
		if err := d.Decode(&r.log); err != nil {
			r.l.Errorf("%v", err)
		}
	} else {
		r.l.Errorf("log not found in storage")
	}
}

// Thread Unsafe
func (r *Raft) persistToStorage() {
	var termData bytes.Buffer
	if err := gob.NewEncoder(&termData).Encode(r.term); err != nil {
		r.l.Errorf("%v", err)
	}
	r.s.Set("currentTerm", termData.Bytes())

	var votedData bytes.Buffer
	if err := gob.NewEncoder(&votedData).Encode(r.vote); err != nil {
		r.l.Errorf("%v", err)
	}
	r.s.Set("votedFor", votedData.Bytes())

	var logData bytes.Buffer
	if err := gob.NewEncoder(&logData).Encode(r.log); err != nil {
		r.l.Errorf("%v", err)
	}
	r.s.Set("log", logData.Bytes())
}

func (r *Raft) commitChanSender() {
	for range r.newCommitReadyC {
		r.mu.Lock()
		savedTerm := r.term
		savedApplyIndex := r.applyIndex
		var entries []Entry
		if r.commitIndex > r.applyIndex {
			entries = r.log[r.applyIndex+1 : r.commitIndex+1]
			r.applyIndex = r.commitIndex
		}
		r.mu.Unlock()
		r.l.Infof("RUN commitChanSender() entries=%v, savedLastApplied=%d", entries, savedApplyIndex)

		for i, entry := range entries {
			r.commitC <- CommitEntry{
				Term:  savedTerm,
				Index: savedApplyIndex + i + 1,
				Data:  entry.Data,
				Size:  entry.Size,
			}
		}
	}
	r.l.Infof("commitChanSender done")
}
