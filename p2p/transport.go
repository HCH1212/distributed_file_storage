package p2p

import "net"

// TODO
// Peer 一个代表远程节点的接口
type Peer interface {
	net.Conn           // 直接嵌入conn的接口
	Send([]byte) error // 针对节点的发送功能
	CloseStream()
}

// Transport 处理网络中节点之间通信的任何东西。它可以是以下形式：(TCP, UDP, websockets, ...)
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC // 消费：获取消息通道(只读)
	Close() error
}
