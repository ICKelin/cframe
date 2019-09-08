package cframed

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/songgao/water"

	"github.com/ICKelin/cframe/proto"
	"github.com/xtaci/smux"
)

type Server struct {
	addr       string
	tun        *water.Interface
	routeTable *Route
}

func NewServer(addr string) *Server {
	cfg := water.Config{}
	cfg.DeviceType = water.TUN
	tun, err := water.New(cfg)
	if err != nil {
		panic(err)
	}

	return &Server{
		tun:        tun,
		addr:       addr,
		routeTable: NewRouter(),
	}
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	defer lis.Close()
	go s.readTun()

	for {
		conn, err := lis.Accept()
		if err != nil {
			break
		}

		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn net.Conn) {
	server, err := smux.Server(conn, nil)
	if err != nil {
		return
	}

	for {
		stream, err := server.AcceptStream()
		if err != nil {
			break
		}

		s.handleStream(stream)
	}
}

func (s *Server) readTun() {
	buf := make([]byte, 1<<16)
	for {
		nr, err := s.tun.Read(buf)
		if err != nil {
			break
		}

		p := Packet(buf[:nr])
		if p.Invalid() {
			fmt.Println("invalid packet")
			continue
		}

		dst := p.Dst()
		fmt.Println("dst", dst)
		conn := s.routeTable.NextHop(dst)
		if conn != nil {
			err := s.sendPacket(conn, buf[:nr])
			if err != nil {
				log.Println(err)
			}
		} else {
			log.Println("no route to ", dst)
		}
	}
}

func (s *Server) handleStream(stream net.Conn) {
	defer stream.Close()

	hs, err := s.readHandshake(stream)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(hs)
	s.routeTable.NewItem(net.ParseIP(hs.IP), hs.Mask, stream)
	defer s.routeTable.RemoveItem(net.ParseIP(hs.IP), hs.Mask)
	for {
		p, err := s.readPacket(stream)
		if err != nil {
			log.Println("read packet fail: ", err)
			break
		}

		_, err = s.tun.Write(p)
		if err != nil {
			log.Println(err)
		}
	}
}

// 1 byte version
// 2 byte cmd
// 2 byte body length
func (s *Server) readHandshake(stream net.Conn) (*proto.Handshake, error) {
	var header [4]byte
	_, err := io.ReadFull(stream, header[:])
	if err != nil {
		return nil, err
	}

	// TODO: verify

	bodylength := (int(header[2]) << 8) + int(header[3])
	buf := make([]byte, bodylength)

	_, err = io.ReadFull(stream, buf)
	if err != nil {
		return nil, err
	}

	hs := proto.Handshake{}
	err = json.Unmarshal(buf, &hs)
	if err != nil {
		return nil, err
	}

	// TODO: handshake response

	return &hs, nil
}

func (s *Server) readPacket(stream net.Conn) ([]byte, error) {
	var header [4]byte
	_, err := io.ReadFull(stream, header[:])
	if err != nil {
		return nil, err
	}

	bodylength := (int(header[2]) << 8) + int(header[3])
	buf := make([]byte, bodylength)

	_, err = io.ReadFull(stream, buf)
	return buf, err
}

func (s *Server) sendPacket(stream net.Conn, buf []byte) error {
	out := make([]byte, 4)
	out[0] = 1
	out[1] = 1

	binary.BigEndian.PutUint16(out[2:4], uint16(len(buf)))
	out = append(out, buf...)
	_, err := stream.Write(out)
	return err
}
