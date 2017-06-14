package server

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/utils/log"
)

type etcdServer struct {
	Server     *embed.Etcd
	cfg        *embed.Config
	ClientPort string
}

func newEtcdEmbedConfig(ctx *EtcdContext) (*embed.Config, error) {
	etcdCfg := embed.NewConfig()
	lcurl, err := url.Parse(fmt.Sprintf("http://0.0.0.0:%s", ctx.ClientPort))
	if err != nil {
		return nil, errors.Wrap(err, "port cannot be parsed")
	}
	etcdCfg.LCUrls = []url.URL{*lcurl}

	acurl, err := url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(ctx.Host, ctx.ClientPort)))
	if err != nil {
		return nil, errors.Wrap(err, "port cannot be parsed")
	}
	etcdCfg.ACUrls = []url.URL{*acurl}

	lpurl, err := url.Parse(fmt.Sprintf("http://0.0.0.0:%s", ctx.PeerPort))
	if err != nil {
		return nil, errors.Wrap(err, "etcd peer port cannot be parsed")
	}
	etcdCfg.LPUrls = []url.URL{*lpurl}

	apurl, err := url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(ctx.Host, ctx.PeerPort)))
	if err != nil {
		return nil, errors.Wrap(err, "etcd peer port cannot be parsed")
	}
	etcdCfg.APUrls = []url.URL{*apurl}

	etcdCfg.Dir = ctx.DataDir

	etcdCfg.Name = ctx.Name

	if ctx.isInitialCluster() {
		etcdCfg.InitialCluster = etcdCfg.InitialClusterFromName(ctx.Name)
	} else if !ctx.existsDataDir() {
		// If data dir exists, it is launched with recovery mode.
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

func newEtcdServer(cfg *embed.Config) (*etcdServer, error) {
	es := &etcdServer{}
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		os.RemoveAll(cfg.Dir)
		return nil, err
	}
	es.Server = e
	es.cfg = cfg
	return es, nil
}

func (e *etcdServer) close() {
	defer os.RemoveAll(e.Server.Config().Dir)
	ep := make([]string, len(e.cfg.ACUrls))
	for i, e := range e.cfg.ACUrls {
		ep[i] = e.String()
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: ep,
		// TODO: assign with default value, not hard coding
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Info("Hard shutdown")
	} else {
		res, err := cli.MemberRemove(context.Background(), uint64(e.Server.Server.ID()))
		if err != nil {
			log.Info("Failed to remove myself")
			e.Server.Server.TransferLeadership()
		} else {
			log.Info("Success to remove myself")
			log.Info(res)
		}
	}
	e.Server.Close()
}
