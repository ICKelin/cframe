package controller

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
}

func ParseConfig(path string) (*Config, error) {
	cnt, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(cnt)
}

func parseConfig(cnt []byte) (*Config, error) {
	var c Config
	err := toml.Unmarshal(cnt, &c)
	return &c, err
}

func (c *Config) String() string {
	b, _ := json.MarshalIndent(c, "", "  ")
	return string(b)
}
