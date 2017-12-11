package etcd

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/polypb"
)

type ClientAPI interface {
	GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)
	PutBackupMeta(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error
	RemoveBackupMeta(key polypb.BackupMetaKey) error
	UpdateLSN(key polypb.BackupMetaKey, lsn string) error

	GetNodeMeta(key polypb.NodeMetaKey) ([]*polypb.NodeMeta, error)
	PutNodeMeta(key polypb.NodeMetaKey, meta *polypb.NodeMeta) error
	RemoveNodeMeta(key polypb.NodeMetaKey) error
	//Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	//Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	//Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	//Locker(key string) sync.Locker
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

func (c *Client) GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
	res, err := c.cli.KV.Get(context.TODO(), string(key),
		clientv3.WithPrefix(), clientv3.WithIgnoreLease())
	if err != nil {
		return nil, err
	}
	result := make(polypb.BackupMetaSlice, len(res.Kvs))
	for i, kv := range res.Kvs {
		meta := &polypb.BackupMeta{}
		if err := proto.Unmarshal(kv.Value, meta); err != nil {
			return nil, err
		}
		result[i] = meta
	}
	return result, nil
}

func (c *Client) PutBackupMeta(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error {
	data, err := meta.Marshal()
	if err != nil {
		return err
	}
	_, err = c.cli.KV.Put(context.TODO(), string(key), string(data))
	return err
}

func (c *Client) RemoveBackupMeta(key polypb.BackupMetaKey) error {
	_, err := c.cli.KV.Delete(context.TODO(), string(key), clientv3.WithPrefix())
	return err
}

func (c *Client) UpdateLSN(key polypb.BackupMetaKey, lsn string) error {
	locker := c.Locker("lock-" + string(key))
	locker.Lock()
	defer locker.Unlock()

	metas, err := c.GetBackupMeta(key)
	if err != nil {
		return err
	}
	if len(metas) != 1 {
		return errors.New(fmt.Sprintf("fetched wrong metadata: %q", metas))
	}
	m := metas[0]
	m.ToLsn = lsn
	return c.PutBackupMeta(key, m)
}

func (c *Client) GetNodeMeta(key polypb.NodeMetaKey) ([]*polypb.NodeMeta, error) {
	res, err := c.cli.KV.Get(context.TODO(), string(key), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	result := make([]*polypb.NodeMeta, len(res.Kvs))
	for i, kv := range res.Kvs {
		meta := &polypb.NodeMeta{}
		if err := proto.Unmarshal(kv.Value, meta); err != nil {
			return nil, err
		}
		result[i] = meta
	}
	return result, nil
}

func (c *Client) PutNodeMeta(key polypb.NodeMetaKey, meta *polypb.NodeMeta) error {
	data, err := meta.Marshal()
	if err != nil {
		return err
	}
	_, err = c.cli.KV.Put(context.TODO(), string(key), string(data))
	return err
}

func (c *Client) RemoveNodeMeta(key polypb.NodeMetaKey) error {
	_, err := c.cli.KV.Delete(context.TODO(), string(key), clientv3.WithPrefix())
	return err
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
