package vraft

// thread unsafe
func (r *Raft) CheckCommit(index int) bool {
	//TODO : 检查提交情况
	return r.commitIndex >= index
}

// thread unsafe
func (r *Raft) checkQuorumLogs() {
	savedCommitIndex := r.commitIndex
	for i := r.commitIndex + 1; i < len(r.log); i++ {
		if r.log[i].Term == r.term {
			matchCount := 1
			for _, peer := range r.peers {
				if r.matchIndex[peer.id] >= i {
					matchCount++
				}
			}
			if matchCount*2 > len(r.peers)+1 {
				r.commitIndex = i
			}
		}
	}
	if r.commitIndex != savedCommitIndex { // 有新提交
		r.l.Infof("leader sets commitIndex := %d", r.commitIndex)
		r.newCommitReadyC <- struct{}{}     // 通知应用
		r.appendEntriesReadyC <- struct{}{} // 当前批次Entries已经提交，可以再发送AE
	}
}

// thread unsafe
func (r *Raft) checkQuorumVotes() {
	r.l.Infof("正在checkQuorumVotes，当前选票：%d,%d", len(r.votes), len(r.peers))
	// TODO 应该检查votes里的term是否符合当前term
	if len(r.votes)*2 > len(r.peers)+1 {
		r.l.Infof("收到足够选票，正在变为Leader")
		r.becomeLeader()
	}
}
