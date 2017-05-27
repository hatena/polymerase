package server

import (
	"context"
	"net"
	"time"

	"github.com/soheilhy/cmux"
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
}

func NewServer(cfg *Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	var err error

	s.grpc = grpc.NewServer()

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
	tempbackuppb.RegisterBackupTransferServiceServer(s.grpc, s.tempBackupSvc)

	s.storageSvc = storage.NewStorageService(s.storage)
	storagepb.RegisterStorageServiceServer(s.grpc, s.storageSvc)

	return s, nil
}

func (s *Server) Start(ctx context.Context) error {
	l, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}

	m := cmux.New(l)

	grpcl := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	go s.grpc.Serve(grpcl)
	go func() {
		e, err := etcd.NewEtcdServer()
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
	}()

	if err := m.Serve(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	s.grpc.GracefulStop()
	stopped <- struct{}{}
}
