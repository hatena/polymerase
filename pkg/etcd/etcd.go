package etcd

import "github.com/coreos/etcd/embed"

func NewEtcdServer() (*embed.Etcd, error) {
	cfg := embed.NewConfig()
	cfg.Dir = "default.etcd"
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	return e, nil
}
