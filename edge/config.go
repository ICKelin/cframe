package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Name         string       `toml:"name"`
	Controller   string       `toml:"controller"`
	ListenAddr   string       `toml:"listen_addr"`
	Type         string       `toml:"type"`
	AuthConfig   AuthConfig   `toml:"cframe_auth"`
	AliVPCConfig AliVPCConfig `toml:"alivpc"`
	Log          Log          `toml:"log"`
}

type AuthConfig struct {
	AccessKey string `toml:"cframe_key"`
	SecretKey string `toml:"cframe_secret"`
}

type AliVPCConfig struct {
	AccessKey    string `toml:"access_key"`
	AccessSecret string `toml:"access_secret"`
}

type Log struct {
	Days  int64  `toml:"days"`
	Level string `toml:"level"`
	Path  string `toml:"path"`
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

	if len(cfg.Type) == 0 {
		return nil, fmt.Errorf("type MUST configured")
	}

	// if len(cfg.AccessKey) == 0 {
	// 	cfg.AccessKey = os.Getenv("access_key")
	// }

	// if len(cfg.AccessSecret) == 0 {
	// 	cfg.AccessSecret = os.Getenv("access_secret")
	// }

	// if cfg.AccessKey == "" ||
	// 	cfg.AccessSecret == "" {
	// 	return nil, fmt.Errorf("access_key and secrect_key MUST configured")
	// }

	return &cfg, nil
}
