package storage

import (
	"errors"
	"io"
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
	return &storagepb.GetKeysAtPointResponse{
		Keys: bfiles,
	}, nil
}

func (s *StorageService) GetFileByKey(req *storagepb.GetFileByKeyRequest, stream storagepb.StorageService_GetFileByKeyServer) error {
	r, err := s.storage.GetFileStream(req.Key)
	if err != nil {
		return err
	}
	chunk := make([]byte, 1<<20)
	for {
		n, err := r.Read(chunk)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		stream.Send(&storagepb.FileStream{
			Content: chunk[:n],
		})
	}
}
