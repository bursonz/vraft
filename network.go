package vraft

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type NetworkInterface interface {
	// Send 发送消息到指定peer
	Send(m Message)

	// Recv 从返回的<-chan中接收消息
	Recv() chan Message

	// Disconnect 断开指定Peer的连接（向所有节点发送MsgUnreachable或MsgDisconnect）
	Disconnect(id int)
	Connect(id int)

	// CurrentBW 查看过去一秒的带宽占用
	CurrentBW() Bandwidth
}

type Conn struct {
	from, to int
}

type Drop struct {
	rate float64
}

type Delay struct {
	d    time.Duration
	rate float64
}

// Network 集群网络
type Network struct {
	rand        *rand.Rand
	mu          sync.Mutex
	isolatedMap map[int]bool // 集群连接池
	bandwidth   map[int]Bandwidth
	dropMap     map[Conn]Drop
	delayMap    map[Conn]Delay

	recvQueues map[int]chan Message // 接收队列
}

// 建立集群网络
func NewNetwork(nodes map[int]Peer, delay time.Duration) *Network {
	rn := &Network{
		rand:        rand.New(rand.NewSource(RandSeed)),
		recvQueues:  make(map[int]chan Message),
		dropMap:     make(map[Conn]Drop),
		delayMap:    make(map[Conn]Delay),
		bandwidth:   make(map[int]Bandwidth),
		isolatedMap: make(map[int]bool),
	}
	for _, peer := range nodes {
		rn.recvQueues[peer.id] = make(chan Message, 1024) // 创建接收队列
		rn.bandwidth[peer.id] = peer.bandwidth
		rn.isolatedMap[peer.id] = false
		conn := Conn{
			from: peer.id,
			to:   0,
		}
		for _, p := range nodes {
			if p.id == peer.id {
				continue
			} else {
				conn.to = p.id
				rn.dropMap[conn] = Drop{rate: 0}
				rn.delayMap[conn] = Delay{
					d:    delay,
					rate: 1,
				}
			}
		}
	}

	return rn
}

// NewNodeNetwork 获得节点网络
func (net *Network) NewNodeNetwork(id int) *NodeNetwork {
	return &NodeNetwork{
		net:             net,
		id:              id,
		connPool:        make(map[int]bool),       // 节点的连接池
		bw:              net.bandwidth[id],        // 带宽
		sendC:           make(chan Message, 1024), //发送队列
		lastCountBytes:  0,                        // 从现在统计流量
		lastCheckBWTime: time.Now(),               // 从现在计时
		l:               NewDefaultLogger(fmt.Sprintf("[NNet-%d]", id)),
	}
}

// send 模拟数据包在网络中的传输过程
// 模拟drop, 模拟delay, toC <- msg
func (net *Network) send(m Message) {
	net.mu.Lock()
	toC := net.recvQueues[m.To]
	if net.isolatedMap[m.To] {
		toC = nil
	}
	drop := net.dropMap[Conn{m.From, m.To}]
	dl := net.delayMap[Conn{m.From, m.To}] // from == to 时，应该为0
	net.mu.Unlock()

	if toC == nil {
		return
	}
	if drop.rate != 0 && net.rand.Float64() < drop.rate {
		return
	}
	if dl.d != 0 {
		if dl.rate == 1 {
			time.Sleep(time.Duration(dl.d))
		} else if net.rand.Float64() < dl.rate {
			rd := net.rand.Int63n(int64(dl.d))
			time.Sleep(time.Duration(rd))
		}
	}

	// use marshal/unmarshal to copy message to avoid data race.
	newMsg := m.clone()

	select {
	case toC <- newMsg:
		//default:
		// drop messages when the receiver queue is full.
		//TODO:应该队满后应该扔掉消息吗
	}
}

// recvFrom 返回目标
func (net *Network) recvFrom(from int) chan Message {
	net.mu.Lock()
	fromC := net.recvQueues[from]
	net.mu.Unlock()
	return fromC
}

func (net *Network) drop(from, to int, rate float64) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.dropMap[Conn{from, to}] = Drop{rate}
}

func (net *Network) delay(from, to int, d time.Duration, rate float64) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.delayMap[Conn{from, to}] = Delay{d, rate}
}

func (net *Network) disconnect(id int) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.isolatedMap[id] = true
}

func (net *Network) connect(id int) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.isolatedMap[id] = false
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
			nn.lastCountBytes += m.Size // 记录传输流量
			nn.net.send(m)              //
			nn.mu.Unlock()
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

// Bandwidth 带宽的结构体定义,bit数,每ms时间,unit换算单位(默认1000)
type Bandwidth struct {
	bits float64
	ms   float64
	unit float64
}

func NewBandwidth(bits float64, ms float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bits,
		ms:   ms,
		unit: unit,
	}
}

func NewBandwidthWithBitsAndMs(bits float64, ms float64, u ...float64) Bandwidth {
	// 检查换算单位
	unit := float64(1000)
	if len(u) != 0 {
		unit = u[0]
	}
	return Bandwidth{
		bits: bits,
		ms:   ms,
		unit: unit,
	}
}

// bits per ms
func (b *Bandwidth) bpms() float64 {
	return b.bits / b.ms
}

// K bits per ms
func (b *Bandwidth) Kbpms() float64 {
	return b.bpms() / b.unit
}

// M bits per ms
func (b *Bandwidth) Mbpms() float64 {
	return b.Kbpms() / b.unit
}

// bits per ms
func (b *Bandwidth) bps() float64 {
	return b.bpms() * 1000
}

// K bits per ms
func (b *Bandwidth) Kbps() float64 {
	return b.bps() / b.unit
}

// M bits per ms
func (b *Bandwidth) Mbps() float64 {
	return b.Kbps() / b.unit
}

// Bytes per ms
func (b *Bandwidth) Bpms() float64 {
	return b.bpms() / 8
}

// K Bytes per ms
func (b *Bandwidth) KBpms() float64 {
	return b.Bpms() / b.unit
}

// M Bytes per ms
func (b *Bandwidth) MBpms() float64 {
	return b.KBpms() / b.unit
}

// Bytes per s
func (b *Bandwidth) Bps() float64 {
	return b.Bpms() * 1000
}

// K Bytes per s
func (b *Bandwidth) KBps() float64 {
	return b.Bps() / b.unit
}

// M Bytes per s
func (b *Bandwidth) MBps() float64 {
	return b.KBps() / b.unit
}

//// Bandwidth 表示每秒传输速率
//type Bandwidth struct {
//	toMb         float64
//	toMbPerMs    float64
//	toKb         float64
//	toKbPerMs    float64
//	toBits       float64
//	toBitsPerMs  float64
//	toMB         float64
//	toMBPerMs    float64
//	toKB         float64
//	toKBPerMs    float64
//	toBytes      float64
//	toBytesPerMs float64
//}
//
//func NewBandwidth(n uint64,) Bandwidth { // 1Mbps = 1000 Kbps = 1000000 bps
//	return Bandwidth{
//		toMb:         n,
//		toKb:         n * 1000,
//		toBits:       n * 1000 * 1000,
//		toMB:         n / 8,
//		toKB:         n * 1000 / 8,
//		toBytes:      n * 1000 * 1000 / 8,
//		toBytesPerMs: n * 1000 / 8,
//	}
//}
//
//func NewBandwidthFromBitsAndMs(bits float64, ms float64) Bandwidth {
//	return NewBandwidth(bits / 1000 / 1000)
//}
