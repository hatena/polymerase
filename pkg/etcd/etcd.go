package etcd

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/utils/log"
)

type EtcdContext struct {
	ClientPort string
	PeerPort   string
	DataDir    string
	JoinAddr   string
	Name       string
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

	etcdCfg.Name = ctx.Name

	if ctx.isInitialCluster() {
		etcdCfg.InitialCluster = etcdCfg.InitialClusterFromName(ctx.Name)
	} else {
		cluster, err := ctx.AddMember(apurl.String())
		if err != nil {
			return nil, errors.Wrap(err, "AddMember API is failed")
		}
		log.Info(cluster)
		etcdCfg.ClusterState = embed.ClusterStateFlagExisting
		etcdCfg.InitialCluster = cluster
	}

	return etcdCfg, nil
}

func (c *EtcdContext) isInitialCluster() bool {
	return c.JoinAddr == ""
}

func (c *EtcdContext) AddMember(peerUrl string) (string, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{c.JoinAddr},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return "", err
	}
	defer cli.Close()
	res, err := cli.MemberAdd(context.Background(), []string{peerUrl})
	if err != nil {
		return "", err
	}
	log.Info(res.Members)
	newID := res.Member.ID
	var buf bytes.Buffer
	//fmt.Fprintf(&buf, "%s=%s", res.Member.Name, peerUrl)
	for _, m := range res.Members {
		for _, u := range m.PeerURLs {
			n := m.Name
			if m.ID == newID {
				n = c.Name
			}
			fmt.Fprintf(&buf, "%s=%s,", n, u)
		}
	}
	if l := buf.Len(); l > 0 {
		buf.Truncate(l - 1)
	}
	return buf.String(), nil
}

func NewEtcdServer(cfg *embed.Config) (*embed.Etcd, error) {
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	return e, nil
}
