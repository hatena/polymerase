package storage

import (
	"errors"
	"time"

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

func (s *StorageService) GetKeysAtPoint(ctx context.Context, req *storagepb.GetKeysAtPointRequest) (*storagepb.GetKeysAtPointResponse, error) {
	t, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return &storagepb.GetKeysAtPointResponse{}, err
	}
	t = t.AddDate(0, 0, 1)
	bfiles, _ := s.storage.SearchConsecutiveIncBackups(req.Db, t)
	keys := make([]*storagepb.BackupFileInfo, len(bfiles), len(bfiles))
	for i, f := range bfiles {
		keys[i] = &storagepb.BackupFileInfo{
			Key:         f.Key,
			StorageType: f.StorageType,
			BackupType:  f.BackupType,
		}
	}
	return &storagepb.GetKeysAtPointResponse{
		Keys: keys,
	}, nil
}
