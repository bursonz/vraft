package vraft

type Entry struct {
	Term  int
	Index int
	Data  interface{}
	Size  int // how many bytes per Entry
	//Type  EntryType
}

//type LogEntry struct {
//	Data interface{}
//	Term int
//	Size int
//}

type CommitEntry struct {
	Term  int
	Index int
	Data  interface{}
	Size  int // how many bytes per Entry
}
