package p2p

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// RPC 封装了在网络中两个节点之间通过每个传输层发送的任意数据
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
