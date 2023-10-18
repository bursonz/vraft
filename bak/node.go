package bak

import (
	"sync"
)

// NodeInterface 描述node对外接口
type NodeInterface interface {
	Run()
	Shutdown()
	ConnectPeer(peer Peer)    // 连接节点
	DisconnectPeer(peer Peer) // 断开节点
	DisconnectAll()           // 断开所有节点
	Send(m Message)

	//Recv(recvC <-chan Message)

	CheckBandWidth() int
}

type Node struct {
	id    uint64
	peers map[uint64]Peer

	r *Raft
	s Storage // Persistence

	// VRaft Configuration
	vraft bool

	// Network Configuration
	net *NodeNetwork

	// Channels
	commitC chan<- CommitEntry

	readyC chan interface{}
	quit   chan interface{}

	l *DefaultLogger

	wg sync.WaitGroup
	mu sync.Mutex
}

func NewNode(id uint64, peers map[uint64]Peer, net *NodeNetwork, storage Storage, ready chan interface{}, commitC chan<- CommitEntry, vraft bool, peersMetrics map[int][]float64, criterias []bool) *Node {
	// TODO
	n := &Node{
		id:      id,
		peers:   peers,
		s:       storage,
		vraft:   false,
		net:     net,
		commitC: nil,
		readyC:  ready,
		quit:    nil,
		l:       nil,
		wg:      sync.WaitGroup{},
		mu:      sync.Mutex{},
	}
	n.id = id
	n.l = NewDefaultLogger("[Node]")
	return n
}

func (n *Node) Run() {
	n.mu.Lock()     // 进入临界区
	n.r = NewRaft() // raft TODO
	n.mu.Unlock()   // 退出临界区

	n.wg.Add(1)
	go func() { // 开始监听recvC
		defer n.wg.Done()
		var recvC chan Message
		for {
			if recvC == nil {
				recvC = n.net.Recv()
			}
			select {
			case <-n.quit: // 退出
				return
			case m := <-recvC: // 接收消息
				n.l.Debugf("Run()接受到了一个消息")
				n.wg.Add(1)
				go func(msg Message) { // 异步处理
					n.handleMessage(msg) // 处理消息
					n.wg.Done()
				}(m)
			}
		}
	}()
}

func (n *Node) Shutdown() { // TODO
	n.r.Stop()
	close(n.quit)
	n.net.Stop()
	n.wg.Wait()
}

func (n *Node) ConnectPeer(peer Peer) error { //TODO
	n.mu.Lock()
	defer n.mu.Unlock()

	// 判断是否未连接(不需要，直接map覆盖)
	// 1. 检查node中的信息
	n.peers[peer.id] = peer
	// 2. 检查node网络中的信息

	n.net.Connect(peer.id) // 应该直接修改nodeNet中的逻辑，不需要手动判断

	return nil
}

// DisconnectPeer disconnects this node from the peer identified by peer's id.
func (n *Node) DisconnectPeer(peer uint64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 1. 检查node中的信息

	delete(n.peers, peer) // 直接删除peer信息即可
	// 2. 检查node网络中的信息
	n.net.Disconnect(peer)
	return nil
}

func (n *Node) DisconnectAll() {
	n.net.ResetConnPool() // 重置连接池
}

// CheckBandWidth 检测带宽
// return:
// -1 占用低于50%，节点数量翻倍或++
// 0  占用高于50%，节点数量不变
// 1  占用高于99%，节点数量减半或--
func (n *Node) CheckBandWidth() int {
	currentBW := n.net.CurrentBW()                          // 当前剩余BW（Tokens）
	totalBW := n.net.bw                                     // 总带宽BW（cap）
	rateBW := totalBW.bps() - currentBW.bps()/totalBW.bps() // 占用率

	if rateBW <= 0.5 { // 占用小于50%
		return -1
	}
	if rateBW >= 0.95 { // 占用大于99%
		return 1
	}
	return 0
}

func (n *Node) Send(m Message) error { // TODO
	if m.To != None {                  // 消息合法
		if m.To != n.id { // 非本地消息
			n.net.Send(m)
		} else { // 本地消息
			err := n.handleMessage(m)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 处理消息, 只负责Raft算法外的部分工作
func (n *Node) handleMessage(m Message) error {

	switch m.To {
	case None:
		n.l.Errorf("handleMessage(): 未指定接收对象")
		return nil
	case n.id: // 处理本地消息
		fallthrough
	default: // 其他Raft消息抛给 Raft
		err := n.r.HandleMessage(m)
		if err != nil {
			return err
		}
	}
	return nil
}
