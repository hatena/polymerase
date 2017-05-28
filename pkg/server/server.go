package server

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/coreos/etcd/embed"
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
	etcdServer    *embed.Etcd
	etcdCfg       *embed.Config
}

func NewServer(cfg *Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	var err error

	//s.grpc = grpc.NewServer()

	etcdCfg := embed.NewConfig()
	lcurl, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", cfg.Port))
	if err != nil {
		return nil, err
	}
	etcdCfg.LCUrls = []url.URL{*lcurl}

	acurl, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", cfg.Port))
	if err != nil {
		return nil, err
	}
	etcdCfg.ACUrls = []url.URL{*acurl}

	lpurl, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", cfg.EtcdPeerPort))
	if err != nil {
		return nil, err
	}
	etcdCfg.LPUrls = []url.URL{*lpurl}

	apurl, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%s", cfg.EtcdPeerPort))
	if err != nil {
		return nil, err
	}
	etcdCfg.APUrls = []url.URL{*apurl}

	etcdCfg.Dir = cfg.EtcdDataDir

	// For now, local storage only
	s.storage, err = storage.NewLocalBackupStorage(&storage.LocalStorageConfig{
		Config:     cfg.Config,
		BackupsDir: cfg.BackupsDir,
	})
	if err != nil {
		return nil, err
	}

	s.manager = tempbackup.NewTempBackupManager(s.storage, &tempbackup.TempBackupManagerConfig{
		Config:  cfg.Config,
		TempDir: cfg.TempDir,
	})

	s.tempBackupSvc = tempbackup.NewBackupTransferService(s.manager)
	//tempbackuppb.RegisterBackupTransferServiceServer(s.grpc, s.tempBackupSvc)

	s.storageSvc = storage.NewStorageService(s.storage)
	//storagepb.RegisterStorageServiceServer(s.grpc, s.storageSvc)

	etcdCfg.ServiceRegister = func(gs *grpc.Server) {
		tempbackuppb.RegisterBackupTransferServiceServer(gs, s.tempBackupSvc)
		storagepb.RegisterStorageServiceServer(gs, s.storageSvc)
	}

	s.etcdCfg = etcdCfg

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	//l, err := net.Listen("tcp", s.cfg.Addr)
	//if err != nil {
	//	return err
	//}
	//
	//m := cmux.New(l)
	//
	//grpcl := m.MatchWithWriters(
	//	cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	//
	//go s.grpc.Serve(grpcl)
	go func() {

	}()

	e, err := etcd.NewEtcdServer(s.etcdCfg)
	s.etcdServer = e
	if err != nil {
		log.Info(err)
	}
	defer e.Close()
	select {
	case <-e.Server.ReadyNotify():
		log.Info("Server is ready")
	case <-time.After(60 * time.Second):
		e.Server.Stop()
		log.Info("Server took too long to start")
	}
	log.Info(<-e.Err())
	//
	//if err := m.Serve(); err != nil {
	//	return err
	//}
	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	s.etcdServer.Close()
	stopped <- struct{}{}
}
