package p2p

import "net"

// Peer 一个代表远程节点的接口
type Peer interface {
	RemoteAddr() net.Addr
	Close() error
}

// Transport 处理网络中节点之间通信的任何东西。它可以是以下形式：(TCP, UDP, websockets, ...)
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC // 消费：获取消息通道(只读)
	Close() error
}
