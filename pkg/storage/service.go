package storage

import (
	"errors"
	"io"
	"time"

	"github.com/fujiwara/shapeio"
	log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"golang.org/x/net/context"
)

type StorageService struct {
	storage   BackupStorage
	rateLimit float64
}

func NewStorageService(storage BackupStorage, rateLimit uint64) *StorageService {
	return &StorageService{
		storage:   storage,
		rateLimit: float64(rateLimit),
	}
}

func (s *StorageService) GetLatestToLSN(
	ctx context.Context, req *storagepb.GetLatestToLSNRequest,
) (*storagepb.GetLatestToLSNResponse, error) {
	lsn, err := s.storage.GetLatestToLSN(req.Db)
	if err != nil {
		log.WithField("db", req.Db).Info("Not found")
		return &storagepb.GetLatestToLSNResponse{Lsn: ""}, errors.New("Not found such a db")
	}
	return &storagepb.GetLatestToLSNResponse{
		Lsn: lsn,
	}, nil
}

func (s *StorageService) GetKeysAtPoint(
	ctx context.Context, req *storagepb.GetKeysAtPointRequest,
) (*storagepb.GetKeysAtPointResponse, error) {
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

func (s *StorageService) GetFileByKey(
	req *storagepb.GetFileByKeyRequest, stream storagepb.StorageService_GetFileByKeyServer,
) error {
	r, err := s.storage.GetFileStream(req.Key)
	if err != nil {
		return err
	}
	sr := shapeio.NewReader(r)
	sr.SetRateLimit(s.rateLimit)
	chunk := make([]byte, 1<<20)
	for {
		n, err := sr.Read(chunk)
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

func (s *StorageService) PurgePrevBackup(
	ctx context.Context, req *storagepb.PurgePrevBackupRequest,
) (*storagepb.PurgePrevBackupResponse, error) {
	key, err := s.storage.GetKPastBackupKey(req.Db, 2)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{
			Message: "There is no backup to purge.",
		}, nil
	}
	log.WithField("key", key).Info("Purge key")
	err = s.storage.RemoveBackups(key)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{}, err
	}
	return &storagepb.PurgePrevBackupResponse{
		Message: "Purge succeeds",
	}, nil
}
