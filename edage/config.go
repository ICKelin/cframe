package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Name       string `toml:"name"`
	Controller string `toml:"controller"`
	ListenAddr string `toml:"listen_addr"`
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
