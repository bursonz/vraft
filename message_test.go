package raft

import (
	"fmt"
	"testing"
)

// 测试 Message 编解码后的len([]bytes)
func TestMessage_size(t *testing.T) {
	m1 := &Message{
		Type:        MsgAppendEntries,
		From:        1,
		To:          2,
		Term:        3,
		LeaderId:    4,
		LogTerm:     5,
		LogIndex:    6,
		CommitIndex: 7,
		Entries: []Entry{
			{1, "dsjahdkdj 13jd09j2890j9siajdoisajd 0j1209d8u1j209djw09dh810j2809dj0jas0ikojxc091 jk092j09dja0s9jd0sj 0ij10j d0289j8dj809wjasd "},
			{1, "123"},
			{1, "123"},
			{1, "123"},
		},
		Reject: false,
	}
	fmt.Println(m1)
	b := m1.ToBytes()
	fmt.Println(b)
	m2 := &Message{}
	m2.FromBytes(b)
	fmt.Println(m2)
	println(len(b))
}
