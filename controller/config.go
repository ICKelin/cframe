package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ApiAddr       string         `toml:"api_addr"`
	ListenAddr    string         `toml:"listen_addr"`
	BuildInEdages []*EdageConfig `toml:"edages"`
}

type EdageConfig struct {
	Name     string `toml:"name"`
	HostAddr string `toml:"host_addr"`
	Cidr     string `toml:"cidr"`
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
