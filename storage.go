package raft

type StorageIFace interface {
	Set(key string, value []byte) (bool, error)
	Get(key string) ([]byte, bool, error)
	HasData() (bool, error)
	Entries(lo, hi, maxSize int) (Entry, error)
	Term(i int) (int, error)
	LastIndex() (int, error)
	FirstIndex() (int, error)
	//Snapshot() (Snapshot, error)
}
