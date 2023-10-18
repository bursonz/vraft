package bak

func (c *Cluster) CheckLeader() (leader int, term int) {
	return 0, 0
}

// CheckSingleLeader checks that only a single node thinks it's the leader.
// Returns the leader's id and term
func (c *Cluster) CheckSingleLeader() (leader int, term int) {
	return 0, 0
}

func (c *Cluster) CheckNoLeader() {
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
func (c *Cluster) CheckCommitted(cmd int) (count int, index int) {
	return -1, -1
}

// CheckCommittedN verifies that cmd was committed by exactly n connected
// servers.
// 验证包含该日志cmd的节点数量是否等于n
func (c *Cluster) CheckCommittedN(cmd int, n int) {
}

// CheckNotCommitted verifies that no command equal to cmd has been committed
// by any of the active servers yet.
// 验证任何激活的节点中没有任何相等于cmd的提交
func (c *Cluster) CheckNotCommitted(cmd int) {
}

func (c *Cluster) CheckProOrderList() []int {
	return nil
}
