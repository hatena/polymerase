package server

import (
	"context"
	"net"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/soheilhy/cmux"
	"github.com/taku-k/polymerase/pkg/api"
	"github.com/taku-k/polymerase/pkg/storage"
	storagepb "github.com/taku-k/polymerase/pkg/storage/proto"
	"github.com/taku-k/polymerase/pkg/tempbackup"
	tempbackuppb "github.com/taku-k/polymerase/pkg/tempbackup/proto"
	"google.golang.org/grpc"
)

type Server struct {
	cfg           *Config
	grpc          *grpc.Server
	storage       storage.BackupStorage
	manager       *tempbackup.TempBackupManager
	tempBackupSvc *tempbackup.TempBackupTransferService
	storageSvc    *storage.StorageService

	// Deprecated
	app *api.App
}

func NewServer(cfg *Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
	}

	var err error

	s.grpc = grpc.NewServer()

	// For now, local storage only
	s.storage, err = storage.NewLocalBackupStorage(cfg.Config)
	if err != nil {
		return nil, err
	}

	s.manager = tempbackup.NewTempBackupManager(s.storage, cfg.Config)

	s.tempBackupSvc = tempbackup.NewBackupTransferService(s.manager)
	tempbackuppb.RegisterBackupTransferServiceServer(s.grpc, s.tempBackupSvc)

	s.storageSvc = storage.NewStorageService(s.storage)
	storagepb.RegisterStorageServiceServer(s.grpc, s.storageSvc)

	apiCfg := &api.Config{
		Config:        cfg.Config,
		HTTPApiPrefix: cfg.HTTPApiPrefix,
	}
	s.app, err = api.NewApp(apiCfg)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.cfg.Addr)
	if err != nil {
		return err
	}

	m := cmux.New(l)

	httpl := m.Match(cmux.HTTP1Fast())
	grpcl := m.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))

	go s.grpc.Serve(grpcl)
	go s.app.Run(httpl)

	if err := m.Serve(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context, stopped chan struct{}) {
	go s.grpc.GracefulStop()
	if err := s.app.Engine.Shutdown(ctx); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	stopped <- struct{}{}
}
