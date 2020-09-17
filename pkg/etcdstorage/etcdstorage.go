package etcdstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type Etcd struct {
	cli *clientv3.Client
}

func NewEtcd(endpoints []string) *Etcd {
	cfg := clientv3.Config{
		Endpoints: endpoints,
	}

	conn, err := clientv3.New(cfg)
	if err != nil {
		// just panic....
		panic(err)
	}

	return &Etcd{
		cli: conn,
	}
}

func (s *Etcd) Set(key string, val interface{}) error {
	b, _ := json.Marshal(val)
	_, err := s.cli.Put(context.Background(), key, string(b))
	return err
}

func (s *Etcd) SetWithExpiration(key string, val interface{},
	exp time.Duration) error {
	b, _ := json.Marshal(val)
	resp, err := s.cli.Grant(context.TODO(), int64(exp.Seconds()))
	if err != nil {
		return err
	}

	ctx := context.Background()
	_, err = s.cli.Put(ctx, key, string(b), clientv3.WithLease(resp.ID))
	return err
}

func (s *Etcd) Get(key string, obj interface{}) error {
	resp, err := s.cli.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) <= 0 {
		return fmt.Errorf("empty value")
	}

	return json.Unmarshal(resp.Kvs[0].Value, obj)
}

func (s *Etcd) Del(key string) {
	s.cli.Delete(context.Background(), key)
}

func (s *Etcd) DelPrefix(prefix string) {
	s.cli.Delete(context.Background(), prefix, clientv3.WithPrefix())
}

func (s *Etcd) List(root string) (map[string]string, error) {
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

func (s *Etcd) Watch(prefix string) clientv3.WatchChan {
	return s.cli.Watch(context.Background(), prefix,
		clientv3.WithPrefix(), clientv3.WithPrevKV())

}
