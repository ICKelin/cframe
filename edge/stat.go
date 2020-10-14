package main

import (
	"sync/atomic"
	"time"

	"github.com/ICKelin/cframe/codec"
)

var msg = &codec.ReportMsg{}

func AddTrafficIn(traffic int64) {
	atomic.AddInt64(&msg.TrafficIn, traffic)
}

func AddTrafficOut(traffic int64) {
	atomic.AddInt64(&msg.TrafficOut, traffic)

}

func ResetStat() *codec.ReportMsg {
	p := msg
	p.Timestamp = time.Now().Unix()
	msg = &codec.ReportMsg{}

	return p
}
