package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Controller string `toml:"controller"`
	Local      *Node  `toml:"local"`
}

type Node struct {
	ListenAddr string `toml:"listen_addr"` // 监听地址，内网
	Addr       string `toml:"addr"`        // 监听地址，外网
	CIDR       string `toml:"cidr"`
}

func ParseConfig(path string) (*Config, error) {
	cnt, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = toml.Unmarshal(cnt, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
