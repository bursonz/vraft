package vraft

import "time"

func (r *Raft) handleRequestVote(m Message) {
	r.mu.Lock()

	lastLogIndex, lastLogTerm := r.lastLogIndexAndTerm() // 最后一条日志的 index & term
	r.l.Debugf("收到RequestVote: %+v [currentTerm=%d, votedFor=%d, log index/term=(%d, %d)]", m, r.term, r.vote, lastLogIndex, lastLogTerm)

	reply := Message{
		Type:     MsgRequestVoteResp,
		From:     r.id,
		To:       m.From,
		LeaderId: r.vote,
		Reject:   true, // 先拒绝
	}

	switch {
	case m.Term < r.term:
		// 拒绝
	case m.Term > r.term:
		r.l.Infof("发现更高任期 = %v", m.Term)
		r.becomeFollower(m.Term) // 此时term相同
		fallthrough              // 状态变更不影响投票
	case m.Term == r.term:
		if r.vote == None || r.vote == m.LeaderId { // 没有投票 或 已经投了这个人
			// 日志是否比当前：更新或相同
			if (m.LogTerm > lastLogTerm) || // 最后一条日志的term大于当前节点，或
				(m.LogTerm == lastLogTerm && // 最后一条日志的term相同，且
					m.LogIndex >= lastLogIndex) { // 最后一条日志的index大于等于当前节点
				r.vote = m.LeaderId
				reply.Reject = false
				reply.LeaderId = m.LeaderId
				r.lastElectionTime = time.Now()
			}
		}
	}

	reply.Term = r.term  // 记录最后的节点term
	r.persistToStorage() // 持久化

	r.l.Debugf("发送RequestVoteResp: %+v [currentTerm=%d, votedFor=%d]", reply, r.term, r.vote)
	r.mu.Unlock()
	r.n.Send(reply) // 发送

}

func (r *Raft) handleRequestVoteReply(m Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.l.Debugf("RequestVoteResp: %+v [currentTerm=%d, votedFor=%d]", m, r.term, r.vote)

	// 状态发生变更, 不再响应
	if r.state != StateCandidate {
		r.l.Infof("收到RequestVoteResp但状态已变更， 当前状态 = %v", r.state)
		return
	}

	switch {
	case m.Term < r.term:
		// 不执行操作
		return
	case m.Term > r.term:
		// 任期比我大
		r.l.Infof("发现更高任期 = %v", m.Term)
		r.becomeFollower(m.Term) // 变为从节点
		return
	case m.Term == r.term:
		r.l.Infof("处理选票")
		if !m.Reject { // 收到选票
			r.votes[m.From] = m.Term // 记录选票
			r.l.Infof("当前选票为 %+v", r.votes)
			r.checkQuorumVotes() // 检查选票，若合格，则变为leader
		}
	}

}

func (r *Raft) handleAppendEntries(m Message) {
	r.mu.Lock()

	r.l.Infof("AppendEntries: %+v", m)

	reply := Message{
		Type:   MsgAppendResp,
		From:   r.id,
		To:     m.From,
		Reject: true, // 默认拒绝
		Size:   0,    // TODO finish size
	}

	switch {
	case m.Term < r.term:
		// 拒绝
	case m.Term > r.term:
		// leader's term > current term -- follower behind
		// 任期大
		r.l.Infof("发现更高任期 = %v", m.Term)
		r.l.Infof("... term 落后于 MsgAppend")
		r.becomeFollower(m.Term) // 变为从节点
		fallthrough
	case m.Term == r.term:
		// 【检查状态】是否还是Follower
		if r.state != StateFollower {
			r.becomeFollower(m.Term)
		}
		r.lastElectionTime = time.Now() // 重置选举超时

		// 【匹配条目】
		if m.LogIndex == -1 || // 没有日志（重写全部日志），或
			(m.LogIndex < len(r.log) && // prevLogIndex 是否存在（index永远比len少1，len=1时，index=0）
				m.LogTerm == r.log[m.LogIndex].Term) { // prevLogTerm是否一致
			// prevLog匹配成功
			logInsertIndex := m.LogIndex + 1 // 从prev的下一个开始写
			newEntriesIndex := 0             // 相对于 m.Entries 而言的偏移量

			// 【筛检条目】
			for { // 梳理要追加的logs
				// 检查索引是否越界
				if logInsertIndex >= len(r.log) || // 插入位置不溢出r.lastLogIndex
					newEntriesIndex >= len(m.Entries) { // 插入偏移量不溢出 m.lastInsertLogIndex
					break
				}
				// 检查任期一致性，不一致的要统统覆盖
				if r.log[logInsertIndex].Term != m.Entries[newEntriesIndex].Term {
					break
				}

				logInsertIndex += 1
				newEntriesIndex += 1
			}

			// 【追加日志】
			if newEntriesIndex < len(m.Entries) { // insertIndex < length, 有entry可以追加
				r.l.Infof("... 正在插入 entries=%v 到 index=%d 处", m.Entries[newEntriesIndex:], logInsertIndex)
				r.log = append(r.log[:logInsertIndex], m.Entries[newEntriesIndex:]...) // 追加日志
				r.l.Infof("... 现在日志为 : %v", r.log)
			}

			// 【检查提交】
			if m.CommitIndex > r.commitIndex { // 有新提交
				r.commitIndex = intMin(m.CommitIndex, len(r.log)-1) // min(leaderCommitted, 本地日志数量)
				r.l.Infof("... 正在设置 commitIndex=%d", r.commitIndex)
				// TODO: 这里有锁，可以通过增加Chan的缓冲区，来避免等待
				r.newCommitReadyC <- struct{}{} // 通知后台进行apply

			}

			reply.Reject = false // 回复成功

			reply.LogIndex = len(r.log) // lastLog Index
		} else {
			// PrevLogIndex / PrevLogTerm 不匹配
			// ConflictIndex / ConflictTerm 帮助 leader 快更快了解信息
			if m.LogIndex >= len(r.log) { // prevLogIndex 超过了当前 r.log 的范围
				reply.LogIndex = len(r.log)
				reply.LogTerm = -1
			} else { // PrevLogTerm 不匹配
				reply.LogTerm = r.log[m.LogIndex].Term

				var i int
				for i = m.LogIndex - 1; i >= 0; i-- {
					if r.log[i].Term != reply.LogTerm {
						break
					}
				}
				reply.LogIndex = i + 1
			}
		}
	} // end switch

	reply.Term = r.term
	//reply.CommitIndex = r.commitIndex
	r.persistToStorage()

	r.n.Send(reply) // 异步 是否需要goroutine？
	r.mu.Unlock()
}

func (r *Raft) handleAppendEntriesReply(m Message) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.state != StateLeader {
		return
	}

	switch {
	case m.Term < r.term:
		//
	case m.Term > r.term:
		r.l.Infof("term 在AppendEntries中过期")
		r.becomeFollower(m.Term)
		return
	case m.Term == r.term:
		if !m.Reject { // 成功
			r.nextIndex[m.From] = m.LogIndex // nextIndex[peer] == len(peer.log)
			r.matchIndex[m.From] = m.LogIndex - 1
			r.l.Infof("收到 AppendEntriesReply from %d Success: nextIndex := %v, matchIndex := %v; commitIndex := %d", m.From, r.nextIndex, r.matchIndex, r.commitIndex)

			r.checkQuorumLogs() // 检查日志提交情况，若提交，则通知应用
		} else { // conflict 通过额外信息快速收敛 nextIndex
			if m.LogTerm >= 0 { // logTerm 冲突
				lastIndexOfTerm := -1
				for i := len(r.log) - 1; i >= 0; i-- {
					if r.log[i].Term == m.LogTerm {
						lastIndexOfTerm = i
						break
					}
				}
				if lastIndexOfTerm >= 0 {
					r.nextIndex[m.From] = lastIndexOfTerm + 1
				} else {
					r.nextIndex[m.From] = m.LogIndex
				}
			} else { // logIndex 冲突
				r.nextIndex[m.From] = m.LogIndex
			}
			r.l.Infof("收到 AppendEntriesReply from %d Rejected: nextIndex := %d", m.From, r.nextIndex[m.From]-1)
		}
	}
}

func (r *Raft) handleBiasVote(m Message) {

}

func (r *Raft) handleBiasVoteReply(m Message) {

}

// Leader发送ForwardAppend 给 viceLeader
// viceLeader将 暂存的AppendEntries消息头补足后转发给Follower
func (r *Raft) handleForwardAppend(m Message) {

}
