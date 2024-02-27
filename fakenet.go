package raft

import (
	"math/rand"
	"sync"
	"time"
)

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
