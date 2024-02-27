package raft

import (
	"math/rand"
	"time"
)

// 150-300ms
func (r *Raft) getElectionTimeout() time.Duration {
	if r.electionTimeout.Seconds() == 0 {
		r.electionTimeout = time.Duration(150+rand.Intn(150)) * time.Millisecond
	}
	//return r.electionTimeout
	return time.Duration(150+rand.Intn(150)) * time.Millisecond
}

func (r *Raft) runElectionTimer() {
	timeoutDuration := r.getElectionTimeout() // 生成超时时间
	r.mu.Lock()
	lastTerm := r.term // 获取当前term
	r.mu.Unlock()
	r.l.Infof("开始选举计时, 超时时长为: (%v), term=%d", timeoutDuration, lastTerm)

	ticker := time.NewTicker(5 * time.Millisecond) // 5ms 检查一次
	defer ticker.Stop()

	for {
		<-ticker.C

		r.mu.Lock()

		// 状态变化 - 退出计时
		if r.state != StateCandidate && r.state != StateFollower {
			r.l.Infof("选举计时中, 节点状态变为:%d, 结束当前计时", r.state)
			r.mu.Unlock()
			return
		}

		// 任期变化 - 退出计时
		if lastTerm != r.term {
			r.l.Infof("选举计时中, 节点Term从 %d 变为 %d, 结束当前计时", lastTerm, r.term)
			r.mu.Unlock()
			return
		}

		// 超时选举 - 变为candidate - 发起选举
		if elapsed := time.Since(r.lastElectionTime); elapsed >= timeoutDuration {
			r.l.Infof("选举超时")
			r.becomeCandidate()
			r.mu.Unlock()
			return
		}

		r.mu.Unlock()
	}
}
