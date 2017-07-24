package server

import (
	"context"

	"github.com/taku-k/polymerase/pkg/server/admin"
	"google.golang.org/grpc"
)

type adminServer struct {
	server *Server
}

func newAdminServer(s *Server) *adminServer {
	return &adminServer{s}
}

func (s *adminServer) RegisterService(g *grpc.Server) {
	admin.RegisterAdminServer(g, s)
}

func (s *adminServer) Backups(
	ctx context.Context, req *admin.BackupsRequest,
) (*admin.BackupsResponse, error) {
	return &admin.BackupsResponse{}, nil
}
