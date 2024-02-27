package raft

import (
	"fmt"
	"log"
	"math/rand"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	rand.Seed(RandSeed)
	rand.NewSource(RandSeed)
	fmt.Println("seed", RandSeed)
}

// Logger  config
const (
	LoggerLevel = LoggerLevelDebug
)

// Rand config
const (
	RandSeed = 3407
)

// ElectionElapse config
const (
	BaseElapse      = 150
	RandElapseRange = 150
)

// ID
const (
	None int = -1 // 空节点id
)
