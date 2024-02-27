package raft

// Thread Unsafe
func (r *Raft) sendRequestVote(id int) {
	lastLogIndex, lastLogTerm := r.lastLogIndexAndTerm()

	m := Message{
		Type:     MsgRequestVote,
		From:     r.id,
		To:       id,
		Term:     r.term,
		LeaderId: r.id,
		LogTerm:  lastLogTerm,
		LogIndex: lastLogIndex,
		Reject:   false,
	}
	r.n.Send(m)
}

func (r *Raft) broadcastAppendEntries() {
	r.mu.Lock()
	if r.state != StateLeader {
		r.mu.Unlock()
		return
	}
	savedTerm := r.term
	r.mu.Unlock()
	for _, peer := range r.peers {
		go func(id int) {
			r.mu.Lock()
			nextIndex := r.nextIndex[id]
			prevLogIndex := nextIndex - 1
			prevLogTerm := -1
			r.l.Debugf("nextIndexMap[]: %v, nextIndex:%d, r.log[]: %+v", r.nextIndex, nextIndex, r.log)
			if prevLogIndex >= 0 {
				prevLogTerm = r.log[prevLogIndex].Term
			}
			entries := r.log[nextIndex:]

			m := Message{
				Type:        MsgAppendEntries,
				From:        r.id,
				To:          id,
				Term:        savedTerm,
				LeaderId:    r.id,
				LogTerm:     prevLogTerm,
				LogIndex:    prevLogIndex,
				CommitIndex: r.commitIndex,
				Entries:     entries,
				Reject:      false,
			}
			r.mu.Unlock()

			r.l.Infof("sending AppendEntries to %v: ni=%d, args=%+v", id, nextIndex, m)
			// Closed:是否应该考虑在发送完后直接将nextIndex置为下一次发送间隔？
			// No, 因为消息的收发是异步的，会造成不断重复
			r.n.Send(m)
		}(peer.id)
	}
}

func (r *Raft) send(m Message) {
	m.From = r.id
	r.n.Send(m)
}
