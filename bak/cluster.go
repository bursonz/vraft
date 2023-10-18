package bak

import (
	"sync"
	"testing"
)

type Cluster struct {
	nodes   []*Node
	storage []*MapStorage

	lastLeader uint64

	// VRaft Configuration
	vraft        bool
	peersMetrics [][]float64
	criterias    []bool

	// Log Storage
	commitC []chan CommitEntry // 持久化队列
	commits [][]CommitEntry    // 已提交的日志

	network *Network

	//connected []bool // 可以被代替
	alive []bool

	l *DefaultLogger

	n  int // node number of this cluster
	t  *testing.T
	mu sync.Mutex
}

// NewCluster creates a new test Cluster, initialized with n nodes
// connected to each other.
func NewCluster() *Cluster {
	c := &Cluster{}
	c.l = NewDefaultLogger("[Cluster]")
	return nil
}

func (c *Cluster) Shutdown() {

}

// DisconnectPeer isolates peer from the Cluster
func (c *Cluster) DisconnectPeer(id uint64) {

}

// ReconnectPeer connect peer to the Cluster
func (c *Cluster) ReconnectPeer(id uint64) {

}

// CrashPeer "crashes" a server by disconnecting it from all peers and then
// asking it to shut down. We're not going to use the same server instance
// again, but its storage is retained.
// 使某个peer崩溃（关机）
func (c *Cluster) CrashPeer(id uint64) {}

// RestartPeer "restarts" a server by creating a new Node instance and giving
// it the appropriate storage, reconnecting it to peers.
// 重启peer
func (c *Cluster) RestartPeer(id uint64) {}
