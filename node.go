package raft

import (
	"strconv"
	"sync"
)

type NodeIFace interface {
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
	peers Peers

	// NetworkConfiguration
	net NetworkIFace

	// 引用资源实例
	r *Raft // 可以无锁访问id、peers等不可变信息
	s StorageIFace
	l LoggerIFace

	// Channels
	propC   chan<- interface{} // 提案队列
	commitC chan<- CommitEntry // 提交队列
	readyC  <-chan interface{} // 通知
	quitC   chan interface{}   // 退出

	// Lockers
	wg sync.WaitGroup
	mu sync.Mutex
}

func NewNode(id int, peers Peers, storage StorageIFace, network NetworkIFace,
	readyC <-chan interface{},
	commitC chan<- CommitEntry) *Node {
	n := &Node{
		id:      id,
		peers:   peers,
		net:     network,
		s:       storage,
		commitC: commitC,
		readyC:  readyC,
		quitC:   make(chan interface{}),
	}
	n.l = NewDefaultLogger("[Node-" + strconv.Itoa(id) + " ]")
	return n
}

func (n *Node) Run() {
	n.mu.Lock()
	n.r = NewRaft(n.id, n.peers, n, n.readyC, n.commitC)
	n.net.Run()
	n.mu.Unlock()

	n.l.Infof("[%v] is Running...", n.r.id)

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
	//n.peers[peer.id] = peer
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
