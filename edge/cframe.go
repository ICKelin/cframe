package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/xtaci/kcp-go"

	"github.com/ICKelin/cframe/codec"
	log "github.com/ICKelin/cframe/pkg/logs"
)

type Server struct {
	registry *Registry

	// server监听udp地址
	laddr string

	// peers connection
	// as a kcp client
	peerConns map[string]*peerConn

	// 虚拟设备接口
	iface *Interface
}

type peerConn struct {
	// conn *net.UDPConn
	// conn *kcp.UDPSession
	conn net.Conn
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
	lconn, err := kcp.ListenWithOptions(s.laddr, nil, 10, 3)
	if err != nil {
		return err
	}
	defer lconn.Close()

	go s.readLocal()
	s.readRemote(lconn)
	return nil
}

func (s *Server) readRemote(lconn *kcp.Listener) {
	for {
		conn, err := lconn.AcceptKCP()
		if err != nil {
			log.Error("%v", err)
			break
		}

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn *kcp.UDPSession) {
	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(1, 20, 2, 1)
	conn.SetReadBuffer(4194304)
	conn.SetWriteBuffer(4194304)

	defer conn.Close()
	lenbuf := make([]byte, 2)
	for {
		_, err := io.ReadFull(conn, lenbuf)
		if err != nil {
			log.Error("read full fail: %v", err)
			return
		}

		length := binary.BigEndian.Uint16(lenbuf)
		log.Debug("pkt size: %d", length)

		buf := make([]byte, length)
		nr, err := io.ReadFull(conn, buf)
		if err != nil {
			log.Error("%v", err)
			return
		}

		p := Packet(buf[:nr])
		if p.Invalid() {
			log.Error("invalid ipv4 packet")
			continue
		}

		src := p.Src()
		dst := p.Dst()
		log.Debug("tuple %s => %s", src, dst)

		s.iface.Write(buf[:nr])
	}
}

func (s *Server) readLocal() {
	for {
		pkt, err := s.iface.Read()
		if err != nil {
			log.Error("read iface error: %v", err)
			continue
		}

		p := Packet(pkt)
		if p.Invalid() {
			log.Error("invalid ipv4 packet")
			continue
		}

		src := p.Src()
		dst := p.Dst()
		log.Info("local tuple %s => %s\n", src, dst)

		// report src ip as edge host ip
		s.registry.Report(src)

		peer, err := s.route(dst)
		if err != nil {
			log.Error("[E] not route to host: ", dst)
			continue
		}

		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(len(pkt)))
		buf = append(buf, pkt...)
		_, err = peer.Write(buf)
		if err != nil {
			log.Error("[E] write to peer: ", err)
		}
	}
}

func (s *Server) route(dst string) (net.Conn, error) {
	for _, p := range s.peerConns {
		_, ipnet, err := net.ParseCIDR(p.cidr)
		if err != nil {
			log.Error("parse cidr fail: %v", err)
			continue
		}

		sp := strings.Split(p.cidr, "/")
		if len(sp) != 2 {
			log.Error("parse cidr fail: %v", err)
			continue
		}

		dstCidr := fmt.Sprintf("%s/%s", dst, sp[1])
		_, dstNet, err := net.ParseCIDR(dstCidr)
		if err != nil {
			log.Error("parse cidr fail: %v", err)
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
	log.Info("add peer: ", peer)
	// if _, ok := s.peerConns[peer.Cidr]; ok {
	// 	log.Printf("host %s already added\n", peer.HostAddr)
	// 	return
	// }

	err := s.connectPeer(peer)
	if err != nil {
		log.Error("add peer %v fail: %v", peer, err)
	}

	out, err := execCmd("route", []string{"add", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	if err != nil {
		log.Error("route add -net %s dev %s, %s %v\n",
			peer.Cidr, s.iface.tun.Name(), out, err)
		// 移除peer
		s.disconnPeer(peer.Cidr)
		return
	}
	log.Info("route add -net %s dev %s, %s %v\n",
		peer.Cidr, s.iface.tun.Name(), out, err)
}

func (s *Server) AddPeers(peers []*codec.Host) {
	for _, p := range peers {
		s.AddPeer(p)
	}
}

func (s *Server) DelPeer(peer *codec.Host) {
	log.Info("del peer: ", peer)
	s.disconnPeer(peer.Cidr)

	out, err := execCmd("route", []string{"del", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	log.Info("route del -net %s dev %s, %s %v",
		peer.Cidr, s.iface.tun.Name(), out, err)
}

func (s *Server) connectPeer(node *codec.Host) error {
	conn, err := kcp.DialWithOptions(node.HostAddr, nil, 10, 3)
	if err != nil {
		log.Error("%v", err)
		return err
	}

	conn.SetStreamMode(true)
	conn.SetWriteDelay(false)
	conn.SetNoDelay(1, 20, 2, 1)
	conn.SetReadBuffer(4194304)
	conn.SetWriteBuffer(4194304)
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
	log.Info("delete peer %s", key)
}
