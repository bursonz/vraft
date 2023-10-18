package bak

type Tracker interface {
}

type Peer struct {
	id uint64

	bandwidth Bandwidth

	metric []float64
	option Option
}

type PeerProgress struct {
	next, match uint64
}

type Option struct {
	// Network
	bandwidth      float64 // transfer capability per second - higher is better
	averageLatency float64 // - lower is better

	// CPU, RAM, ROM - higher is better
	cpu float64
	ram float64
	rom float64

	// Reliability Assessment 可靠性
	FR   float64 // Failure Rate - lower is better
	MTBF float64 // Mean Time Between Failure - higher is better
	MTTR float64 // Mean Time Between Repair - lower is better
	MTTF float64 // Mean Time To Failure - higher is better
}
