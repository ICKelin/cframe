package main

import (
	"log"
	"net"
)

type Server struct {
	laddr    string
	peers    []*Node
	peerConn []*net.UDPConn
}

func NewServer(laddr string, peers []*Node) *Server {
	return &Server{
		laddr:    laddr,
		peers:    peers,
		peerConn: make([]*net.UDPConn, 0),
	}
}

func (s *Server) ListenAndServe() error {
	laddr, err := net.ResolveUDPAddr("udp", s.laddr)
	if err != nil {
		return err
	}

	lconn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}
	defer lconn.Close()

	iface, err := NewInterface()
	if err != nil {
		return err
	}
	defer iface.Close()
	iface.Up()

	for _, node := range s.peers {
		raddr, err := net.ResolveUDPAddr("udp", node.Addr)
		if err != nil {
			log.Println("[E] ", err)
			continue
		}
		conn, err := net.DialUDP("udp", nil, raddr)
		if err != nil {
			log.Println("[E] ", err)
			continue
		}

		s.peerConn = append(s.peerConn, conn)
	}

	go s.readLocal(lconn, iface)
	s.readRemote(lconn, iface)
	return nil
}

func (s *Server) readRemote(lconn *net.UDPConn, iface *Interface) {
	buf := make([]byte, 1024*64)
	for {
		nr, _, err := lconn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			return
		}

		p := Packet(buf[:nr])
		if p.Invalid() {
			log.Println("[E] invalid ipv4 packet")
			continue
		}

		src := p.Src()
		dst := p.Dst()
		log.Printf("[D] %s => %s\n", src, dst)

		iface.Write(buf[:nr])
	}
}

func (s *Server) readLocal(lconn *net.UDPConn, iface *Interface) {
	for {
		buf, err := iface.Read()
		if err != nil {
			log.Println("[E] read iface error: ", err)
			continue
		}

		p := Packet(buf)
		if p.Invalid() {
			log.Println("[E] invalid ipv4 packet")
			continue
		}

		src := p.Src()
		dst := p.Dst()
		log.Printf("[D] %s => %s\n", src, dst)

		for _, conn := range s.peerConn {
			_, err := conn.Write(buf)
			if err != nil {
				log.Println("[E] write to peer: ", err)
				continue
			}
		}
	}
}
