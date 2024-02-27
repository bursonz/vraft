package raft

import "time"

type HardState struct {
	Term int
	Vote int
	Log  []Entry
}

type SoftState struct {
	CurrentIndex int
	CommitIndex  int
}

type StateType byte

const (
	StateDead StateType = iota
	StateCrashed
	StateFollower
	StateCandidate
	StateLeader
)

func (s StateType) String() string {
	switch s {
	case StateFollower:
		return "Follower"
	case StateCandidate:
		return "Candidate"
	case StateLeader:
		return "Leader"
	case StateCrashed:
		return "Crashed"
	default:
		panic("unreachable")
	}
}

func (r *Raft) becomeCandidate() {
	// 变更状态
	r.state = StateCandidate
	// 任期+1
	r.term += 1
	currentTerm := r.term
	// 投票自己
	r.leader = r.id
	r.votes = make(map[int]int) // 重置选票桶
	r.votes[r.id] = currentTerm
	r.checkQuorumVotes()

	// 发送RequestVote

	r.l.Infof("正在发送RequestVote")
	for _, peer := range r.peers {
		go func(id int) {
			r.mu.Lock()
			lastLogIndex, lastLogTerm := r.lastLogIndexAndTerm()
			r.mu.Unlock()

			m := Message{
				Type:     MsgRequestVote,
				From:     r.id,
				To:       id,
				Term:     currentTerm,
				LeaderId: r.id,
				LogIndex: lastLogIndex,
				LogTerm:  lastLogTerm,
				Reject:   false,
			}
			r.n.Send(m)
		}(peer.id)
	}
	r.lastElectionTime = time.Now()
	go r.runElectionTimer()
}

func (r *Raft) becomeFollower(term int, leader int) {
	r.l.Debugf("becomes Follower with term=%d; log=%v", term, r.log)
	r.state = StateFollower
	r.term = term
	r.leader = leader
	r.votes = make(map[int]int) // 重置选票

	r.lastElectionTime = time.Now()
	go r.runElectionTimer() // 开始计时
}

func (r *Raft) becomeLeader() {
	r.state = StateLeader
	r.leader = r.id
	r.votes = make(map[int]int) // 重置选票

	for _, peer := range r.peers {
		r.nextIndex[peer.id] = len(r.log) // 下一个 insertIndex
		r.matchIndex[peer.id] = -1        // leader首次上台会重置匹配的日志索引
	}

	r.l.Debugf("becomes Leader; term=%d, nextIndex=%v, matchIndex=%v; log=%v", r.term, r.nextIndex, r.matchIndex, r.log)

	// 每50ms 发送心跳消息， 若有消息发送就重置
	go func(heartbeatElapse time.Duration) {
		r.broadcastAppendEntries() // 上台后第一次广播

		t := time.NewTimer(heartbeatElapse)
		defer t.Stop()

		for {
			doSend := false
			select {
			case <-t.C:
				doSend = true

				t.Stop()
				t.Reset(heartbeatElapse)
			case _, ok := <-r.appendEntriesReadyC:
				if ok {
					doSend = true
				} else { // chan closed
					return
				}
				// 重置心跳计时器.
				if !t.Stop() {
					<-t.C
				}
				t.Reset(heartbeatElapse)
			}
			if doSend {
				r.mu.Lock()
				// 检查当前状态
				if r.state != StateLeader {
					r.mu.Unlock()
					return
				}
				r.mu.Unlock()
				// 发起广播
				r.broadcastAppendEntries()
			}
		}

	}(50 * time.Millisecond)
}
