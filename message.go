package vraft

type MessageType int

const (
	MsgHup MessageType = iota
	MsgBeat
	MsgProposal

	// Log

	MsgAppend
	MsgAppendResp

	// Election

	MsgRequestVote
	MsgRequestVoteResp
	MsgHeartbeat
	MsgHeartbeatResp

	// Status

	MsgCheckQuorum

	// VRaft
	MsgBiasVote
	MsgBiasVoteResp
	MsgAppendForward // ViceLeader 转发给 Follower，Follower 返回 MsgAppendResp 给 Leader
)

type Message struct {
	Type MessageType // 消息类型

	From int // 如果是转发消息，则from应该为leader, 始终为leader
	To   int // 发送给谁

	Term     int // 当前node发送此消息时的所在任期
	LeaderId int // 当前node投票给谁,可以表示为leader

	// LogInfo 起始\冲突\已提交的日志的信息
	LogTerm     int // prev or conflict or leader's committed term. 起始\冲突\已提交的 LogTerm
	LogIndex    int // prev or conflict or leader's committed term. 起始\冲突\已提交的 LogIndex
	CommitIndex int // 已提交 index

	Entries []Entry // Entries

	Reject     bool // 拒绝或接受请求, 常用于回复消息
	RejectHint int  // 拒绝并改正, 修正leaderId

	// VRaft
	Size int // 单位是byte,表示整个message的大小
	//PreOrderedPeers []int // 更新预排序序列, index == index, value == peerId
	//ForwardPeers    []uint64 // 要转发给的节点, index == index, value == peerId
	//ForwardIndexes  []uint64 // 转发的起始 Index

	//ForwardMessages []Message // vice-leader应该转发的消息,需要附加上日志才能发送
	//Context  []byte
	//Response []Message
}

func (m *Message) clone() Message {
	newMsg := Message{
		Type:        m.Type,
		To:          m.To,
		From:        m.From,
		Term:        m.Term,
		LeaderId:    m.LeaderId,
		LogTerm:     m.LogTerm,
		LogIndex:    m.LogIndex,
		CommitIndex: m.CommitIndex,
		Entries:     m.Entries[:],
		Reject:      m.Reject,
		RejectHint:  m.RejectHint,
		Size:        m.Size,
		//PreOrderedPeers: m.PreOrderedPeers[:],
		//ForwardMessages: m.ForwardMessages[:],
	}
	return newMsg
}

//type HardState struct {
//	Term   int
//	Vote   int
//	Commit int
//}
//
//type SoftState struct {
//}
