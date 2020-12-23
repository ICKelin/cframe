package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/edge/vpc"
	log "github.com/ICKelin/cframe/pkg/logs"
)

type Server struct {
	registry *Registry

	// secret
	key string

	// server listen udp address
	laddr string

	// peers connection
	peerConns map[string]*peerConn

	// tun device wrap
	iface *Interface

	vpcInstance vpc.IVPC
}

type peerConn struct {
	addr string
	// conn *net.UDPConn
	// conn *kcp.UDPSession
	// conn net.Conn
	cidr string
}

func NewServer(laddr, key string, iface *Interface, vpcInstance vpc.IVPC) *Server {
	return &Server{
		laddr:       laddr,
		key:         key,
		peerConns:   make(map[string]*peerConn),
		iface:       iface,
		vpcInstance: vpcInstance,
	}
}

func (s *Server) SetRegistry(r *Registry) {
	s.registry = r
}

func (s *Server) SetVPCInstance(vpcInstance vpc.IVPC) {
	if s.vpcInstance == nil {
		s.vpcInstance = vpcInstance
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

	go s.readLocal(lconn)
	go s.ping(lconn)
	s.readRemote(lconn)
	return nil
}

func (s *Server) ping(sock *net.UDPConn) {
	tc := time.NewTicker(time.Second * 5)
	defer tc.Stop()
	ping := "ping"
	buf := make([]byte, 0, len(s.key)+len(ping))
	buf = append(buf, []byte(s.key)...)
	buf = append(buf, []byte(ping)...)
	for range tc.C {
		for _, p := range s.peerConns {
			log.Debug("ping %s", p.addr)
			raddr, err := net.ResolveUDPAddr("udp", p.addr)
			if err != nil {
				log.Error("resolve %s fail: %v", err)
				continue
			}

			sock.WriteToUDP(buf, raddr)
		}
	}
}

func (s *Server) readRemote(lconn *net.UDPConn) {
	buf := make([]byte, 1024*64)
	key := s.key
	klen := len(key)
	ping := "ping"
	plen := len(ping)
	for {
		nr, _, err := lconn.ReadFromUDP(buf)
		if err != nil {
			log.Error("read full fail: %v", err)
			continue
		}

		if nr < klen {
			log.Error("pkt to small")
			continue
		}

		// decode key
		rkey := buf[:klen]
		if string(rkey) != key {
			log.Error("access forbidden!!")
			continue
		}

		pkt := buf[klen:nr]

		if len(pkt) >= plen && string(pkt[:plen]) == ping {
			log.Debug("recv ping from remote")
			continue
		}

		p := Packet(pkt)
		if p.Invalid() {
			log.Error("invalid ipv4 packet")
			continue
		}

		src := p.Src()
		dst := p.Dst()
		log.Debug("tuple %s => %s", src, dst)

		AddTrafficIn(int64(nr))
		s.iface.Write(pkt)
	}
}

func (s *Server) readLocal(sock *net.UDPConn) {
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

		AddTrafficOut(int64(len(pkt)))
		src := p.Src()
		dst := p.Dst()
		log.Debug("tuple %s => %s", src, dst)

		peer, err := s.route(dst)
		if err != nil {
			log.Error("[E] not route to host: ", dst)
			continue
		}

		raddr, err := net.ResolveUDPAddr("udp", peer)
		if err != nil {
			log.Error("parse %s fail: %v", peer, err)
			continue
		}

		// encode key
		buf := make([]byte, 0, len(pkt)+len(s.key))
		buf = append(buf, []byte(s.key)...)
		buf = append(buf, pkt...)
		_, e := sock.WriteToUDP(buf, raddr)
		if e != nil {
			log.Error("%v", e)
		}
	}
}

func (s *Server) route(dst string) (string, error) {
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
			return p.addr, nil
		}
	}

	return "", fmt.Errorf("no route")
}

func (s *Server) AddPeer(peer *codec.Edge) error {
	log.Info("add peer: ", peer)

	// remove old item
	log.Info("removing old local route item")
	out, err := execCmd("route", []string{"del", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	if err != nil {
		log.Info("route del -net %s dev %s, %s %v",
			peer.Cidr, s.iface.tun.Name(), out, err)
	}

	// add local route item
	log.Info("adding new local route item")
	out, err = execCmd("route", []string{"add", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	if err != nil {
		log.Error("route add -net %s dev %s, %s %v\n",
			peer.Cidr, s.iface.tun.Name(), out, err)
		return err
	}
	log.Info("route add -net %s dev %s, %s %v\n",
		peer.Cidr, s.iface.tun.Name(), out, err)

	if s.vpcInstance != nil {
		// add vpc route entry
		// route to current instance
		log.Info("adding new vpc route item")
		err := s.vpcInstance.CreateRoute(peer.Cidr)
		if err != nil {
			log.Error("create vpc route fail: %v", err)
			// Do not return
		}
	}

	// finaly, add memory route
	s.peerConns[peer.Cidr] = &peerConn{
		addr: peer.ListenAddr,
		cidr: peer.Cidr,
	}

	return nil
}

func (s *Server) AddPeers(peers []*codec.Edge) {
	for _, p := range peers {
		s.AddPeer(p)
	}
}

func (s *Server) DelPeer(peer *codec.Edge) {
	log.Info("del peer: ", peer)
	delete(s.peerConns, peer.Cidr)

	out, err := execCmd("route", []string{"del", "-net",
		peer.Cidr, "dev", s.iface.tun.Name()})
	log.Info("route del -net %s dev %s, %s %v",
		peer.Cidr, s.iface.tun.Name(), out, err)
	// TODO: remove vpc route item
}

func (s *Server) AddRoute(msg *codec.AddRouteMsg) {
	s.AddPeer(&codec.Edge{
		Cidr:       msg.Cidr,
		ListenAddr: msg.Nexthop,
	})
}

func (s *Server) DelRoute(msg *codec.DelRouteMsg) {
	s.DelPeer(&codec.Edge{
		Cidr:       msg.Cidr,
		ListenAddr: msg.Nexthop,
	})
}
