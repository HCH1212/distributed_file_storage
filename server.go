package main

import (
	"bytes"
	"distributed_file_storage/p2p"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string // 引导节点
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		PathTransformFunc: opts.PathTransformFunc,
		Root:              opts.StorageRoot,
	}
	return &FileServer{
		FileServerOpts: opts,
		store:          NewStore(storeOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

//// TODO
//// broadcast 方法利用了 io.MultiWriter 的强大功能，实现了高效的“一写多发”。它避免了写一个循环，然后逐个发送数据给每个对等节点的繁琐过程，使代码更加简洁和优雅
//func (s *FileServer) broadcast(msg *Message) error {
//	var peers []io.Writer
//	for _, peer := range s.peers {
//		peers = append(peers, peer)
//	}
//
//	mw := io.MultiWriter(peers...)
//	return gob.NewEncoder(mw).Encode(msg)
//}

func (s *FileServer) stream(msg *Message) error {
	var peers []io.Writer
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

func (s *FileServer) Get(key string) (io.Reader, error) {
	return nil, nil
}

func (s *FileServer) Store(key string, r io.Reader) error {
	// 1. 存储文件到磁盘
	// 2. 将文件广播到网络中所有已知节点
	var (
		fileBuffer = new(bytes.Buffer)
		tee        = io.TeeReader(r, fileBuffer)
	)

	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, fileBuffer)
		if err != nil {
			return err
		}

		fmt.Println("received and written bytes to disk: ", n)
	}

	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil {
				log.Println(err)
				return
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) not found", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("written %d bytes to disk\n", n)

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if addr == "" {
			continue
		}
		go func(addr string) {
			fmt.Println("attempting to connect with remote: ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

// Message 中是any，gob 在编码和解码接口类型时，必须提前知道接口可能包含的具体类型
func init() {
	gob.Register(MessageStoreFile{})
}
