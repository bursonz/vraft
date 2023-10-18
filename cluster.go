package vraft

import (
	"sync"
	"testing"
	"time"
)

type ClusterInterface interface {
	Shutdown()
	DisconnectPeer(id int)
	ReconnectPeer(id int)
	CrashPeer(id int)
	RestartPeer(id int)

	SubmitToServer(id int, cmd interface{}) bool
	Propose(cmd interface{}) bool

	CheckSingleLeader() (leader int, term int)
	CheckNoLeader()
	CheckLeader() (leader int, term int)

	CheckCommitted(cmd interface{}) (count int, index int)
	CheckCommittedN(cmd interface{}, n int)
	CheckNotCommitted(cmd interface{})

	collectCommits(i int)
}
type Cluster struct {
	nodes    map[int]*Node
	nodeList []int
	storage  map[int]*MapStorage

	network *Network

	lastLeader int

	peers map[int]Peer

	connected map[int]bool // 可以被代替
	alive     map[int]bool

	// Log Storage
	commitChans map[int]chan CommitEntry // 持久化队列
	commits     map[int][]CommitEntry    // 已提交的日志

	l  *DefaultLogger
	n  int        // node number of this cluster
	t  *testing.T // test signature
	mu sync.Mutex
}

func NewCluster(t *testing.T, peers map[int]Peer, delay time.Duration) *Cluster {
	n := len(peers)
	nodes := make(map[int]*Node, n)
	storage := make(map[int]*MapStorage, n)
	connected := make(map[int]bool, n)
	alive := make(map[int]bool, n)
	commitChans := make(map[int]chan CommitEntry, n)
	commits := make(map[int][]CommitEntry, n)
	readyC := make(chan interface{})

	net := NewNetwork(peers, delay)

	nodeList := make([]int, n)

	// Create all Nodes in this cluster
	for i, peer := range peers {
		nodeList[i] = peer.id

		nodePeers := make(map[int]Peer)
		for _, p := range peers {
			if p.id != peer.id {
				nodePeers[p.id] = p
			}
		}

		storage[peer.id] = NewMapStorage()
		commitChans[peer.id] = make(chan CommitEntry)
		alive[peer.id] = true

		nodes[peer.id] = NewNode(peer.id, nodePeers, storage[peer.id],
			net.NewNodeNetwork(peer.id),
			readyC, commitChans[peer.id])
		nodes[peer.id].Run()
	}

	// Connect all peers to each other.
	for _, node := range nodes {
		for _, peer := range peers {
			if peer.id != node.id {
				node.ConnectPeer(peer)
			}
		}

		connected[node.id] = true
	}
	close(readyC)

	c := &Cluster{
		nodes:       nodes,
		nodeList:    nodeList,
		storage:     storage,
		network:     net,
		peers:       peers,
		commitChans: commitChans,
		commits:     commits,
		connected:   connected,
		alive:       alive,
		n:           n,
		t:           t,
	}

	c.l = NewDefaultLogger("[TEST] ")

	for i := 0; i < n; i++ { // 收集日志
		p := nodeList[i]
		go c.collectCommits(p)
	}
	return c
}

func (c *Cluster) Shutdown() {
	c.l.Warningf("集群正在关闭")
	// disconnect
	for i := 0; i < c.n; i++ {
		//id := c.nodeList[i]
		//c.nodes[id].DisconnectAll()
		//c.connected[id] = false
		c.nodes[i].DisconnectAll()
		c.connected[i] = false
	}
	for i := 0; i < c.n; i++ {
		//id := c.nodeList[i]
		//c.alive[id] = false
		//c.nodes[id].Shutdown()
		c.alive[i] = false
		c.nodes[i].Shutdown()
	}
	for i := 0; i < c.n; i++ {
		//id := c.nodeList[i]
		close(c.commitChans[i])
	}
}

// DisconnectPeer isolates peer from the Cluster
func (c *Cluster) DisconnectPeer(id int) {
	c.l.Testf("Disconnect %d", id)
	// let peer disconnect all nodes
	c.nodes[id].DisconnectAll()
	// let each node disconnect with peer
	for i := 0; i < c.n; i++ {
		p := c.nodeList[i]
		if p != id {
			c.nodes[p].DisconnectPeer(id)
		}
	}

	c.connected[id] = false
}

func (c *Cluster) ReconnectPeer(id int) {
	c.l.Testf("Reconnect %d", id)
	for i := 0; i < c.n; i++ {
		p := c.nodeList[i]
		if i != id && c.alive[p] {
			c.nodes[id].ConnectPeer(c.peers[p])
			c.nodes[p].ConnectPeer(c.peers[p])
		}
	}
	c.connected[id] = true
}

func (c *Cluster) CrashPeer(id int) {
	c.l.Testf("Crash %d", id)
	c.DisconnectPeer(id)
	c.alive[id] = false
	c.nodes[id].Shutdown()

	// Clear out the commits slice for the crashed server; Raft assumes the client
	// has no persistent state. Once this server comes back online it will replay
	// the whole log to us.
	c.mu.Lock()
	c.commits[id] = c.commits[id][:0]
	c.mu.Unlock()
}

func (c *Cluster) RestartPeer(id int) {
	peers := make(map[int]Peer)
	for _, peer := range c.peers {
		if peer.id != id {
			peers[peer.id] = peer
		}
	}
	ready := make(chan interface{})
	c.nodes[id] = NewNode(id, peers, c.storage[id],
		c.network.NewNodeNetwork(id),
		ready, c.commitChans[id])
	c.nodes[id].Run()
	c.ReconnectPeer(id)
	close(ready)
	c.alive[id] = true
	sleepMs(20)
}

func (c *Cluster) SubmitToServer(id int, cmd interface{}) bool {
	return c.nodes[id].r.Submit(cmd)
}

func (c *Cluster) Propose(cmd interface{}) bool {
	leader := int(0)
	leader, _ = c.CheckSingleLeader()
	return c.nodes[leader].r.Submit(cmd)
}

// CheckSingleLeader checks that only a single node thinks it's the leader.
// Returns the leader's id and term
func (c *Cluster) CheckSingleLeader() (leader int, term int) {
	// only n rounds find the leader.
	// if didn't find, return -1,-1
	for r := 0; r < 8; r++ { // 8 rounds
		leader = -1
		term = -1
		for i := 0; i < c.n; i++ {
			id := c.nodeList[i]
			if c.connected[id] {
				_, t, state := c.nodes[id].r.Report()
				if state == StateLeader {
					if leader < 0 {
						leader = id
						term = t
					} else {
						c.t.Fatalf("both %d and %d think they're leaders", leader, id)
					}
				}
			}
		}
		if leader >= 0 {
			return leader, term
		}
		sleepMs(150) // wait 150 ms for cluster's election
	}

	c.t.Fatalf("leader not found")
	return -1, -1
}

func (c *Cluster) CheckNoLeader() {
	for i := 0; i < c.n; i++ {
		id := c.nodeList[i]
		if c.connected[id] {
			_, _, state := c.nodes[id].r.Report()
			if state == StateLeader {
				c.t.Fatalf("server %d leader; want none", id)
			}
		}
	}
}

func (c *Cluster) CheckLeader() (leader int, term int) {
	for { // 8 rounds
		leader = -1
		term = -1
		for i := 0; i < c.n; i++ {
			p := c.nodeList[i]
			if c.connected[p] {
				_, t, state := c.nodes[p].r.Report()
				if state == StateLeader {
					if leader < 0 {
						leader = p
						term = t
					} else {
						c.t.Fatalf("both %d and %d think they're leaders", leader, p)
					}
				}
			}
		}
		if leader >= 0 {
			return leader, term
		}
		sleepMs(10) // wait 150 ms for cluster's election
	}
}

// CheckCommitted verifies that all connected servers have cmd committed with
// the same index. It also verifies that all commands *before* cmd in
// the commit sequence match. For this to work properly, all commands submitted
// to Raft should be unique positive ints.
// Returns the number of servers that have this command committed, and its
// log index.
// TODO: this check may be too strict. Consider tha a server can commit
// something and crash before notifying the channel. It's a valid commit but
// this checker will fail because it may not match other servers. This scenario
// is described in the paper...
//
// 查找集群中是否存在等于cmd的提交，如果存在则返回节点数量和索引
func (c *Cluster) CheckCommitted(cmd interface{}) (count int, index int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Find the length of the commits slice for connected servers.
	// 检查所有 node 的 commits 长度
	commitsLen := -1
	for i := 0; i < c.n; i++ {
		if c.connected[i] { // 对于所有网络正常的节点，commits都应该一致
			if commitsLen >= 0 {
				// If this was set already, expect the new length to be the same.
				if len(c.commits[i]) != commitsLen {
					c.t.Fatalf("commits[%d] = %d, commitsLen = %d", i, c.commits[i], commitsLen)
				}
			} else {
				commitsLen = len(c.commits[i]) // 获取有提交的节点的已提交长度
			}
		}
	}
	// now, commitsLen == committedIndex

	// Check consistency of commits from the start and to the command we're asked
	// about. This loop will return once a command=cmd is found.
	// 检查所有 node 的 每一个 commits 的内容是否一致
	for j := 0; j < commitsLen; j++ { // each col == committed index
		cmdContent := -1
		for i := 0; i < c.n; i++ { // each row == node
			p := c.nodeList[i]
			if c.connected[p] {
				content := c.commits[p][j].Data.(int) // 获取内容
				if cmdContent >= 0 {
					if content != cmdContent {
						c.t.Errorf("got %d, want %d at c.commits[%d][%d]", content, cmdContent, p, j)
					}
				} else {
					cmdContent = content
				}
			}
		}

		if cmdContent == cmd { // 找到cmd
			// Check consistency of Index.
			index = -1
			count = 0
			for i := 0; i < c.n; i++ {
				p := c.nodeList[i]
				if c.connected[p] {
					if index >= 0 && c.commits[p][j].Index != index {
						c.t.Errorf("got Index=%d, want %d at h.commits[%d][%d]", c.commits[p][j].Index, index, p, j)
					} else {
						index = c.commits[p][j].Index
					}
					count++
				}
			}
			return count, index
		}
	}

	// If there's no early return, we haven't found the command we were looking
	// for.
	c.t.Errorf("cmd=%d not found in commits", cmd)
	return -1, -1
}

// CheckCommittedN verifies that cmd was committed by exactly n connected
// servers.
// 验证包含该日志cmd的节点数量是否等于n
func (c *Cluster) CheckCommittedN(cmd interface{}, n int) {
	nodeCount, _ := c.CheckCommitted(cmd)
	if nodeCount != n {
		c.t.Errorf("CheckCommittedN got nc=%d, want %d", nodeCount, n)
	}
}

// CheckNotCommitted verifies that no command equal to cmd has been committed
// by any of the active servers yet.
// 验证任何激活的节点中没有任何相等于cmd的提交
func (c *Cluster) CheckNotCommitted(cmd interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.n; i++ {
		if c.connected[i] {
			for j := 0; j < len(c.commits[i]); j++ {
				gotCmd := c.commits[i][j].Data
				if gotCmd == cmd {
					c.t.Errorf("found %d at commits[%d][%d], expected none", cmd, i, j)
				}
			}
		}
	}
}

func (c *Cluster) collectCommits(id int) {
	for commitEntry := range c.commitChans[id] {
		c.mu.Lock()
		c.l.Testf("collectCommits(%d) got %+v", id, commitEntry)
		c.commits[id] = append(c.commits[id], commitEntry)
		c.mu.Unlock()
	}
}
