package etcd

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

type ClientAPI interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	Locker(key string) sync.Locker
	Close()
}

type Client struct {
	cli     *clientv3.Client
	session *concurrency.Session
}

func NewClient(cfg clientv3.Config) (ClientAPI, error) {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	session, err := concurrency.NewSession(cli)
	return &Client{
		cli:     cli,
		session: session,
	}, err
}

func NewTestClient(cli *clientv3.Client) (ClientAPI, error) {
	session, err := concurrency.NewSession(cli)
	return &Client{
		cli:     cli,
		session: session,
	}, err
}

func (c *Client) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return c.cli.KV.Get(ctx, key, opts...)
}

func (c *Client) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	return c.cli.KV.Put(ctx, key, val, opts...)
}

func (c *Client) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return c.cli.KV.Delete(ctx, key, opts...)
}

func (c *Client) Locker(key string) sync.Locker {
	return concurrency.NewLocker(c.session, key)
}

func (c *Client) Close() {
	c.cli.Close()
}
