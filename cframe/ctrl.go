package cframe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	defaultScheme = "http://"
)

type Node struct {
	Gateway string `json:"gateway"`
	CIDR    string `json:"cidr"`
	Name    string `json:"name"`
}

type CtrlConfig struct {
	Scheme string `toml:"scheme"`
	Addr   string `toml:"addr"`
}

type CtrlClient struct {
	scheme string
	addr   string
}

func NewCtrlClient(cfg CtrlConfig) *CtrlClient {
	if len(cfg.Scheme) == 0 {
		cfg.Scheme = defaultScheme
	}

	return &CtrlClient{
		addr:   cfg.Addr,
		scheme: cfg.Scheme,
	}
}

func (c *CtrlClient) GetNodes() ([]*Node, error) {
	url := fmt.Sprintf("%s%s/api/v1/nodes", c.scheme, c.addr)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodes []*Node
	err = json.Unmarshal(cnt, &nodes)
	return nodes, err
}
