package raft

type NetworkIFace interface {
	// Send 发送消息到指定peer
	Send(m Message)

	// Recv 从返回的<-chan中接收消息
	Recv() chan Message

	// Disconnect 断开指定Peer的连接（向所有节点发送MsgUnreachable或MsgDisconnect）
	Disconnect(id int)
	Connect(id int)
	ResetConnPool()
	Run()
	Stop()
}
