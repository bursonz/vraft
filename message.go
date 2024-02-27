package raft

import (
	"bytes"
	"encoding/gob"
	"log"
)

type MessageType int

const (
	MsgHup MessageType = iota
	MsgBeat
	MsgProposal

	// Log

	MsgAppendEntries
	MsgAppendEntriesResp

	// Election

	MsgRequestVote
	MsgRequestVoteResp
	MsgHeartbeat
	MsgHeartbeatResp

	// Status

	MsgCheckQuorum
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

	Reject bool // 拒绝或接受请求, 常用于回复消息
	//RejectHint int  // 消息提示
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
		//RejectHint:  m.RejectHint,
	}
	return newMsg
}

// ToBytes 将 Message 转化为二进制流
func (m *Message) ToBytes() []byte {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(m)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
		return nil
	}
	//fmt.Print(buf.Bytes())
	return buf.Bytes()
}

func (m *Message) FromBytes(b []byte) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(m)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	//fmt.Print(m)
}
