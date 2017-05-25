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

func (s *StorageService) GetLatestToLSN(ctx context.Context, req *pb.GetLatestToLSNRequest) (*pb.GetLatestToLSNResponse, error) {
	lsn, err := s.storage.GetLastLSN(req.Db)
	if err != nil {
		log.WithField("db", req.Db).Info("Not found")
		return &pb.GetLatestToLSNResponse{Lsn: ""}, errors.New("Not found such a db")
	}
	return &pb.GetLatestToLSNResponse{
		Lsn: lsn,
	}, nil
}
