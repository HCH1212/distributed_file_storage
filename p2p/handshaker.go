package p2p

// HandshakeFunc ....
type HandshakeFunc func(Peer) error

// NOPHandshakeFunc 不进行握手
func NOPHandshakeFunc(Peer) error {
	return nil
}
