package p2p

// Peer 一个代表远程节点的接口
type Peer interface{}

// Transport 处理网络中节点之间通信的任何东西。它可以是以下形式：(TCP, UDP, websockets, ...)
type Transport interface {
	ListenAndAccept() error
}
