package vraft

import (
	"fmt"
	"testing"
	"time"
	"vraft/topsis"
)

// 带宽，延迟，CPU，RAM，ROM，平均修复时间，平均故障间隔
func genVRaftCluster(t *testing.T, n int) *Cluster {
	//bw := map[int]Bandwidth{
	//	100: NewBandwidthWithMbps(100),
	//	50:  NewBandwidthWithMbps(50),
	//	20:  NewBandwidthWithMbps(20),
	//	10:  NewBandwidthWithMbps(10),
	//}
	peers := []Peer{
		// 带宽，延迟，CPU，RAM，ROM，平均修复时间，平均故障间隔
		{0, NewBW(100), 20, 10, 10, 9, 5, 9000},
		{1, NewBW(100), 50, 8, 4, 7, 6, 7500},
		{2, NewBW(50), 25, 8, 2, 9, 8, 10000},
		{3, NewBW(50), 20, 8, 7, 7, 8, 8000},
		{4, NewBW(20), 25, 7, 5, 4, 6, 9000},
		{5, NewBW(100), 30, 6, 4, 5, 6, 7500},
		{6, NewBW(20), 50, 8, 4, 5, 6, 10000},
		{7, NewBW(50), 25, 5, 2, 7, 8, 8000},
		{8, NewBW(20), 25, 8, 7, 3, 8, 9000},
		{9, NewBW(100), 20, 7, 4, 9, 6, 10000},
		{10, NewBW(10), 25, 6, 5, 5, 6, 7500},
		{11, NewBW(100), 50, 8, 4, 5, 6, 9000},
		{12, NewBW(50), 25, 8, 5, 2, 7, 10000},
		{13, NewBW(20), 20, 8, 2, 7, 8, 8000},
		{14, NewBW(10), 25, 8, 3, 3, 4, 10000},
		{15, NewBW(50), 25, 7, 4, 9, 6, 10000},
		{16, NewBW(100), 20, 10, 4, 5, 4, 10000},
		{17, NewBW(50), 50, 8, 4, 5, 6, 9000},
		{18, NewBW(10), 30, 8, 2, 2, 8, 8000},
		{19, NewBW(100), 25, 8, 7, 3, 8, 9000},
		{20, NewBW(10), 20, 7, 4, 9, 6, 10000},
		{21, NewBW(100), 25, 6, 4, 5, 6, 7500},
		{22, NewBW(50), 50, 8, 4, 5, 4, 10000},
		{23, NewBW(100), 25, 8, 2, 2, 8, 8000},
		{24, NewBW(50), 25, 8, 3, 3, 8, 10000},
		{25, NewBW(100), 20, 10, 7, 9, 6, 9000},
		{26, NewBW(10), 25, 6, 4, 5, 6, 8000},
		{27, NewBW(50), 20, 8, 7, 7, 6, 10000},
		{28, NewBW(20), 25, 7, 5, 2, 4, 8000},
		{29, NewBW(10), 25, 8, 5, 9, 8, 9000},
		{30, NewBW(20), 30, 8, 3, 3, 8, 7500},
		{31, NewBW(100), 25, 7, 4, 4, 6, 10000},
	}
	peersMap := make(map[int]Peer)
	for i := 0; i < n; i++ {
		// 【Random】
		//peer := Peer{
		//	id: i,
		//	bandwidth: Bandwidth{
		//		bits: 100 * 1000 * 1000, //1 Mbps = 1*1000*1000 bits per second
		//		ms:   1000,
		//		unit: 1000,
		//	},
		//	AvgLatency: 20,
		//	CPU:        100,
		//	RAM:        100,
		//	ROM:        100,
		//	//FR:         100,
		//	//MTBF:       10000,
		//	MTTR: 10000,
		//	MTTF: 10000,
		//}
		// 【Fixed】
		peer := peers[i]
		peersMap[peer.id] = peer
	}
	criterias := []bool{true, false, true, true, true, false, true}
	c := NewVRaftCluster(t, peersMap, time.Millisecond*20, criterias)
	return c
}

func TestTopsis(t *testing.T) {
	// 带宽，延迟，CPU，RAM，ROM，可靠性
	m := [][]float64{
		{100, 20, 6, 4, 5, 6},
		{100, 50, 8, 4, 5, 6},
		{20, 25, 8, 2, 2, 8},
		{30, 25, 8, 3, 3, 8},
		{40, 25, 7, 4, 4, 6},
		{100, 25, 6, 4, 5, 6},
		{100, 50, 8, 4, 5, 6},
		{20, 25, 8, 2, 2, 8},
		{30, 25, 8, 3, 3, 8},
		{40, 25, 7, 4, 4, 6},
		{100, 25, 6, 4, 5, 6},
		{100, 50, 8, 4, 5, 6},
		{20, 25, 8, 2, 2, 8},
	}

	criterias := []bool{true, false, true, true, true, false, true}
	//options := []int{1, 2, 3, 4}
	result := topsis.EntropyTopsis(m, criterias)
	fmt.Println(result)
}

func TestRaft_PreOrder(t *testing.T) {
	c := genVRaftCluster(t, 30)
	defer c.Shutdown()

	sleepMs(300)
	preOrederList := c.CheckProOrderList()
	fmt.Println("preOrderList", preOrederList)
}

func TestRaft_AppendEntries_PreOrder(t *testing.T) {
	c := genVRaftCluster(t, 5)
	defer c.Shutdown()

	sleepMs(300)
	leader, _ := c.CheckSingleLeader()
	c.SubmitToServer(leader, nil)
	sleepMs(300)
	fmt.Println("Leader's preOrderList", c.nodes[leader].r.preOrderedPeers)
	otherPeer := (leader + 1) % c.n
	fmt.Println("OtherNode's preOrderList", c.nodes[otherPeer].r.preOrderedPeers)
}

func TestRaft_PreOrder_CrashRecovery(t *testing.T) {
	c := genVRaftCluster(t, 5)
	defer c.Shutdown()

	sleepMs(300)
	leader, _ := c.CheckSingleLeader()
	c.SubmitToServer(leader, 1)
	sleepMs(600)
	fmt.Println("Leader's preOrderList", c.nodes[leader].r.preOrderedPeers)
	otherPeer := (leader + 1) % c.n
	fmt.Println("OtherNode's preOrderList", c.nodes[otherPeer].r.preOrderedPeers)
	sleepMs(600)
	c.CrashPeer(leader)
	sleepMs(300)
	newLeader, _ := c.CheckSingleLeader()

	c.SubmitToServer(newLeader, 2)
	sleepMs(300)
	c.CheckCommitted(2)
}

func TestRaft_CrashRecovery_Time(t *testing.T) {
	n := 30 // 集群节点数量

	var resultM [][]int64
	var timeResults []int64
	var timeStamp time.Time

	lastLeader := -1
	currentLeader := -1
	// Raft Cluster
	for m := 3; m <= n; {
		timeResults = make([]int64, 0)
		for i := 0; i < 10; i++ {
			c := genCluster(t, m)
			sleepMs(600) // warmup
			lastLeader, _ = c.CheckSingleLeader()
			//c.CrashPeer(leader)
			c.DisconnectPeer(lastLeader)
			timeStamp = time.Now()
			currentLeader, _ = c.CheckLeader()
			timeResults = append(timeResults, time.Since(timeStamp).Milliseconds())
			c.Shutdown()
			sleepMs(100)
			//tlog("###Crash Recovery Success! use time: %d ms", timeResults[0])
		}
		resultM = append(resultM, timeResults)
		m += 2
	}
	fmt.Print("###Crash Recovery Success! use time: ")
	//fmt.Println(timeResults)
	fmt.Println(resultM)
	fmt.Printf("lastLeader: %d, currentLeader: %d", lastLeader, currentLeader)
}

func TestRaft_CutMetrics(t *testing.T) {

}

func TestRaft_PreOrder_CrashRecovery_Time(t *testing.T) {
	n := 30 // 集群节点数量
	var resultM [][]int64
	var timeResults []int64
	//var avgResults []int64
	var timeStamp time.Time
	// Raft Cluster
	lastLeader := -1
	currentLeader := -1
	for m := 3; m <= n; {
		timeResults = make([]int64, 0)
		for i := 0; i < 10; i++ {
			c := genVRaftCluster(t, m)
			sleepMs(600) // warmup , let preOrder get ready to work
			//otherPeer := (leader - 1 + n) % c.n
			//for {
			//	if len(c.nodes[otherPeer].r.preOrderedPeers) == 0 {
			//		continue
			//	} else {
			//		break
			//	}
			//}
			lastLeader, _ = c.CheckSingleLeader()
			//c.CrashPeer(leader)
			c.DisconnectPeer(lastLeader)
			timeStamp = time.Now()
			currentLeader, _ = c.CheckLeader()
			timeResults = append(timeResults, time.Since(timeStamp).Milliseconds())
			c.Shutdown()
			sleepMs(100)

		}
		resultM = append(resultM, timeResults)
		m += 2
	}
	fmt.Println("###Crash Recovery Success! use time: ")
	//fmt.Println(timeResults)
	fmt.Println(resultM)
	fmt.Printf("lastLeader: %d, currentLeader: %d", lastLeader, currentLeader)

}
