package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/ICKelin/cframe/edge/vpc"
	log "github.com/ICKelin/cframe/pkg/logs"
	"github.com/xtaci/smux"
)

type Server struct {
	// secret
	key string

	// server listen udp address
	laddr string

	// peers connection
	peerConnMu sync.RWMutex
	peerConns  map[string]*peerConn

	// tun device wrap
	iface *Interface

	// snd buffer
	sndq []chan *sendReq

	vpcInstance vpc.IVPC
}

type sendReq struct {
	buf  []byte
	conn net.Conn
}

type peerConn struct {
	addr string
	conn *smux.Stream
	cidr string
}

func NewServer(laddr, key string, iface *Interface) *Server {
	s := &Server{
		laddr:     laddr,
		key:       key,
		peerConns: make(map[string]*peerConn),
		iface:     iface,
		sndq:      make([]chan *sendReq, 1000),
	}

	go s.readLocal()
	for i := 0; i < 1000; i++ {
		s.sndq[i] = make(chan *sendReq, 10000)
		go s.writePeer(s.sndq[i])
	}
	return s
}

func (s *Server) SetVPCInstance(vpcInstance vpc.IVPC) {
	if s.vpcInstance == nil {
		s.vpcInstance = vpcInstance
	}
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.laddr)
	if err != nil {
		return err
	}
	defer lis.Close()

	for {
		rconn, err := lis.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(rconn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	sess, err := smux.Server(conn, nil)
	if err != nil {
		log.Error("run smux server fail: %v", err)
		return
	}
	defer sess.Close()

	for {
		stream, err := sess.AcceptStream()
		if err != nil {
			log.Error("accept stream fail: %v", err)
			break
		}
		log.Info("accept stream from: %v", stream.RemoteAddr())
		go s.handleStream(stream)
	}
}

func (s *Server) handleStream(stream *smux.Stream) {
	defer stream.Close()

	plen := make([]byte, 2)
	for {
		_, err := io.ReadFull(stream, plen)
		if err != nil {
			log.Error("read packet len fail: %v", err)
			break
		}

		size := binary.BigEndian.Uint16(plen)

		body := make([]byte, size)
		nr, err := io.ReadFull(stream, body)
		if err != nil {
			log.Error("read packet len fail: %v", err)
			break
		}

		pkt := body[:nr]
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

		AddTrafficOut(int64(len(pkt)))
		src := p.Src()
		dst := p.Dst()
		log.Debug("tuple %s => %s", src, dst)

		peer, err := s.route(dst)
		if err != nil {
			log.Error("[E] not route to host: ", dst)
			continue
		}

		plen := make([]byte, 2)
		binary.BigEndian.PutUint16(plen, uint16(len(pkt)))
		buf := make([]byte, 0, len(pkt)+2)
		buf = append(buf, plen...)
		buf = append(buf, pkt...)

		idx := time33(dst) % uint64(len(s.sndq))
		select {
		case s.sndq[idx] <- &sendReq{buf, peer}:
		default:
			log.Warn("sndq[%d] is full", idx)
			s.write(buf, peer)
		}
	}
}

func (s *Server) writePeer(sndq chan *sendReq) {
	for req := range sndq {
		peer := req.conn
		buf := req.buf
		s.write(buf, peer)
	}
}

func (s *Server) write(buf []byte, peer net.Conn) {
	peer.SetWriteDeadline(time.Now().Add(time.Second * 3))
	nw, err := peer.Write(buf)
	peer.SetWriteDeadline(time.Time{})
	if err != nil {
		log.Error("write to peer fail %v", err)
		return
	}

	if nw != len(buf) {
		log.Error("stream write not full")
		return
	}
}

func (s *Server) route(dst string) (*smux.Stream, error) {
	s.peerConnMu.RLock()
	defer s.peerConnMu.RUnlock()

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
			// ignore peer ip address
			ip, _, _ := net.SplitHostPort(p.addr)
			if ip == dst {
				continue
			}

			return p.conn, nil
		}
	}

	return nil, fmt.Errorf("no route")
}

func (s *Server) addRoute(peer *codec.Edge) error {
	log.Info("adding peer: %v", peer)

	var sess *smux.Session
	var stream *smux.Stream
	for {
		lconn, err := net.Dial("tcp", peer.ListenAddr)
		if err != nil {
			log.Error("dial peer fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}
		sess, err = smux.Client(lconn, nil)
		if err != nil {
			log.Error("smux client fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}
		stream, err = sess.OpenStream()
		if err != nil {
			log.Error("smux open stream fail: %v", err)
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}

	ipmask := strings.Split(peer.Cidr, "/")
	cidrtype := "-net"
	if len(ipmask) == 1 || ipmask[1] == "32" {
		cidrtype = "-host"
	}

	// add vpc route
	if s.vpcInstance != nil {
		// add vpc route entry
		// route to current instance
		err := s.vpcInstance.CreateRoute(peer.Cidr)
		if err != nil {
			log.Error("create vpc route fail: %v", err)
			AddErrorLog(err)
			// Do not return
		}
	}

	// add local static route
	execCmd("route", []string{"del", cidrtype,
		peer.Cidr, "dev", s.iface.tun.Name()})

	out, err := execCmd("route", []string{"add", cidrtype,
		peer.Cidr, "dev", s.iface.tun.Name()})
	if err != nil {
		log.Error("route add %s %s dev %s, %s %v\n",
			peer.Cidr, cidrtype, s.iface.tun.Name(), out, err)
		AddErrorLog(err)
		return err
	}

	// add memory route
	if cidrtype == "-host" {
		peer.Cidr = fmt.Sprintf("%s/32", ipmask[0])
	}

	s.peerConnMu.Lock()
	defer s.peerConnMu.Unlock()

	s.peerConns[peer.Cidr] = &peerConn{
		addr: peer.ListenAddr,
		cidr: peer.Cidr,
		conn: stream,
	}

	go s.deadlineCheck(peer, sess)
	log.Info("added peer %v OK", peer)
	log.Info("==========================\n")
	return nil
}

func (s *Server) deadlineCheck(peer *codec.Edge, sess *smux.Session) {
	tick := time.NewTicker(time.Second * 1)
	for range tick.C {
		if !sess.IsClosed() {
			continue
		}

		log.Info("receive dead channel for peer %s", peer.ListenAddr)
		for {
			s.peerConnMu.Lock()
			_, ok := s.peerConns[peer.Cidr]
			// peer edge has been removed
			if !ok {
				s.peerConnMu.Unlock()
				log.Warn("peer %s has not session", peer.Cidr)
				break
			}
			s.peerConnMu.Unlock()

			// reconnect
			lconn, err := net.Dial("tcp", peer.ListenAddr)
			if err != nil {
				log.Error("dial peer fail: %v", err)
				time.Sleep(time.Second * 3)
				continue
			}

			smuxSess, err := smux.Client(lconn, nil)
			if err != nil {
				log.Error("smux client fail: %v", err)
				time.Sleep(time.Second * 3)
				continue
			}

			stream, err := smuxSess.OpenStream()
			if err != nil {
				log.Error("smux open stream fail: %v", err)
				time.Sleep(time.Second * 3)
				continue
			}

			s.peerConnMu.Lock()
			s.peerConns[peer.Cidr] = &peerConn{
				addr: peer.ListenAddr,
				cidr: peer.Cidr,
				conn: stream,
			}
			s.peerConnMu.Unlock()
			sess = smuxSess
			break
		}
	}
}

func (s *Server) delRoute(peer *codec.Edge) {
	log.Info("del peer: %v", peer)
	ipmask := strings.Split(peer.Cidr, "/")
	cidrtype := "-net"
	if len(ipmask) == 1 || ipmask[1] == "32" {
		cidrtype = "-host"
	}
	out, err := execCmd("route", []string{"del", cidrtype,
		peer.Cidr, "dev", s.iface.tun.Name()})
	log.Info("route del %s %s dev %s, %s %v",
		cidrtype, peer.Cidr, s.iface.tun.Name(), out, err)

	if cidrtype == "-host" {
		peer.Cidr = fmt.Sprintf("%s/32", ipmask[0])
	}

	s.peerConnMu.Lock()
	defer s.peerConnMu.Unlock()
	peerConn, ok := s.peerConns[peer.Cidr]
	if ok {
		peerConn.conn.Close()
	}

	delete(s.peerConns, peer.Cidr)
	log.Info("del peer %s OK", peer)
	log.Info("==========================\n")
}

func (s *Server) AddPeers(peers []*codec.Edge) {
	for _, p := range peers {
		go s.addRoute(p)
	}
}

func (s *Server) AddPeer(peer *codec.Edge) {
	go s.addRoute(peer)
}

func (s *Server) DelPeer(peer *codec.Edge) {
	go s.delRoute(peer)
}

func (s *Server) AddRoute(msg *codec.AddRouteMsg) {
	go s.addRoute(&codec.Edge{
		Cidr:       msg.Cidr,
		ListenAddr: msg.Nexthop,
	})
}

func (s *Server) DelRoute(msg *codec.DelRouteMsg) {
	go s.delRoute(&codec.Edge{
		Cidr:       msg.Cidr,
		ListenAddr: msg.Nexthop,
	})
}

func time33(s string) uint64 {
	var hash uint64 = 5381

	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}

	return hash
}
