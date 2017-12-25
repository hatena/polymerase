package server

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/storage"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type Server struct {
	cfg           *base.ServerConfig
	grpc          *grpc.Server
	backupManager *storage.BackupManager
	storageSvc    *storage.Service
	etcdServer    *etcd.EtcdServer
	etcdCfg       *embed.Config
}

func NewServer(cfg *base.ServerConfig) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	etcdCfg, err := etcd.NewEtcdEmbedConfig(&etcd.Context{
		Host:       cfg.Host,
		ClientPort: cfg.Port,
		PeerPort:   cfg.EtcdPeerPort,
		DataDir:    cfg.EtcdDataDir(),
		JoinAddr:   cfg.JoinAddr,
		NodeID:     cfg.NodeID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "etcd embed config cannot be created")
	}
	s.etcdCfg = etcdCfg

	s.backupManager = storage.NewBackupManager(s.cfg)

	s.storageSvc = storage.NewService(s.backupManager, cfg.ServeRateLimit, s.cfg)

	s.etcdCfg.ServiceRegister = func(gs *grpc.Server) {
		storagepb.RegisterStorageServiceServer(gs, s.storageSvc)
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	es, err := etcd.NewEtcdServer(s.etcdCfg)
	s.etcdServer = es
	if err != nil {
		return errors.Wrap(err, "etcd server cannot be started")
	}
	defer es.Server.Close()
	select {
	case <-es.Server.Server.ReadyNotify():
		log.Println("Server is ready")
	case <-time.After(60 * time.Second):
		es.Server.Server.Stop()
		log.Println("Server took too long to start")
	}

	// Create etcd client
	cli, err := etcd.NewClient(clientv3.Config{
		Endpoints:   []string{net.JoinHostPort(s.cfg.Host, s.cfg.Port)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}

	// Inject etcd client after launching embedded etcd server
	s.storageSvc.EtcdCli = cli
	s.backupManager.EtcdCli = cli

	// Start status sampling
	go s.startWriteStatus(s.cfg.StatusSampleInterval)
	if err := s.backupManager.RestoreBackupInfo(); err != nil {
		log.Print(err)
		cli.Close()
		return err
	}

	log.Println(<-es.Server.Err())

	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	if s.storageSvc.EtcdCli != nil {
		s.storageSvc.EtcdCli.Close()
	}
	if s.etcdServer != nil {
		s.etcdServer.Close()
	}
	stopped <- struct{}{}
}

func (s *Server) CleanupEtcdDir() {
	os.RemoveAll(s.etcdCfg.Dir)
}

func (s *Server) startWriteStatus(freq time.Duration) {
	recorder := newStatusRecorder(
		s.backupManager.EtcdCli, s.cfg.StoreDir.Path, s.cfg.NodeID, s.cfg)

	// Do WriteStatus before ticker starts
	if err := recorder.writeStatus(context.Background()); err != nil {
		log.Println(err)
		return
	}

	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := recorder.writeStatus(context.Background()); err != nil {
				log.Println(err)
			}
		}
	}
}
