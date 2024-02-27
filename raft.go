package vraft

import (
	"fmt"
	"sync"
	"time"
)

type RaftInterface interface {
	// Setup
	Stop()
	Submit(data []byte) bool
	CheckCommit(index int) bool

	// Step 集中处理 Message
	Step(m Message) error

	// Services
	//sendRequestVote() error // send
	//sendRequestVoteReply() error
	//sendHeartbeat() error
	//sendHeartbeatReply() error
	//sendAppendEntries(entries []Entry) error
	//sendAppendEntriesReply() error
	broadcastAppendEntries()

	// Message Handler
	handleRequestVote(m Message) // recv & handle
	handleRequestVoteReply(m Message)
	handleAppendEntries(m Message)
	handleAppendEntriesReply(m Message)
	//handleHeartbeat(m Message) // 可以用append nothing代替
	//handleHeartbeatReply(m Message)

	// Leader Election
	getElectionTimeout() time.Duration
	runElectionTimer()

	// Raft State Change
	becomeFollower(term int)
	becomeCandidate()
	becomeLeader()

	// Persistent
	restoreFromStorage()
	persistToStorage()

	// Utilities
	Report() (id int, term int, state StateType) // 报告节点情况
	lastLogIndexAndTerm() (lastLogIndex int, lastLogTerm int)
	// Check & Do something
	checkQuorumVotes()
	checkQuorumLogs()
}

type Raft struct {
	id    int
	peers map[int]Peer

	// VRaft
	preOrderedPeers []int
	viceLeaders     int // 个数

	// Persistent raft state on all servers
	term   int // current term
	leader int // last voted id
	log    []Entry

	// Volatile raft state on all servers
	state            StateType
	lastElectionTime time.Time
	currentIndex     int // 已追加的日志索引 currentIndex == len(r.log)  不必要的字段 可以从entry[]算出来
	commitIndex      int // 已提交的日志索引 达成共识的索引 applyIndex <= commitIndex <= currentIndex
	applyIndex       int // 已应用的日志索引 <= commitIndex

	// Volatile raft state on candidate
	votes map[int]int // key == peerId, value == votedTerm

	// Volatile raft state on leader (Reinitialized after election)
	nextIndex  map[int]int // 下次要插入的位置
	matchIndex map[int]int // 当前匹配的位置   一般情况下 nextIndex = matchIndex +1

	// Channels
	commitC             chan<- CommitEntry
	newCommitReadyC     chan struct{}
	appendEntriesReadyC chan struct{}

	electionTimeout time.Duration

	n  *Node   // node 运行 Raft 算法, 并处理 Message 的收、发、应用
	s  Storage // storage is used to persist state
	l  *DefaultLogger
	mu sync.Mutex
}

func NewRaft(id int, peers map[int]Peer, n *Node, s Storage,
	readyC <-chan interface{},
	commitC chan<- CommitEntry) *Raft {
	r := &Raft{
		id:                  id,
		peers:               peers,
		leader:              -1,
		currentIndex:        -1,
		commitIndex:         -1,
		applyIndex:          -1,
		state:               StateFollower,
		nextIndex:           make(map[int]int),
		matchIndex:          make(map[int]int),
		votes:               make(map[int]int),
		commitC:             commitC,
		newCommitReadyC:     make(chan struct{}, 128),
		appendEntriesReadyC: make(chan struct{}, 1),
		n:                   n,
		s:                   s,
		l:                   NewDefaultLogger(fmt.Sprintf("[Raft-%d ]", id)),
	}

	if r.s.HasData() {
		r.restoreFromStorage()
	}

	go func() {
		<-readyC
		r.mu.Lock()
		r.lastElectionTime = time.Now()
		r.mu.Unlock()
		r.runElectionTimer()
	}()

	// go 检查commitC
	go r.commitChanSender()
	return r
}

func (r *Raft) Report() (id int, term int, state StateType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.id, r.term, r.state
}

func (r *Raft) Submit(data interface{}) bool {
	r.mu.Lock()
	r.l.Infof("Submit received by %v: %v", r.state, data)
	if r.state == StateLeader {
		r.log = append(r.log, Entry{Data: data, Term: r.term, Size: 100}) //TODO Size
		r.persistToStorage()
		r.l.Infof("... log=%v", r.log)
		r.mu.Unlock()
		r.appendEntriesReadyC <- struct{}{} // 这个地方可以改为在node中投递tick
		return true
	}
	r.mu.Unlock()
	return false
}

func (r *Raft) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state = StateDead
	r.l.Warningf("becomes Dead")
	close(r.newCommitReadyC)
}

func (r *Raft) lastLogIndexAndTerm() (lastLogIndex int, lastLogTerm int) {
	if len(r.log) > 0 {
		lastIndex := len(r.log) - 1
		return lastIndex, r.log[lastIndex].Term
	} else {
		return -1, -1

	}
}
