package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/ICKelin/cframe/codec"
)

type Server struct {
	registry *Registry

	// server监听udp地址
	laddr string

	// 与其他宿主机的udp connect
	peerConns map[string]*peerConn

	// 虚拟设备接口
	iface *Interface
}

type peerConn struct {
	conn *net.UDPConn
	cidr string
}

func NewServer(laddr string, iface *Interface) *Server {
	return &Server{
		laddr:     laddr,
		peerConns: make(map[string]*peerConn),
		iface:     iface,
	}
}
func (s *Server) SetRegistry(r *Registry) {
	s.registry = r
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

	go s.readLocal(lconn)
	s.readRemote(lconn)
	return nil
}

func (s *Server) readRemote(lconn *net.UDPConn) {
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

		s.iface.Write(buf[:nr])
	}
}

func (s *Server) readLocal(lconn *net.UDPConn) {
	for {
		buf, err := s.iface.Read()
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

		// report src ip as edage host ip
		s.registry.Report(src)

		peer, err := s.route(dst)
		if err != nil {
			log.Println("[E] not route to host: ", dst)
			continue
		}

		_, err = peer.Write(buf)
		if err != nil {
			log.Println("[E] write to peer: ", err)
		}
	}
}

func (s *Server) route(dst string) (*net.UDPConn, error) {
	for _, p := range s.peerConns {
		_, ipnet, err := net.ParseCIDR(p.cidr)
		if err != nil {
			log.Println("parse cidr fail: ", err)
			continue
		}

		sp := strings.Split(p.cidr, "/")
		if len(sp) != 2 {
			log.Println("parse cidr fail: ", err)
			continue
		}

		dstCidr := fmt.Sprintf("%s/%s", dst, sp[1])
		_, dstNet, err := net.ParseCIDR(dstCidr)
		if err != nil {
			log.Println("parse cidr fail: ", err)
			continue
		}

		if ipnet.String() == dstNet.String() {
			return p.conn, nil
		}
	}

	return nil, fmt.Errorf("no route")
}

func (s *Server) AddPeer(peer *codec.Host) {
	s.DelPeer(peer)
	log.Println("[I] add peer: ", peer)
	// if _, ok := s.peerConns[peer.Cidr]; ok {
	// 	log.Printf("host %s already added\n", peer.HostAddr)
	// 	return
	// }

	err := s.connectPeer(peer)
	if err != nil {
		log.Printf("[E] add peer %v fail: %v\n", peer, err)
	}

	out, err := execCmd("route", []string{"add", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	if err != nil {
		log.Printf("[I] route add -net %s dev %s, %s %v\n",
			peer.Cidr, s.iface.tun.Name(), out, err)
		// 移除peer
		s.disconnPeer(peer.Cidr)
		return
	}
	log.Printf("[I] route add -net %s dev %s, %s %v\n",
		peer.Cidr, s.iface.tun.Name(), out, err)
}

func (s *Server) AddPeers(peers []*codec.Host) {
	for _, p := range peers {
		s.AddPeer(p)
	}
}

func (s *Server) DelPeer(peer *codec.Host) {
	log.Println("[I] del peer: ", peer)
	s.disconnPeer(peer.Cidr)

	out, err := execCmd("route", []string{"del", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	log.Printf("[I] route del -net %s dev %s, %s %v\n",
		peer.Cidr, s.iface.tun.Name(), out, err)
}

func (s *Server) connectPeer(node *codec.Host) error {
	raddr, err := net.ResolveUDPAddr("udp", node.HostAddr)
	if err != nil {
		log.Println("[E] ", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Println("[E] ", err)
		return err
	}

	peer := &peerConn{
		conn: conn,
		cidr: node.Cidr,
	}

	s.peerConns[peer.cidr] = peer
	return nil
}

func (s *Server) disconnPeer(key string) {
	p := s.peerConns[key]
	if p != nil {
		p.conn.Close()
	}

	delete(s.peerConns, key)
	log.Printf("[I] delete peer %s\n", key)
}
