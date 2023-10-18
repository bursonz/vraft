package vraft

func (r *Raft) Step(m Message) error {
	// 关机的节点不会响应任何消息
	if r.state == StateDead {
		return nil
	}

	//switch {
	//case m.Term == 0:
	//	// local message
	//case m.Term > r.term: // 让位
	//	r.becomeFollower(m.Term)
	//	r.persistToStorage()
	//case m.Term < r.term: // 过期消息，拒绝
	//case m.Term == r.term:
	//
	//}

	switch m.Type {
	case MsgRequestVote:
		r.handleRequestVote(m)
	case MsgRequestVoteResp:
		r.handleRequestVoteReply(m)
	case MsgAppend:
		r.handleAppendEntries(m)
	case MsgAppendResp:
		r.handleAppendEntriesReply(m)
	}
	return nil
}
