package main

import (
	"os"
	"sync"
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

var msgMu sync.Mutex
var msg = &codec.ReportMsg{}

func AddTrafficIn(traffic int64) {
	msgMu.Lock()
	defer msgMu.Unlock()
	msg.TrafficIn += traffic
}

func AddTrafficOut(traffic int64) {
	msgMu.Lock()
	defer msgMu.Unlock()
	msg.TrafficOut += traffic
}

func AddErrorLog(err error) {
	msgMu.Lock()
	defer msgMu.Unlock()
	msg.Error = append(msg.Error, err.Error())
}

func ResetStat() *codec.ReportMsg {
	m := msg
	m.Timestamp = time.Now().Unix()
	cpu, _ := p.CPUPercent()
	mem, _ := p.MemoryPercent()
	m.CPU = int32(cpu)
	m.Mem = int32(mem)
	msg = &codec.ReportMsg{Error: make([]string, 0, 3)}

	return m
}
