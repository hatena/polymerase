package api

import (
	"net"

	"github.com/taku-k/xtralab/pkg/backup"
	pb "github.com/taku-k/xtralab/pkg/backup/proto"
	"github.com/taku-k/xtralab/pkg/config"
	"github.com/taku-k/xtralab/pkg/storage"
	"google.golang.org/grpc"
)

func NewgRPCServer(conf *config.Config, l net.Listener) {
	//lis, err := net.Listen("tcp", ":10110")
	//if err != nil {
	//	return
	//}
	server := grpc.NewServer()

	s, err := storage.NewLocalBackupStorage(conf)
	if err != nil {
		panic(err)
	}
	m := backup.NewTempBackupManager(s, conf)
	svc := backup.NewBackupTransferService(m)

	pb.RegisterBackupTransferServiceServer(server, svc)
	server.Serve(l)
}
