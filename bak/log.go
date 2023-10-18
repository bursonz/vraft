package bak

type LogIndexType uint64

//type EntryType uint8

//const (
//	EntryNormal EntryType = iota + 1
//)

type Entry struct {
	Term  uint64
	Index uint64
	Data  []byte
	Size  uint64 // how many bytes per Entry
	//Type  EntryType
}

type CommitEntry struct {
	Term  uint64
	Index uint64
	Data  []byte
	Size  uint64 // how many bytes per Entry
}
