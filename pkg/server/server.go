package server

import (
	"context"
	"time"

	"github.com/coreos/etcd/embed"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/storage"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/tempbackup"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"github.com/taku-k/polymerase/pkg/utils/log"
	"google.golang.org/grpc"
)

type Server struct {
	cfg           *Config
	grpc          *grpc.Server
	storage       storage.BackupStorage
	manager       *tempbackup.TempBackupManager
	tempBackupSvc *tempbackup.TempBackupTransferService
	storageSvc    *storage.StorageService
	etcdServer    *etcd.EtcdServer
	etcdCfg       *embed.Config
}

func NewServer(cfg *Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	etcdCfg, err := etcd.NewEtcdEmbedConfig(&etcd.EtcdContext{
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
		Config:     cfg.Config,
		BackupsDir: cfg.BackupsDir(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "backup storage configuration is failed")
	}

	mngr, err := tempbackup.NewTempBackupManager(s.storage, &tempbackup.TempBackupManagerConfig{
		Config:  cfg.Config,
		TempDir: cfg.TempDir(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to setup TempBackupManager")
	}
	s.manager = mngr

	s.tempBackupSvc = tempbackup.NewBackupTransferService(s.manager)

	s.storageSvc = storage.NewStorageService(s.storage)

	s.etcdCfg.ServiceRegister = func(gs *grpc.Server) {
		tempbackuppb.RegisterBackupTransferServiceServer(gs, s.tempBackupSvc)
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
		log.Info("Server is ready")
	case <-time.After(60 * time.Second):
		es.Server.Server.Stop()
		log.Info("Server took too long to start")
	}
	log.Info(<-es.Server.Err())

	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	if s.etcdServer != nil {
		s.etcdServer.Close()
	}
	stopped <- struct{}{}
}
