package bak

import (
	"math/rand"
	"time"
)

// runElectionTimer 设置一个计时器来触发选举超时
func (r *Raft) runElectionTimer() {
	timeoutDuration := r.electionTimeout() // 生成超时时间
	r.mu.Lock()
	lastTerm := r.currentTerm // 获取当前term
	r.mu.Unlock()
	r.logger.Infof("开始选举计时, 超时时长为: (%v), term=%d", timeoutDuration, lastTerm)
	ticker := time.NewTicker(5 * time.Millisecond) // 5ms 检查一次
	defer ticker.Stop()

	for {
		<-ticker.C // 触发检查

		r.mu.Lock()

		if r.state != StateCandidate && r.state != StateFollower {
			r.logger.Infof("选举计时中, 节点状态变为:%s, 结束当前计时", r.state)
			r.mu.Unlock()
			return
		}

		if lastTerm != r.currentTerm {
			r.logger.Infof("选举计时中, 节点Term从 %d 变为 %d, 结束当前计时", lastTerm, r.currentTerm)
			r.mu.Unlock()
			return
		}

		// 选举计时器 超时
		if elapsed := time.Since(r.lastElectionTime); elapsed >= timeoutDuration {
			r.logger.Infof("选举超时")
			// 拦截选票流程到 BiasVote
			if len(r.preOrderedPeers) != 0 && r.vraft { // 存在预排序序列
				r.logger.Infof("存在ProOrder序列, 发起BiasVote")

				vote := r.preOrderedPeers[0]
				if vote != r.votedFor { // 如果第二次超时
					r.votedFor = vote
					m := Message{
						Type:    MsgBiasVote,
						To:      r.preOrderedPeers[0],
						From:    r.id,
						Term:    r.currentTerm,
						VoteFor: r.preOrderedPeers[0],
						Size:    MsgNormalSize,
					}
					r.node.Send(m) // 发送BiasVote
					// TODO:如何回应?
					// 重置定时，再等一个回合
					r.lastElectionTime = time.Now()
					// 缩短等待不一定是一件好事
					timeoutDuration = time.Duration(150) * time.Millisecond
					r.mu.Unlock()
					continue
				}
			}
			r.becomeCandidate()
			r.mu.Unlock()
			return
		}
		r.mu.Unlock()
	}
}

// electionTimeout 生成一个随机超时间隔
func (r *Raft) electionTimeout() time.Duration {
	if r.vraft && len(r.preOrderedPeers) != 0 {
		return time.Duration(BaseElapse) * time.Millisecond
	} else {
		return time.Duration(BaseElapse+rand.Intn(RandElapseRange)) * time.Millisecond
	}
}
