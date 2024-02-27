package raft

type Peers map[int]Peer

type Peer struct {
	id   int
	term int
	vote int // Voted

	nextIndex  int
	matchIndex int

	bandwidth Bandwidth
}

func (p *Peer) Ping() (bool, float32) {
	//TODO
	return false, 0
}

func (p *Peer) ReportHardState() {

}
