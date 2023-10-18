package bak

import "time"

func (r *Raft) becomeCandidate() {
	r.state = StateCandidate
	r.currentTerm += 1
	currentTerm := r.currentTerm
	r.lastElectionTime = time.Now()
	r.votedFor = r.id // vote for itself
	r.logger.Infof("becomes Candidate (currentTerm=%d); log=%v", currentTerm, r.log)

	r.votes[r.id] = currentTerm // vote itself
	r.logger.Infof("### CurrentBiasVote %d", len(r.votes))
	if r.checkQuorumVotes() {
		r.logger.Infof("wins election with %d votes by BiasVote", len(r.votes))
		r.becomeLeader()
		return
	}

	for _, peer := range r.peers {
		go func(id uint64) {
			r.mu.Lock()
			lastLogIndex, lastLogTerm := r.lastLogIndexAndTerm()
			r.mu.Unlock()

			request := Message{
				Type:     MsgRequestVote,
				From:     r.id,
				To:       id,
				Term:     currentTerm,
				VoteFor:  r.id,
				LogTerm:  lastLogTerm,
				LogIndex: lastLogIndex,

				Size: MsgNormalSize,
			}
			r.node.Send(request)

		}(peer.id)
	}
	// Run another election time, in case this election is not successful
	go r.runElectionTimer()
}
func (r *Raft) becomeFollower(term uint64) { //TODO

}

func (r *Raft) becomeLeader() { //TODO

}
