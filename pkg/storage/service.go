package storage

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
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

func (s *StorageService) GetLatestToLSN(ctx context.Context, req *storagepb.GetLatestToLSNRequest) (*storagepb.GetLatestToLSNResponse, error) {
	lsn, err := s.storage.GetLatestToLSN(req.Db)
	if err != nil {
		log.WithField("db", req.Db).Info("Not found")
		return &storagepb.GetLatestToLSNResponse{Lsn: ""}, errors.New("Not found such a db")
	}
	return &storagepb.GetLatestToLSNResponse{
		Lsn: lsn,
	}, nil
}
