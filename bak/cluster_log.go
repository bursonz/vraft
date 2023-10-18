package bak

// SubmitToServer submits the command to serverId.
// 向Raft提交日志
func (c *Cluster) SubmitToServer(id int, cmd interface{}) bool {
	return true
}

// collectCommits reads channel commitChans[i] and adds all received entries
// to the corresponding commits[i]. It's blocking and should be run in a
// separate goroutine. It returns when commitChans[i] is closed.
// 收集已提交（达成共识）的日志
func (c *Cluster) collectCommits(i int) {
}

func (c *Cluster) Propose(cmd interface{}) bool {
	return true
}
