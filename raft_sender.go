package vraft

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
		Size:     0, // TODO: finish size
	}
	r.n.Send(m)
}

func (r *Raft) broadcastRequestVote(peers []Peer) {

}

func (r *Raft) sendBiasVote(id int) {
	m := Message{
		Type:     MsgBiasVote,
		From:     r.id,
		To:       id,
		Term:     r.term,
		LeaderId: id,
	}
	r.leader = id

	r.l.Infof("正在发送BiasVote：[To: %d, Term: %d]", id, r.term)
	r.send(m)
}
func (r *Raft) requestBiasVote() bool {
	// 进入
	preOrderedPeers := r.preOrderedPeers
	// 检查是否有条件发起BiasVote
	if !r.n.vraft || // 不支持vraft
		len(preOrderedPeers) == 0 { // 没有预排序序列
		r.l.Warningf("不支持BiasVote")
		return false
	}

	// 获取必要数据
	// TODO:检查应该向谁投票
	voteFor := r.leader
	if r.id == preOrderedPeers[0] {
		return false
	}
	//if voteFor != preOrderedPeers[0] {
	//	r.l.Warningf("BiasVote失效")
	//	return false
	//}
	voteFor = preOrderedPeers[0]

	//for order, peerId := range preOrderedPeers{
	//	if peerId == votedFor{
	//		votedFor= preOrderedPeers[order+1]
	//	}
	//	if
	//}

	r.sendBiasVote(voteFor)

	return true
}
func (r *Raft) broadcastBiasVote() {

}

func (r *Raft) broadcastAppendEntries() {
	r.mu.Lock()
	if r.state != StateLeader {
		r.mu.Unlock()
		return
	}
	proOrderedList := r.preOrderedPeers
	savedTerm := r.term
	//viceLeadersNum := r.viceLeaders
	//peersNum := len(r.peers)
	//TODO VRaft
	r.mu.Unlock()
	r.l.Warningf("发送PreOrderList：%v", proOrderedList)
	//if r.vraft && (len(r.preOrderedPeers) != 0 || r.preOrderedPeers == nil) {
	//	switch {
	//	case
	//	}
	//	if r.viceLeaders == len(r.peers) {
	//
	//	}
	//	return
	//}

	for _, peer := range r.peers {
		go func(id int, preOrderList []int) {
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
				Type:            MsgAppend,
				From:            r.id,
				To:              id,
				Term:            savedTerm,
				LeaderId:        r.id,
				LogTerm:         prevLogTerm,
				LogIndex:        prevLogIndex,
				CommitIndex:     r.commitIndex,
				Entries:         entries,
				Reject:          false,
				Size:            0, // TODO finish size
				PreOrderedPeers: preOrderList,
			}
			r.mu.Unlock()

			r.l.Infof("sending AppendEntries to %v: ni=%d, args=%+v", id, nextIndex, m)
			// Closed:是否应该考虑在发送完后直接将nextIndex置为下一次发送间隔？
			// No, 因为消息的收发是异步的，会造成不断重复
			r.n.Send(m)
		}(peer.id, r.preOrderedPeers)
	}
}

//func (r *Raft) sendBiasVote(id int) {
//
//}

// 发送ForwardAppend给viceLeader, 附带消息头
func (r *Raft) sendForwardAppend(id int, m ...Message) {
	//
}

func (r *Raft) send(m Message) {
	m.From = r.id
	m.Size = 10 // TODO 统一处理Size
	r.n.Send(m)
}
