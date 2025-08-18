package main

import (
	"bytes"
	"distributed_file_storage/p2p"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000") // 在4000启动，想要连接3000
	s3 := makeServer(":5000", ":3000", ":4000")
	//s1 监听 :3000（作为服务端）
	//s2 监听 :4000，并尝试连接 s1 的 :3000（作为客户端）

	//当 s2（:4000）调用 Dial(":3000") 与 s1（:3000）建立 TCP 连接时：
	//s1 作为服务端，使用固定端口 3000 监听连接。
	//s2 作为客户端，发起连接时需要一个本地端口与服务端通信，但代码中并未指定该端口（由操作系统自动分配）。

	go s1.Start()
	time.Sleep(500 * time.Millisecond)

	go s2.Start()
	time.Sleep(500 * time.Millisecond)

	go s3.Start()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("my big data file here!"))
		s3.Store(key, data)

		if err := s3.store.Delete(key); err != nil {
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}
}
