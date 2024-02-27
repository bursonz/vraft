package raft

import "time"

func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sleepMs(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}
