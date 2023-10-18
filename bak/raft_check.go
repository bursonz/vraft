package bak

type RaftCheck interface {
	checkQuorumCommitIndex() bool
	checkQuorumVotes() bool
}

func (r *Raft) lastLogIndexAndTerm() (lastLogIndex uint64, lastLogTerm uint64) {
	if len(r.log) > 0 {
		lastIndex := len(r.log) - 1
		return uint64(lastIndex), r.log[lastIndex].Term
	} else {
		return -1, -1
	}
}

func (r *Raft) checkStateViceLeader() bool {
	if r.vraft && len(r.preOrderedPeers) != 0 {
		for _, v := range r.preOrderedPeers {
			if v == r.id {
				return true
			}
		}
	}
	return false
}

// 检查是否达到法定票数
// 非线程安全
func (r *Raft) checkQuorumVotes() bool {
	return len(r.votes)*2 > len(r.peers)+1
}

// 非线程安全
func (r *Raft) checkQuorumCommitIndex() (commit uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	matchMap := r.matchIndex
	commit = r.commitIndex
	for {
		count := 1
		commitIndex := commit + 1
		for _, v := range matchMap { //统计频次
			if commitIndex <= v {
				count++
			}
		}
		r.logger.Debugf("checkQuorumCommitIndex: %d")
		if count*2 > len(matchMap)+1 {
			commit = commitIndex // +1
			continue
		}

		if count == 1 {
			return commit
		}
	}
}
