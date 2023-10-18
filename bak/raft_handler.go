package bak

import "time"

type RaftHandler interface {
	// RequestVote
	handleRequestVote()
	handleRequestVoteResp()
	// AppendEntries
	handleAppendEntries()
	handleAppendEntriesResp()
	// HeartBeat
	handleHeartBeat()
	handleHeartBeatResp()
}

// HandleMessage 处理消息
func (r *Raft) HandleMessage(m Message) error {
	if r.state == StateDead {
		r.logger.Errorf("handleMessage(): 未指定接收对象")
		return nil
	}
	switch {
	case m.Term == 0: // 本地消息
	case m.Term > r.currentTerm: // 消息任期大于当前节点
		switch m.Type {
		case MsgAppend:
			r.becomeFollower(m.Term)
		}
	case m.Term < r.currentTerm: // 消息任期小于当前节点
	}

	// Log Append
	switch m.Type {
	case MsgAppend:
		r.handleAppendEntries(m)
	case MsgAppendResp:
		r.handleAppendEntriesResp(m)
	case MsgAppendForward:
		r.handleAppendEntries(m)

	}
	return nil
}

func (r *Raft) handleAppendEntries(m Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// 判断当前节点是否能够运行
	if r.state == StateDead {
		return nil
	}
	r.logger.Infof("AppendEntries: %+v", m)

	// leader's term > current term -- follower behind
	if m.Term > r.currentTerm {
		r.logger.Infof("... term 落后于 MsgAppend")
		r.becomeFollower(m.Term)
	}
	response := Message{
		Type:            MsgAppendResp,
		From:            r.id,
		To:              m.From,
		Term:            m.Term,
		VoteFor:         m.From,
		LogTerm:         0,
		LogIndex:        0,
		CommitIndex:     0,
		Entries:         nil,
		Reject:          false,
		RejectHint:      0,
		Size:            MsgNormalSize,
		PreOrderedPeers: nil,
		ForwardMessages: nil,
	}
	response.Reject = true
	// TODO
	return nil
}
func (r *Raft) handleAppendEntriesResp(m Message) {
}

// 处理 RequestVote 请求消息
func (r *Raft) handleRequestVote(m Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	lastLogIndex, lastLogTerm := r.lastLogIndexAndTerm() // 获取当前日志的last index&term

	response := Message{
		Type:            MsgRequestVoteResp,
		From:            r.id,
		To:              m.From,
		Term:            r.currentTerm,
		VoteFor:         0,
		LogTerm:         0,
		LogIndex:        0,
		CommitIndex:     0,
		Entries:         nil,
		Reject:          true,
		Size:            0,
		PreOrderedPeers: nil,
		ForwardMessages: nil,
	}
	switch {
	case m.Term > r.currentTerm:
		// 落后 peer's term > current term -- vote peer
		r.logger.Infof("... term out of date in RequestVote")
		r.becomeFollower(m.Term)
	case m.Term == r.currentTerm:
		if r.currentTerm == m.Term && //
			(r.votedFor == -1 || r.votedFor == m.VoteFor) &&
			(m.LogTerm > lastLogTerm ||
				(m.LogTerm == lastLogTerm && m.LogIndex >= lastLogIndex)) { // logIndex should be the largest
			response.Reject = false
			r.votedFor = m.VoteFor
			r.lastElectionTime = time.Now()
		}
	case m.Term < r.currentTerm:
		response.VoteFor = r.votedFor
		response.Reject = true
	}
	err := r.node.Send(response)
	if err != nil {
		return err
	}

	return nil
}

// 处理 RequestVoteResp 响应消息
func (r *Raft) handleRequestVoteResp() {

}

//func (r *Raft) handleAppendForward(m Message) {
//	// 先处理AppendEntries
//	r.handleAppendEntries(m)
//	// 然后转发Entries
//	for _, msg := range m.ForwardMessages {
//		go func(m Message) {
//			// TODO： 补齐LogEntries、转发
//		}(msg)
//	}
//}
