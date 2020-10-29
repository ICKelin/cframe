package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ListenAddr     string   `toml:"listen_addr"`
	Etcd           []string `toml:"etcd"`
	MongoUrl       string   `toml:"mongourl"`
	DBName         string   `toml:"dbname"`
	UserCenterAddr string   `toml:"usercenter_addr"`
	RpcAddr        string   `toml:"rpc_addr"`
	Log            Log      `toml:"log"`
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
