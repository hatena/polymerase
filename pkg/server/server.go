package server

import (
	"context"
	"net"

	"github.com/soheilhy/cmux"
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

	if err := m.Serve(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	s.grpc.GracefulStop()
	stopped <- struct{}{}
}
