package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ListenAddr string `toml:"listenAddr"`
	MongoUrl   string `toml:"mongoUrl"`
	DBName     string `toml:"dbname"`
	Log        Log    `toml:"log"`
}

type Log struct {
	Level string `toml:"level"`
	Path  string `toml:"path"`
	Days  int64  `toml:"days"`
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
