package etcd

import (
	"fmt"
	"net/url"

	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
)

type EtcdContext struct {
	ClientPort string
	PeerPort   string
	DataDir    string
}

func NewEtcdEmbedConfig(ctx *EtcdContext) (*embed.Config, error) {
	etcdCfg := embed.NewConfig()
	lcurl, err := url.Parse(fmt.Sprintf("http://localhost:%s", ctx.ClientPort))
	if err != nil {
		return nil, errors.Wrap(err, "port cannot be parsed")
	}
	etcdCfg.LCUrls = []url.URL{*lcurl}

	acurl, err := url.Parse(fmt.Sprintf("http://localhost:%s", ctx.ClientPort))
	if err != nil {
		return nil, errors.Wrap(err, "port cannot be parsed")
	}
	etcdCfg.ACUrls = []url.URL{*acurl}

	lpurl, err := url.Parse(fmt.Sprintf("http://localhost:%s", ctx.PeerPort))
	if err != nil {
		return nil, errors.Wrap(err, "etcd peer port cannot be parsed")
	}
	etcdCfg.LPUrls = []url.URL{*lpurl}

	apurl, err := url.Parse(fmt.Sprintf("http://localhost:%s", ctx.PeerPort))
	if err != nil {
		return nil, errors.Wrap(err, "etcd peer port cannot be parsed")
	}
	etcdCfg.APUrls = []url.URL{*apurl}

	etcdCfg.Dir = ctx.DataDir

	etcdCfg.InitialCluster = etcdCfg.InitialClusterFromName("")

	return etcdCfg, nil
}

func NewEtcdServer(cfg *embed.Config) (*embed.Etcd, error) {
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	return e, nil
}
