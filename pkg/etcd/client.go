package etcd

import (
	"context"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/hatena/polymerase/pkg/polypb"
)

type ClientAPI interface {
	GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)
	PutBackupMeta(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error
	RemoveBackupMeta(key polypb.BackupMetaKey) error

	GetNodeMeta(key polypb.NodeMetaKey) ([]*polypb.NodeMeta, error)
	PutNodeMeta(key polypb.NodeMetaKey, meta *polypb.NodeMeta) error
	RemoveNodeMeta(key polypb.NodeMetaKey) error

	Close()
}

type Client struct {
	cli     *clientv3.Client
	session *concurrency.Session
}

func NewClient(cfg clientv3.Config) (ClientAPI, error) {
	cli, err := clientv3.New(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "Creating etcd client is failed")
	}
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, errors.Wrapf(err, "Creating session is failed")
	}
	return &Client{
		cli:     cli,
		session: session,
	}, nil
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

func (c *Client) Close() {
	c.cli.Close()
	c.session.Close()
}
