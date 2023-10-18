package vraft

import (
	"github.com/fortytw2/leaktest"
	"testing"
	"time"
)

func genCluster(t *testing.T, n int) *Cluster {
	//bw := make(map[int]Bandwidth)
	//peers := []Peer{
	//	{0, 100, 20, 6, 4, 5, 6, 10000, 10000, 10000},
	//	{1, 100, 50, 8, 4, 5, 6, 10000, 10000, 10000},
	//	{2, 20, 25, 8, 2, 2, 8, 10000, 10000, 10000},
	//	{3, 30, 25, 8, 3, 3, 8, 10000, 10000, 10000},
	//	{4, 40, 25, 7, 4, 4, 6, 10000, 10000, 10000},
	//	{5, 100, 25, 6, 4, 5, 6, 10000, 10000, 10000},
	//	{6, 100, 50, 8, 4, 5, 6, 10000, 10000, 10000},
	//	{7, 20, 25, 8, 2, 2, 8, 10000, 10000, 10000},
	//	{8, 30, 25, 8, 3, 3, 8, 10000, 10000, 10000},
	//	{9, 40, 25, 7, 4, 4, 6, 10000, 10000, 10000},
	//	{10, 100, 25, 6, 4, 5, 6, 10000, 10000, 10000},
	//	{11, 100, 50, 8, 4, 5, 6, 10000, 10000, 10000},
	//	{12, 20, 25, 8, 2, 2, 8, 10000, 10000, 10000},
	//	{13, 20, 25, 8, 2, 2, 8, 10000, 10000, 10000},
	//	{14, 30, 25, 8, 3, 3, 8, 10000, 10000, 10000},
	//	{15, 40, 25, 7, 4, 4, 6, 10000, 10000, 10000},
	//}
	peersMap := make(map[int]Peer)
	for i := 0; i < n; i++ {
		peer := Peer{
			id: i,
			bandwidth: Bandwidth{
				bits: 100 * 1000 * 1000, //1 Mbps = 1*1000*1000 bits per second
				ms:   1000,
				unit: 1000,
			},
			AvgLatency: 20,
			CPU:        100,
			RAM:        100,
			ROM:        100,
			FR:         100,
			MTBF:       10000,
			MTTR:       10000,
			MTTF:       10000,
		}
		peersMap[peer.id] = peer
	}
	c := NewCluster(t, peersMap, 20*time.Millisecond)
	return c
}

// TestElectionBasic 选举测试
func TestElectionBasic(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	c.CheckSingleLeader()
}

func TestElectionLeaderDisconnect(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()
	lastLeader, lastTerm := c.CheckSingleLeader()

	c.DisconnectPeer(lastLeader)

	sleepMs(500)

	newLeader, newTerm := c.CheckSingleLeader()

	if newLeader == lastLeader {
		t.Errorf("want new leader to be different from orig leader")
	}
	if newTerm <= lastTerm {
		t.Errorf("want newTerm <= origTerm, got %d and %d", newTerm, lastTerm)
	}

}

func TestElectionLeaderAndAnotherDisconnect(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	lastLeader, _ := c.CheckSingleLeader()

	c.DisconnectPeer(lastLeader)
	otherId := (lastLeader + 1) % 3
	c.DisconnectPeer(otherId)

	// No quorum.
	sleepMs(450)
	c.CheckNoLeader()

	// Reconnect one other server; now we'll have quorum.
	c.ReconnectPeer(otherId)
	c.CheckSingleLeader()
}

func TestDisconnectAllThenRestore(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	sleepMs(100)
	//	Disconnect all servers from the start. There will be no leader.
	for i := 0; i < 3; i++ {
		c.DisconnectPeer(i)
	}
	sleepMs(450)
	c.CheckNoLeader()

	// Reconnect all servers. A leader will be found.
	for i := 0; i < 3; i++ {
		c.ReconnectPeer(i)
	}
	c.CheckSingleLeader()
}

func TestElectionLeaderDisconnectThenReconnect(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()
	lastLeader, _ := c.CheckSingleLeader()

	c.DisconnectPeer(lastLeader)

	sleepMs(350)
	newLeader, newTerm := c.CheckSingleLeader()

	c.ReconnectPeer(lastLeader)
	sleepMs(150)

	againLeader, againTerm := c.CheckSingleLeader()

	if newLeader != againLeader {
		t.Errorf("again leader id got %d; want %d", againLeader, newLeader)
	}
	if againTerm != newTerm {
		t.Errorf("again term got %d; want %d", againTerm, newTerm)
	}
}

func TestElectionLeaderDisconnectThenReconnect5(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 5)
	defer c.Shutdown()

	lastLeader, _ := c.CheckSingleLeader()

	c.DisconnectPeer(lastLeader)
	sleepMs(150)
	newLeader, newTerm := c.CheckSingleLeader()

	c.ReconnectPeer(lastLeader)
	sleepMs(150)

	againLeader, againTerm := c.CheckSingleLeader()

	if newLeader != againLeader {
		t.Errorf("again leader id got %d; want %d", againLeader, newLeader)
	}
	if againTerm != newTerm {
		t.Errorf("again term got %d; want %d", againTerm, newTerm)
	}
}

func TestElectionFollowerComesBack(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	lastLeader, lastTerm := c.CheckSingleLeader()

	otherPeer := (lastLeader + 1) % 3
	c.DisconnectPeer(otherPeer)
	time.Sleep(650 * time.Millisecond)
	c.ReconnectPeer(otherPeer)
	sleepMs(150)

	// We can't have an assertion on the new leader id here because it depends
	// on the relative election timeouts. We can assert that the term changed,
	// however, which implies that re-election has occurred.
	_, newTerm := c.CheckSingleLeader()
	if newTerm <= lastTerm {
		t.Errorf("newTerm=%d, origTerm=%d", newTerm, lastTerm)
	}
}

func TestElectionDisconnectLoop(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	for cycle := 0; cycle < 5; cycle++ {
		lastLeader, _ := c.CheckSingleLeader()

		c.DisconnectPeer(lastLeader)
		otherPeer := (lastLeader + 1) % 3
		c.DisconnectPeer(otherPeer)
		sleepMs(310)
		c.CheckNoLeader()

		// Reconnect both.
		c.ReconnectPeer(otherPeer)
		c.ReconnectPeer(lastLeader)

		// Give it time to settle
		sleepMs(150)
	}
}

// TestCommitOneCommand 测试提交一条日志
func TestCommitOneCommand(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()

	c.l.Testf("submitting 42 to %d", origLeaderId)
	isLeader := c.SubmitToServer(origLeaderId, 42) // 提交一条日志 42
	if !isLeader {
		t.Errorf("want id=%d leader, but it's not", origLeaderId)
	}

	sleepMs(250)
	c.CheckCommittedN(42, 3)
}

// TestSubmitNonLeaderFails 测试向非leader节点提交日志
func TestSubmitNonLeaderFails(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()
	sid := (origLeaderId + 1) % 3
	c.l.Testf("submitting 42 to %d", sid)
	isLeader := c.SubmitToServer(sid, 42) // 提交一条日志 42
	if isLeader {
		t.Errorf("want id=%d !leader, but it is", sid)
	}
	sleepMs(10)
}

// TestCommitMultipleCommands 提交多个日志
func TestCommitMultipleCommands(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()

	values := []int{42, 55, 81}
	for _, v := range values {
		c.l.Testf("submitting %d to %d", v, origLeaderId)
		isLeader := c.SubmitToServer(origLeaderId, v)
		if !isLeader {
			t.Errorf("want id=%d leader, but it's not", origLeaderId)
		}
		sleepMs(100)
	}

	sleepMs(150)
	nc, i1 := c.CheckCommitted(42)
	_, i2 := c.CheckCommitted(55)
	if nc != 3 {
		t.Errorf("want nc=3, got %d", nc)
	}
	if i1 >= i2 {
		t.Errorf("want i1<i2, got i1=%d i2=%d", i1, i2)
	}

	_, i3 := c.CheckCommitted(81)
	if i2 >= i3 {
		t.Errorf("want i2<i3, got i2=%d i3=%d", i2, i3)
	}
}

// TestCommitWithDisconnectionAndRecover 测试提交过程中断开和恢复follower
func TestCommitWithDisconnectionAndRecover(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	// Submit a couple of values to a fully connected cluster.
	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)

	// 检测是否达成共识
	sleepMs(250)
	c.CheckCommittedN(6, 3)

	// 断开一个follower
	dPeerId := (origLeaderId + 1) % 3
	c.DisconnectPeer(dPeerId)
	sleepMs(250)

	// Submit a new command; it will be committed but only to two servers.
	// 提交一个新的日志，该日志只会被另外两个节点提交
	c.SubmitToServer(origLeaderId, 7)
	sleepMs(250)
	c.CheckCommittedN(7, 2) //检测

	// Now reconnect dPeerId and wait a bit; it should find the new command too.
	c.ReconnectPeer(dPeerId) // 恢复连接
	sleepMs(200)
	c.CheckSingleLeader() // 检查leader

	sleepMs(150)
	c.CheckCommittedN(7, 3) // 检查日志是否恢复
}

// TestNoCommitWithNoQuorum 测试
func TestNoCommitWithNoQuorum(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)() // 检查测试是否超时

	c := genCluster(t, 3)
	defer c.Shutdown()

	// Submit a couple of values to a fully connected cluster.
	origLeaderId, origTerm := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)

	sleepMs(250)
	c.CheckCommittedN(6, 3) // 检查提交情况

	// Disconnect both followers.
	dPeer1 := (origLeaderId + 1) % 3
	dPeer2 := (origLeaderId + 2) % 3
	c.DisconnectPeer(dPeer1)
	c.DisconnectPeer(dPeer2)
	sleepMs(250)

	c.SubmitToServer(origLeaderId, 8) // 向leader提交
	sleepMs(250)
	c.CheckNotCommitted(8) // 检查没被提交情况

	// Reconnect both other servers, we'll have quorum now.
	c.ReconnectPeer(dPeer1)
	c.ReconnectPeer(dPeer2)
	sleepMs(600)

	// 8 is still not committed because the term has changed.
	c.CheckNotCommitted(8)

	// A new leader will be elected. It could be a different leader, even though
	// the original's log is longer, because the two reconnected peers can elect
	// each other.
	newLeaderId, againTerm := c.CheckSingleLeader() // 检查脑裂
	if origTerm == againTerm {                      // 新term不应该与之前一一致
		t.Errorf("got origTerm==againTerm==%d; want them different", origTerm)
	}

	// But new values will be committed for sure...
	c.SubmitToServer(newLeaderId, 9)
	c.SubmitToServer(newLeaderId, 10)
	c.SubmitToServer(newLeaderId, 11)
	sleepMs(350)

	for _, v := range []int{9, 10, 11} {
		c.CheckCommittedN(v, 3) // 核对提交数量
	}
}

func TestDisconnectLeaderBriefly(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	// Submit a couple of values to a fully connected cluster.
	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)
	sleepMs(250)
	c.CheckCommittedN(6, 3)

	// Disconnect leader for a short time (less than election timeout in peers).
	c.DisconnectPeer(origLeaderId)
	sleepMs(90)
	c.ReconnectPeer(origLeaderId)
	sleepMs(200)

	c.SubmitToServer(origLeaderId, 7)
	sleepMs(250)
	c.CheckCommittedN(7, 3)
}

// TestCommitsWithLeaderDisconnects 测试leader离线时提交
func TestCommitsWithLeaderDisconnects(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 5)
	defer c.Shutdown()

	// Submit a couple of values to a fully connected cluster.
	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)

	sleepMs(150)
	c.CheckCommittedN(6, 5)

	// Leader disconnected...
	c.DisconnectPeer(origLeaderId)
	sleepMs(10)

	// Submit 7 to original leader, even though it's disconnected.
	c.SubmitToServer(origLeaderId, 7) // 离线提交

	sleepMs(150)
	c.CheckNotCommitted(7) // 检测不该存在的提交

	newLeaderId, _ := c.CheckSingleLeader() // 获取新leader信息

	// Submit 8 to new leader.
	c.SubmitToServer(newLeaderId, 8)
	sleepMs(150)
	c.CheckCommittedN(8, 4) // 核对数量

	// Reconnect old leader and let it settle. The old leader shouldn't be the one
	// winning.
	c.ReconnectPeer(origLeaderId) // 重连
	sleepMs(600)

	finalLeaderId, _ := c.CheckSingleLeader()
	if finalLeaderId == origLeaderId {
		t.Errorf("got finalLeaderId==origLeaderId==%d, want them different", finalLeaderId)
	}

	// Submit 9 and check it's fully committed.
	c.SubmitToServer(newLeaderId, 9)
	sleepMs(150)
	c.CheckCommittedN(9, 5)
	c.CheckCommittedN(8, 5)

	// But 7 is not committed...
	c.CheckNotCommitted(7)
}

// TestCrashFollower 测试 Follower 崩溃
func TestCrashFollower(t *testing.T) {
	// Basic test to verify that crashing a peer doesn't blow up.
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)

	sleepMs(350)
	c.CheckCommittedN(5, 3)

	c.CrashPeer((origLeaderId + 1) % 3)
	sleepMs(350)
	c.CheckCommittedN(5, 2)
}

// TestCrashThenRestartFollower
func TestCrashThenRestartFollower(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)
	c.SubmitToServer(origLeaderId, 7)

	vals := []int{5, 6, 7}

	sleepMs(350)
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}

	c.CrashPeer((origLeaderId + 1) % 3)
	sleepMs(350)
	for _, v := range vals {
		c.CheckCommittedN(v, 2)
	}

	// Restart the crashed follower and give it some time to come up-to-date.
	c.RestartPeer((origLeaderId + 1) % 3)
	sleepMs(650)
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}
}

// TestCrashThenRestartLeader
func TestCrashThenRestartLeader(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)
	c.SubmitToServer(origLeaderId, 7)

	vals := []int{5, 6, 7}

	sleepMs(350)
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}

	c.CrashPeer(origLeaderId)
	sleepMs(350)
	for _, v := range vals {
		c.CheckCommittedN(v, 2)
	}

	c.RestartPeer(origLeaderId)
	sleepMs(550)
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}
}

func TestCrashThenRestartAll(t *testing.T) {
	defer leaktest.CheckTimeout(t, 100*time.Millisecond)()

	c := genCluster(t, 3)
	defer c.Shutdown()

	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)
	c.SubmitToServer(origLeaderId, 7)

	vals := []int{5, 6, 7}

	sleepMs(350)
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}

	for i := 0; i < 3; i++ {
		c.CrashPeer((origLeaderId + i) % 3)
	}

	sleepMs(350)

	for i := 0; i < 3; i++ {
		c.RestartPeer((origLeaderId + i) % 3)
	}

	sleepMs(150)
	newLeaderId, _ := c.CheckSingleLeader()

	c.SubmitToServer(newLeaderId, 8)
	sleepMs(250)

	vals = []int{5, 6, 7, 8}
	for _, v := range vals {
		c.CheckCommittedN(v, 3)
	}
}

func TestReplaceMultipleLogEntries(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	// Submit a couple of values to a fully connected cluster.
	origLeaderId, _ := c.CheckSingleLeader()
	c.SubmitToServer(origLeaderId, 5)
	c.SubmitToServer(origLeaderId, 6)

	sleepMs(250)
	c.CheckCommittedN(6, 3)

	// Leader disconnected...
	c.DisconnectPeer(origLeaderId)
	sleepMs(10)

	// Submit a few entries to the original leader; it's disconnected, so they
	// won't be replicated.
	c.SubmitToServer(origLeaderId, 21)
	sleepMs(5)
	c.SubmitToServer(origLeaderId, 22)
	sleepMs(5)
	c.SubmitToServer(origLeaderId, 23)
	sleepMs(5)
	c.SubmitToServer(origLeaderId, 24)
	sleepMs(5)

	newLeaderId, _ := c.CheckSingleLeader()

	// Submit entries to new leader -- these will be replicated.
	c.SubmitToServer(newLeaderId, 8)
	sleepMs(5)
	c.SubmitToServer(newLeaderId, 9)
	sleepMs(5)
	c.SubmitToServer(newLeaderId, 10)
	sleepMs(250)
	c.CheckNotCommitted(21)
	c.CheckCommittedN(10, 2)

	// Crash/restart new leader to reset its nextIndex, to ensure that the new
	// leader of the cluster (could be the third server after elections) tries
	// to replace the original's servers unreplicated entries from the very end.
	c.CrashPeer(newLeaderId)
	sleepMs(60)
	c.RestartPeer(newLeaderId)

	sleepMs(100)
	finalLeaderId, _ := c.CheckSingleLeader()
	c.ReconnectPeer(origLeaderId)
	sleepMs(400)

	// Submit another entry; this is because leaders won't commit entries from
	// previous terms (paper 5.4.2) so the 8,9,10 may not be committed everywhere
	// after the restart before a new command comes it.
	c.SubmitToServer(finalLeaderId, 11)
	sleepMs(250)

	// At this point, 11 and 10 should be replicated everywhere; 21 won't be.
	c.CheckNotCommitted(21)
	c.CheckCommittedN(11, 3)
	c.CheckCommittedN(10, 3)
}

func TestCrashAfterSubmit(t *testing.T) {
	c := genCluster(t, 3)
	defer c.Shutdown()

	// Wait for a leader to emerge, and submit a command - then immediately
	// crash; the leader should have no time to send an updated LeaderCommit
	// to followers. It doesn't have time to get back AE responses either, so
	// the leader itself won't send it on the commit channel.
	origLeaderId, _ := c.CheckSingleLeader()

	c.SubmitToServer(origLeaderId, 5)
	sleepMs(1)
	c.CrashPeer(origLeaderId)

	// Make sure 5 is not committed when a new leader is elected. Leaders won't
	// commit commands from previous terms.
	sleepMs(10)
	c.CheckSingleLeader()
	sleepMs(300)
	c.CheckNotCommitted(5)

	// The old leader restarts. After a while, 5 is still not committed.
	c.RestartPeer(origLeaderId)
	sleepMs(150)
	newLeaderId, _ := c.CheckSingleLeader()
	c.CheckNotCommitted(5)

	// When we submit a new command, it will be submitted, and so will 5, because
	// it appears in everyone's logs.
	c.SubmitToServer(newLeaderId, 6)
	sleepMs(100)
	c.CheckCommittedN(5, 3)
	c.CheckCommittedN(6, 3)
}

func TestRaftSubmit(t *testing.T) {
	c := genCluster(t, 5)
	defer c.Shutdown()

	sleepMs(300)
	leader, _ := c.CheckSingleLeader()
	sleepMs(300)
	c.SubmitToServer(leader, 0)

	for r := 0; r < 5; r++ {
		for i := 1 + r*100; i < 100*(r+1)+1; i++ {
			//leader, _ = c.CheckSingleLeader()
			sleepMs(10) // rate = 1s/10ms = 1000/10 ms = 100 op/ms
			//c.SubmitToServer(leader, i)
			c.Propose(i)
		}
		sleepMs(300)
	}
	sleepMs(300)
	c.CheckCommitted(500)
	//sleepMs(300)
	//c.CheckCommitted(10001)
}
