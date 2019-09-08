package cframed

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"testing"

	"github.com/ICKelin/cframe/proto"

	"github.com/xtaci/smux"
)

func TestServer(t *testing.T) {
	s := NewServer("127.0.0.1:10222")
	s.ListenAndServe()
}

func TestHandshake(t *testing.T) {
	s := NewServer("127.0.0.1:10222")

	go func() {
		conn, err := net.Dial("tcp", "127.0.0.1:10222")
		if err != nil {
			t.Error(err)
			return
		}

		sess, err := smux.Client(conn, nil)
		if err != nil {
			t.Error(err)
			return
		}

		stream, err := sess.OpenStream()
		if err != nil {
			t.Error(err)
			return
		}

		hs := &proto.Handshake{
			IP:   "192.168.0.1",
			Mask: 24,
		}

		b, err := json.Marshal(hs)
		if err != nil {
			t.Error(err)
			return
		}

		// 1 byte version
		// 2 byte cmd
		// 2 byte body length
		buf := make([]byte, 4)
		buf[0] = 1
		buf[1] = 1

		binary.BigEndian.PutUint16(buf[2:4], uint16(len(b)))
		buf = append(buf, b...)
		stream.Write(buf)
		stream.Close()
	}()
	s.ListenAndServe()
}
