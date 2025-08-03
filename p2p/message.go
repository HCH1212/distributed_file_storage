package p2p

import "net"

// RPC 封装了在网络中两个节点之间通过每个传输层发送的任意数据
type RPC struct {
	From    net.Addr
	Payload []byte
}
