package p2p

import (
	"errors"
	"fmt"
	"net"
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

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr    string           // 监听地址
	HandshakeFunc HandshakeFunc    // 握手处理函数
	Decoder       Decoder          // 解码器
	OnPeer        func(Peer) error // 两个节点成功建立连接并完成握手后的一些操作(回调函数)
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener // 监听接口
	rpcch    chan RPC     // 消息管道
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
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

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, false)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// read loop
	rpc := RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc)
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("TCP read error: %s\n", err)
			continue
		}

		rpc.From = peer.conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
