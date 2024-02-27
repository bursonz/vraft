package raft

import (
	"fmt"
	"sync"
	"time"
)

// NewNodeNetwork 从NetMap获得节点网络
func (net *Network) NewNodeNetwork(id int) *NodeNetwork {
	return &NodeNetwork{
		net:             net,
		id:              id,
		connPool:        make(map[int]bool),       // 节点的连接池
		bw:              net.bandwidth[id],        // 带宽
		sendC:           make(chan Message, 1024), //发送队列
		doneC:           make(chan struct{}),
		lastCountBytes:  0,          // 从现在统计流量
		lastCheckBWTime: time.Now(), // 从现在计时
		l:               NewDefaultLogger(fmt.Sprintf("[NNet-%d]", id)),
	}
}

// NodeNetwork 节点网络
type NodeNetwork struct {
	net *Network
	id  int

	bw Bandwidth

	connPool map[int]bool

	sendC chan Message
	doneC chan struct{}

	totalCountBytes int
	lastCountBytes  int // 上次统计的流量
	lastCheckBWTime time.Time

	l  *DefaultLogger
	mu sync.Mutex
}

func (nn *NodeNetwork) Run() {
	nn.l.Debugf("Run()")
	go nn.runSendWorker()
}

func (nn *NodeNetwork) Stop() {
	close(nn.sendC)
	nn.doneC <- struct{}{}

	nn.l.Warningf("网络正在关闭")
}

func (nn *NodeNetwork) Send(m Message) {
	nn.mu.Lock()
	if nn.connPool[m.To] { // 如果可以连接
		nn.sendC <- m //排队执行
	}
	nn.mu.Unlock()
}

func (nn *NodeNetwork) runSendWorker() {
	for {
		select {
		case m := <-nn.sendC:
			nn.mu.Lock()
			//transTime := time.Duration(float64(m.Size*8)/nn.bw.bps()) * time.Millisecond // TODO:需要设置大小  size/rate = duration
			//TODO:countBytes与sleep的先后可能会影响实验结果
			//time.Sleep(transTime)       // 模拟传输延迟
			nn.lastCountBytes += len(m.ToBytes()) // 记录传输流量
			nn.mu.Unlock()
			go nn.net.send(m) //

		case <-nn.doneC:
			return
		}
	}
}

// Recv 从消息队列中接收消息
// return chan Message
func (nn *NodeNetwork) Recv() chan Message {
	return nn.net.recvFrom(nn.id)
}

func (nn *NodeNetwork) Disconnect(id int) {
	nn.mu.Lock()
	defer nn.mu.Unlock()

	//delete(nn.connPool, id)
	nn.connPool[id] = false
}

func (nn *NodeNetwork) Connect(id int) { // 连接节点到连接池内
	nn.mu.Lock()
	defer nn.mu.Unlock()
	//nn.net.isolatedMap[id] = false // 解除隔离
	nn.connPool[id] = true
}

func (nn *NodeNetwork) CountRecvBytes(size int) {
	nn.mu.Lock()
	nn.lastCountBytes += size
	nn.mu.Unlock()
}

func (nn *NodeNetwork) ResetConnPool() {
	nn.mu.Lock()
	nn.connPool = make(map[int]bool)
	nn.mu.Unlock()
}

func (nn *NodeNetwork) CurrentBW() Bandwidth {
	nn.mu.Lock()
	checkTimeElapse := float64(time.Now().Sub(nn.lastCheckBWTime).Milliseconds()) // 时间间隔
	checkCountBytes := float64(nn.lastCountBytes)                                 // 流量间隔
	nn.lastCheckBWTime = time.Now()
	nn.lastCountBytes = 0
	nn.mu.Unlock()
	return NewBandwidth(checkCountBytes*8, checkTimeElapse)
}
