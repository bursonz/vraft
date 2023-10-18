package bak

type NetworkService interface {
}

type NodeService interface {
}

type RawNodeService interface {
}

type RaftService interface {
	//

	// Raft Election

	becomeFollower(term uint64)
	becomeCandidate()
	becomeLeader()
	becomeViceLeader()

	// Log
	broadcast(msgType MessageType)
}
