package edgemanager

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type EtcdStorage struct {
	cli *clientv3.Client
}

func NewEtcdStorage(endpoints []string) *EtcdStorage {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	conn, err := clientv3.New(cfg)
	if err != nil {
		// just panic....
		panic(err)
	}

	return &EtcdStorage{
		cli: conn,
	}
}

func (s *EtcdStorage) Set(key string, val interface{}) {
	b, _ := json.Marshal(val)
	s.cli.Put(context.Background(), key, string(b))
}

func (s *EtcdStorage) SetWithExpiration(key string, val interface{}, exp time.Duration) error {
	b, _ := json.Marshal(val)
	// minimum lease TTL is 5-second
	resp, err := s.cli.Grant(context.TODO(), int64(exp.Seconds()))
	if err != nil {
		return err
	}

	_, err = s.cli.Put(context.Background(), key, string(b), clientv3.WithLease(resp.ID))
	return err
}

func (s *EtcdStorage) Get(key string, obj interface{}) error {
	resp, err := s.cli.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) <= 0 {
		return fmt.Errorf("empty value")
	}

	return json.Unmarshal(resp.Kvs[0].Value, obj)
}

func (s *EtcdStorage) Del(key string) {
	s.cli.Delete(context.Background(), key)
}

func (s *EtcdStorage) DelPrefix(prefix string) {
	s.cli.Delete(context.Background(), prefix, clientv3.WithPrefix())
}

func (s *EtcdStorage) List(root string) (map[string]string, error) {
	resp, err := s.cli.Get(context.Background(), root, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	for _, kv := range resp.Kvs {
		res[string(kv.Key)] = string(kv.Value)
	}
	return res, nil
}
