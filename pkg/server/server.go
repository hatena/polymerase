package server

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/storage"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"google.golang.org/grpc"
)

type Server struct {
	cfg           *Config
	grpc          *grpc.Server
	storage       storage.BackupStorage
	manager       *tempbackup.TempBackupManager
	tempBackupSvc *tempbackup.TempBackupTransferService
	storageSvc    *storage.StorageService
	etcdServer    *etcdServer
	etcdCfg       *embed.Config
}

func NewServer(cfg *Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	etcdCfg, err := newEtcdEmbedConfig(&EtcdContext{
		Host:       cfg.Host,
		ClientPort: cfg.Port,
		PeerPort:   cfg.EtcdPeerPort,
		DataDir:    cfg.EtcdDataDir(),
		JoinAddr:   cfg.JoinAddr,
		Name:       cfg.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "etcd embed config cannot be created")
	}
	s.etcdCfg = etcdCfg

	// For now, local storage only
	s.storage, err = storage.NewLocalBackupStorage(&storage.LocalStorageConfig{
		Config:         cfg.Config,
		BackupsDir:     cfg.BackupsDir(),
		ServeRateLimit: cfg.ServeRateLimit,
	})
	if err != nil {
		return nil, errors.Wrap(err, "backup storage configuration is failed")
	}

	mngr, err := tempbackup.NewTempBackupManager(s.storage, &tempbackup.TempBackupManagerConfig{
		Config:  cfg.Config,
		TempDir: cfg.TempDir(),
		Name:    cfg.Name,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup TempBackupManager")
	}
	s.manager = mngr

	s.tempBackupSvc = tempbackup.NewBackupTransferService(s.manager)

	s.storageSvc = storage.NewStorageService(s.storage, cfg.ServeRateLimit)

	s.etcdCfg.ServiceRegister = func(gs *grpc.Server) {
		tempbackuppb.RegisterBackupTransferServiceServer(gs, s.tempBackupSvc)
		storagepb.RegisterStorageServiceServer(gs, s.storageSvc)
	}

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	es, err := newEtcdServer(s.etcdCfg)
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
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{net.JoinHostPort(s.cfg.Host, s.cfg.Port)},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	s.manager.EtcdCli = cli

	// Start status sampling
	go s.startWriteStatus(s.cfg.StatusSampleInterval)

	log.Println(<-es.Server.Err())

	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	if s.etcdServer != nil {
		s.etcdServer.close()
	}
	stopped <- struct{}{}
}

func (s *Server) startWriteStatus(freq time.Duration) {
	recorder := status.NewStatusRecorder(
		s.manager.EtcdCli, s.cfg.StoreDir, s.storage, s.cfg.Name, s.cfg.Config)

	// Do WriteStatus before ticker starts
	if err := recorder.WriteStatus(context.Background()); err != nil {
		log.Println(err)
		return
	}

	ticker := time.NewTicker(freq)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := recorder.WriteStatus(context.Background()); err != nil {
				log.Println(err)
			}
		}
	}
}
