package main

import (
	"os"
	"sync/atomic"
	"time"

	"github.com/ICKelin/cframe/codec"
	"github.com/shirou/gopsutil/process"
)

var p *process.Process

func init() {
	p, _ = process.NewProcess(int32(os.Getpid()))
	if p == nil {
		panic("new process fail")
	}
}

var msg = &codec.ReportMsg{}

func AddTrafficIn(traffic int64) {
	atomic.AddInt64(&msg.TrafficIn, traffic)
}

func AddTrafficOut(traffic int64) {
	atomic.AddInt64(&msg.TrafficOut, traffic)
}

func ResetStat() *codec.ReportMsg {
	m := msg
	m.Timestamp = time.Now().Unix()
	cpu, _ := p.CPUPercent()
	mem, _ := p.MemoryPercent()
	m.CPU = int32(cpu)
	m.Mem = int32(mem)
	msg = &codec.ReportMsg{}

	return m
}
