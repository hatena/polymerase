package server

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
)

type EtcdContext struct {
	Host       string
	ClientPort string
	PeerPort   string
	DataDir    string
	JoinAddr   string
	Name       string
}

func (c *EtcdContext) isInitialCluster() bool {
	return c.JoinAddr == ""
}

func (c *EtcdContext) existsDataDir() bool {
	_, err := os.Stat(c.DataDir)
	return err == nil
}

func (c *EtcdContext) AddMember(peerUrl string) (string, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{c.JoinAddr},
		// TODO: assign with default value, not hard coding
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
	log.Println(res.Members)
	newID := res.Member.ID
	var buf bytes.Buffer
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
