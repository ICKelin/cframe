package cframe

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/ICKelin/cframe/proto"
	"github.com/songgao/water"
	"github.com/xtaci/smux"
)

type Client struct {
	tun  *water.Interface
	srv  string
	lan  string
	mask int32
}

func NewClient(srv, lan string, mask int32) *Client {
	cfg := water.Config{}
	cfg.DeviceType = water.TUN
	tun, err := water.New(cfg)
	if err != nil {
		panic(err)
	}

	return &Client{
		tun:  tun,
		srv:  srv,
		lan:  lan,
		mask: mask,
	}
}

func (c *Client) Run() {
	for {
		err := c.run()
		if err != nil {
			log.Println(err)
		}

		log.Println("reconnecting")
		time.Sleep(time.Second * 5)
	}
}

func (c *Client) run() error {
	conn, err := net.Dial("tcp", c.srv)
	if err != nil {
		return err
	}
	defer conn.Close()

	sess, err := smux.Client(conn, nil)
	if err != nil {
		return err
	}

	stream, err := sess.OpenStream()
	if err != nil {
		return err
	}

	c.handshake(stream)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go c.tun2sock(stream, wg)
	go c.sock2tun(stream, wg)

	wg.Wait()

	return nil
}

func (c *Client) handshake(stream net.Conn) error {
	hs := &proto.Handshake{
		IP:   c.lan,
		Mask: c.mask,
	}

	b, err := json.Marshal(hs)
	if err != nil {
		return err
	}

	buf := make([]byte, 4)
	buf[0] = 1
	buf[1] = 1

	binary.BigEndian.PutUint16(buf[2:4], uint16(len(b)))
	buf = append(buf, b...)
	_, err = stream.Write(buf)
	return err
}

func (c *Client) tun2sock(stream net.Conn, wg *sync.WaitGroup) {
	defer stream.Close()
	defer wg.Done()

	buf := make([]byte, 1<<16)
	for {
		nr, err := c.tun.Read(buf)
		if err != nil {
			log.Println(err)
			break
		}

		log.Println("read:", buf[:nr])
		out := make([]byte, 4)
		out[0] = 1
		out[1] = 1

		binary.BigEndian.PutUint16(out[2:4], uint16(nr))
		out = append(out, buf[:nr]...)
		stream.SetWriteDeadline(time.Now().Add(time.Second * 3))
		_, err = stream.Write(out)
		stream.SetWriteDeadline(time.Time{})
	}
}

func (c *Client) sock2tun(stream net.Conn, wg *sync.WaitGroup) {
	defer stream.Close()
	defer wg.Done()

	for {
		var header [4]byte
		_, err := io.ReadFull(stream, header[:])
		if err != nil {
			log.Println(err)
			break
		}

		bodylength := (int(header[2]) << 8) + int(header[3])
		buf := make([]byte, bodylength)

		fmt.Println("read ", bodylength)
		_, err = io.ReadFull(stream, buf)
		if err != nil {
			log.Println(err)
			break
		}

		c.tun.Write(buf)
	}
}
