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
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	_, err := s.cli.Put(ctx, key, string(b))
	return err
}

func (s *Etcd) Get(key string, obj interface{}) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	resp, err := s.cli.Get(ctx, key)
	if err != nil {
		return err
	}
	if len(resp.Kvs) <= 0 {
		return fmt.Errorf("empty value")
	}

	return json.Unmarshal(resp.Kvs[0].Value, obj)
}

func (s *Etcd) Del(key string) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	s.cli.Delete(ctx, key)
}

func (s *Etcd) DelPrefix(prefix string) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	s.cli.Delete(ctx, prefix, clientv3.WithPrefix())
}

func (s *Etcd) List(root string) (map[string]string, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	defer cancel()
	resp, err := s.cli.Get(ctx, root, clientv3.WithPrefix())
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
