package vraft

import (
	"strconv"
	"sync"
)

type Peer struct {
	id int

	// Network
	bandwidth  Bandwidth // transfer capability per second - higher is better
	AvgLatency float64   // avg latency - lower is better
	// CPU, RAM, ROM - higher is better
	CPU float64
	RAM float64
	ROM float64
	// Reliability Assessment 可靠性
	//FR   float64 // Failure Rate - lower is better
	//MTBF float64 // 平均故障间隔时间 Mean Time Between Failure - higher is better
	// MTBF = MTTF + MTTR
	MTTR float64 // 平均恢复前时间 Mean Time Between Repair - lower is better
	MTTF float64 // 平均失效前时间 Mean Time To Failure - higher is better
}

// getMetrics 返回一个 []float64 {
// 带宽,延迟,CPU,RAM,ROM,FR,平均故障间隔MTBF,平均MTTR,MTTF
// }
func (p *Peer) getMetrics() []float64 {
	return []float64{
		p.bandwidth.MBps(),
		p.AvgLatency,
		p.CPU,
		p.RAM,
		p.ROM,
		//p.FR,
		//p.MTBF,
		p.MTTR,
		p.MTTF,
	}
}
func (p *Peer) getCriteria() []bool {
	return []bool{true, false, true, true, true, false, true}
}

type NodeInterface interface {
	Run()
	Shutdown()

	Send(m Message)

	ConnectPeer(peer Peer)
	DisconnectPeer(id int)
	DisconnectAll()

	CheckBandWidth() int
}

type Node struct {
	id    int
	peers map[int]Peer

	// Network Configuration
	net *NodeNetwork

	// VRaft
	vraft     bool
	criterias []bool
	// 引用实例
	r *Raft
	s Storage
	l LoggerService

	// Channels
	commitC chan<- CommitEntry

	readyC <-chan interface{}
	quitC  chan interface{}

	wg sync.WaitGroup
	mu sync.Mutex
}

func NewNode(id int, peers map[int]Peer, storage Storage, network *NodeNetwork,
	readyC <-chan interface{},
	commitC chan<- CommitEntry) *Node {
	n := &Node{
		id:      id,
		peers:   peers,
		net:     network,
		s:       storage,
		commitC: commitC,
		readyC:  readyC,
		vraft:   false,
		quitC:   make(chan interface{}),
	}
	n.l = NewDefaultLogger("[Node-" + strconv.Itoa(id) + " ]")
	n.l.Debugf("len Peers: %d", len(peers))
	return n
}
func NewVRaftNode(id int, peers map[int]Peer, storage Storage, network *NodeNetwork,
	readyC <-chan interface{},
	commitC chan<- CommitEntry,
	criterias []bool) *Node {
	n := &Node{
		id:        id,
		peers:     peers,
		net:       network,
		s:         storage,
		commitC:   commitC,
		readyC:    readyC,
		vraft:     true, // Vraft
		quitC:     make(chan interface{}),
		criterias: criterias,
	}
	n.l = NewDefaultLogger("[Node-" + strconv.Itoa(id) + " ]")
	n.l.Debugf("len Peers: %d", len(peers))
	return n
}
func (n *Node) Run() {
	n.mu.Lock()
	n.r = NewRaft(n.id, n.peers, n, n.s, n.readyC, n.commitC)
	n.net.Run()
	n.mu.Unlock()

	n.l.Infof("[%v] is Running...", n.id)

	n.wg.Add(1)
	go func() { // 开始监听recvC
		defer n.wg.Done()
		var recvC chan Message
		for {
			if recvC == nil {
				recvC = n.net.Recv()
			}
			select {
			case <-n.quitC: // 退出
				n.l.Warningf("节点正在退出")
				return
			case m := <-recvC: // 接收消息
				//n.l.Debugf("Node.Run()接受到了一个消息:%+v", m)
				// 同步处理
				n.r.Step(m)

				// 异步处理
				//	n.wg.Add(1)
				//	go func(m Message) {
				//		n.r.Step(m) // 处理消息
				//		n.wg.Done()
				//	}(m)
			}
		}

	}()

}

func (n *Node) Shutdown() {
	n.l.Warningf("节点正在关闭")
	n.net.Stop()
	n.r.Stop()
	close(n.quitC)
	n.wg.Wait()
}

func (n *Node) ConnectPeer(peer Peer) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// 判断是否未连接(不需要，直接map覆盖)
	// 1. 检查node中的信息
	n.peers[peer.id] = peer
	n.r.peers[peer.id] = peer

	// 2. 检查node网络中的信息
	n.net.Connect(peer.id) // 应该直接修改nodeNet中的逻辑，不需要手动判断

}

func (n *Node) DisconnectPeer(id int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 1. 检查node中的信息
	//delete(n.peers, id) // 直接删除peer信息即可
	//delete(n.r.peers, id)

	// 2. 检查node网络中的信息
	n.net.Disconnect(id)
}

func (n *Node) DisconnectAll() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 1. 检查node中的信息
	//n.peers = make(map[int]Peer) // 重置节点信息
	//n.r.peers = make(map[int]Peer)
	// 2. 检查node网络中的信息
	n.net.ResetConnPool() // 重置连接池

	// TODO: netmap需要处理吗？isolated？
}

func (n *Node) Send(m Message) {
	//if m.To != None { // 消息合法
	//	if m.To != n.id { // 非本地消息
	//		n.net.Send(m)
	//	} else { // 本地消息
	//		n.r.Step(m)
	//	}
	//}
	// TODO 同步？
	n.net.Send(m) // 直接发送，将本地消息处理放在nodeNet上，直接转发给本机
}

// CheckBandWidth 检测带宽
// return:
// -1 占用低于50%，节点数量翻倍或++
// 0  占用高于50%，节点数量不变
// 1  占用高于99%，节点数量减半或--
func (n *Node) CheckBandWidth() int {
	// TODO：应该使用此函数修改数量
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
