package bak

import (
	"log"
	"math/rand"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	rand.Seed(RandSeed)
}

// Message config
const (
	MsgNormalSize = 10
)

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
	None = 0 // 空节点id
)
