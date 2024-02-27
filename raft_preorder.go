package vraft

import (
	"fmt"
	"vraft/topsis"
)

// preOrder 预排序
// Thread Unsafe
func (r *Raft) preOrder() []int {
	r.l.Debugf("正在ProOrder")
	// 1. 准备数据
	// 1.1 生成metrics
	metrics := make([][]float64, 0)
	var peerList []int
	for _, p := range r.peers {
		// peerIdList
		peerList = append(peerList, p.id)
		// peersMetrics
		option := []float64{
			p.bandwidth.MBps(),
			p.AvgLatency,
			p.CPU,
			p.RAM,
			p.ROM,
			p.MTTF,
			p.MTTR,
		}
		metrics = append(metrics, option)
	}
	//peers := r.peers

	// 【Entropy-Topsis Sorting】
	//fmt.Println(len(metrics), metrics)
	//fmt.Println(len(r.peers), r.peers)
	orderedList := topsis.EntropyTopsis(metrics, r.n.criterias, peerList...) // 有序节点序列
	//fmt.Println(orderedList)

	// 【Log Bias Sorting】
	var preOrderedList []int
	peersIndex := r.matchIndex // its a map
	logIdx := len(r.log)
	for {
		r.l.Debugf("currentIndex:%d", logIdx)
		// 循环遍历，按照从大到小的顺序把log加入
		for _, peer := range orderedList {
			// 如果当前index 一致，则加入列表中
			if logIdx == peersIndex[peer] {
				r.l.Debugf("找到peer:%d", peer)
				preOrderedList = append(preOrderedList, peer)
			}
		}
		// 遍历完成
		if len(preOrderedList) == len(orderedList) {
			break
		}
		logIdx--
	}
	r.preOrderedPeers = preOrderedList
	fmt.Println(preOrderedList)
	return preOrderedList
}
