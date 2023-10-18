package bak

import (
	"fmt"
	"sync"
	"time"
)

type StateType int

const (
	StateFollower StateType = iota
	StateCandidate
	StateLeader
	StateViceLeader
	StateCrashed
	StateDead
)

func (s StateType) String() string {
	switch s {
	case StateFollower:
		return "Follower"
	case StateCandidate:
		return "Candidate"
	case StateLeader:
		return "Leader"
	case StateViceLeader:
		return "ViceLeader"
	case StateCrashed:
		return "Crashed"
	default:
		panic("unreachable")
	}
}

type Raft struct {
	id    uint64
	peers map[uint64]Peer

	votes map[uint64]uint64

	// VRaft
	vraft           bool     // 是否开启vraft改进
	preOrderedPeers []uint64 // index == order, value == peerId
	maximumBiasVote int      // 最大副节点尝试次数
	viceLeaders     []uint64 // index == order, value == peerId
	viceLeadersNum  int      // num of vice-leaders

	// Persistent
	// 持久化状态, HardState, 所有节点都有
	currentTerm uint64
	votedFor    uint64 // can instead by votes []int
	log         []Entry

	// Volatile
	// 易失状态, SoftState, 所有节点都有
	lastElectionTime time.Time
	state            StateType
	commitIndex      uint64 // 已提交的日志做因 达成共识的索引 lastApplied <= commitIndex <=
	lastApplied      uint64 // 最后应用日志索引 <= commitIndex

	// Volatile
	// 易失状态, SoftState, leader 专有
	nextIndex  map[uint64]uint64 // prepare to send
	matchIndex map[uint64]uint64 //

	// node 运行 Raft 算法, 并处理 Message 的收发
	node *Node
	// storage is used to persist state
	storage Storage

	// TODO
	// Channels
	commitC      chan<- CommitEntry // 提交 日志 chan
	commitReadyC chan struct{}      // 可以提交 新日志 chan

	// appendEntriesReadyC is an internal notification channel used to trigger
	// sending new AEs to followers when interesting changes occurred.
	appendEntriesReadyC chan struct{}

	logger *DefaultLogger

	// Lock
	mu sync.Mutex // sync lock
}

func NewRaft() *Raft { // TODO
	r := &Raft{}
	r.id = id
	r.storage = s
	r.logger = NewDefaultLogger("[Raft-节点-" + fmt.Sprintf("%d", id) + "] ")
	return r
}

// Stop 停止运行
func (r *Raft) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.state = StateDead
	r.logger.Warningf("正在停止运行")
	close(r.commitReadyC)
}

// Report 汇报状态
func (r *Raft) Report() (id uint64, term uint64, state StateType) { // TODO
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.id, r.currentTerm, r.state
}

// Submit 提交
func (r *Raft) Submit(data []byte) bool {
	r.mu.Lock()

	r.logger.Infof("Submit received by %v: %v", r.state, data)
	if r.state == StateLeader {
		r.log = append(r.log, Entry{Data: data, Term: r.currentTerm})

		r.persistToStorage()

		r.logger.Debugf("... log=%v", r.log) // 输出所有log
		r.mu.Unlock()
		r.appendEntriesReadyC <- struct{}{} // 通知有数据可以append
		return true
	}
	r.mu.Unlock()
	return false
}
