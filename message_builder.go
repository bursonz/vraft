package raft

func NewRequestVote() {

}

func NewRequestVoteResp() {

}

func NewAppendEntries() {

}

func NewAppendEntriesResp() {

}
func NewHeartBeat() {
	NewAppendEntries()
}
func NewHeartBeatReply() {
	NewAppendEntriesResp()
}
