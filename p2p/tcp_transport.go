package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer 通过TCP建立的远程节点
type TCPPeer struct {
	conn net.Conn
	// outbound == true：表示这个连接是由本地节点主动发起（dial）的（出站连接）
	// outbound == false：表示这个连接是由本地节点被动接受（accept）的（入站连接）
	outbound bool // 出站
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{conn: conn, outbound: outbound}
}

type TCPTransport struct {
	listenAddress string        // 监听地址
	listener      net.Listener  // 监听接口
	shakeHands    HandshakeFunc // 握手处理函数
	decoder       Decoder       // 解码器

	mu    sync.RWMutex      // peer锁
	peers map[net.Addr]Peer // peer存储
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
		shakeHands:    NOPHandshakeFunc,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err)
		}

		fmt.Printf("new incoming connection %+v\n", conn)

		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, false)

	if err := t.shakeHands(peer); err != nil {

	}

	// read loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s\n", err)
			continue
		}
	}
}
