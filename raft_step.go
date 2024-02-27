package raft

// Step 消息处理函数，状态机的输入入口
// TODO 应该考虑按照角色进行拆分，但会影响已经拆分好的handler、sender部分
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
	case MsgAppendEntries:
		r.handleAppendEntries(m)
	case MsgAppendEntriesResp:
		r.handleAppendEntriesReply(m)
	}
	return nil
}
