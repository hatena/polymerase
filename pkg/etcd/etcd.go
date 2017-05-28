package etcd

import "github.com/coreos/etcd/embed"

func NewEtcdServer(cfg *embed.Config) (*embed.Etcd, error) {
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}
	return e, nil
}
