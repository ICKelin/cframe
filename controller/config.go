package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ApiAddr       string         `toml:"api_addr"`
	ListenAddr    string         `toml:"listen_addr"`
	Etcd          []string       `toml:"etcd"`
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

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "\t")
	return string(b)
}
