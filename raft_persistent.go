package raft

import (
	"bytes"
	"encoding/gob"
)

// Thread Unsafe
func (r *Raft) restoreFromStorage() {
	if termData, found, _ := r.n.s.Get("currentTerm"); found {
		d := gob.NewDecoder(bytes.NewBuffer(termData))
		if err := d.Decode(&r.term); err != nil {
			r.l.Errorf("%v", err)
		}
	} else {
		r.l.Errorf("currentTerm not found in storage")
	}
	if votedData, found, _ := r.n.s.Get("votedFor"); found {
		d := gob.NewDecoder(bytes.NewBuffer(votedData))
		if err := d.Decode(&r.leader); err != nil {
			r.l.Errorf("%v", err)
		}
	} else {
		r.l.Errorf("votedFor not found in storage")
	}
	if logData, found, _ := r.n.s.Get("log"); found {
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
	r.n.s.Set("currentTerm", termData.Bytes())

	var votedData bytes.Buffer
	if err := gob.NewEncoder(&votedData).Encode(r.leader); err != nil {
		r.l.Errorf("%v", err)
	}
	r.n.s.Set("votedFor", votedData.Bytes())

	var logData bytes.Buffer
	if err := gob.NewEncoder(&logData).Encode(r.log); err != nil {
		r.l.Errorf("%v", err)
	}
	r.n.s.Set("log", logData.Bytes())
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
			}
		}
	}
	r.l.Infof("commitChanSender done")
}
