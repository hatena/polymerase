package storage

import (
	"errors"

	log "github.com/sirupsen/logrus"
	pb "github.com/taku-k/polymerase/pkg/storage/proto"
	"golang.org/x/net/context"
)

type StorageService struct {
	storage BackupStorage
}

func NewStorageService(storage BackupStorage) *StorageService {
	return &StorageService{
		storage: storage,
	}
}

func (s *StorageService) GetLastLSN(ctx context.Context, req *pb.GetLastLSNRequest) (*pb.GetLastLSNResponse, error) {
	lsn, err := s.storage.GetLastLSN(req.Db)
	if err != nil {
		log.WithField("db", req.Db).Info("Not found")
		return &pb.GetLastLSNResponse{Lsn: ""}, errors.New("Not found such a db")
	}
	return &pb.GetLastLSNResponse{
		Lsn: lsn,
	}, nil
}
