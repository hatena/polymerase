package api

import (
	"net"
	"github.com/taku-k/xtralab/pkg/config"
	"google.golang.org/grpc"
	pb "github.com/taku-k/xtralab/pkg/backup/proto"
	"github.com/taku-k/xtralab/pkg/backup"
	"github.com/taku-k/xtralab/pkg/storage"
)


func NewgRPCServer(conf *config.Config) {
	lis, err := net.Listen("tcp", ":10110")
	if err != nil {
		return
	}
	server := grpc.NewServer()

	s, err := storage.NewLocalBackupStorage(conf)
	if err != nil {
		panic(err)
	}
	m := backup.NewTempBackupManager(s, conf)
	svc := backup.NewBackupTransferService(m)

	pb.RegisterBackupTransferServiceServer(server, svc)
	server.Serve(lis)
}