package storage

import (
	"bytes"
	"errors"
	"io"
	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type StorageService struct {
	manager   *BackupManager
	rateLimit float64
	EtcdCli   etcd.ClientAPI
	tempMngr  *TempBackupManager
	cfg       *base.ServerConfig
}

func NewStorageService(
	manager *BackupManager,
	rateLimit uint64,
	tempMngr *TempBackupManager,
	cfg *base.ServerConfig,
) *StorageService {
	return &StorageService{
		manager:   manager,
		rateLimit: float64(rateLimit),
		tempMngr:  tempMngr,
		cfg:       cfg,
	}
}

func (s *StorageService) GetLatestToLSN(
	ctx context.Context, req *storagepb.GetLatestToLSNRequest,
) (*storagepb.GetLatestToLSNResponse, error) {
	db := polypb.DatabaseID(req.Db)
	lsn, err := s.manager.GetLatestToLSN(db)
	return &storagepb.GetLatestToLSNResponse{
		Lsn: lsn,
	}, err
}

func (s *StorageService) GetKeysAtPoint(
	ctx context.Context, req *storagepb.GetKeysAtPointRequest,
) (*storagepb.GetKeysAtPointResponse, error) {
	t, err := time.Parse("2006-01-02", req.From)
	if err != nil {
		return &storagepb.GetKeysAtPointResponse{}, err
	}
	t = t.AddDate(0, 0, 1)
	db := polypb.DatabaseID(req.Db)
	bfiles, _ := s.manager.SearchConsecutiveIncBackups(db, t)
	return &storagepb.GetKeysAtPointResponse{
		Keys: bfiles,
	}, nil
}

func (s *StorageService) GetFileByKey(
	req *storagepb.GetFileByKeyRequest, stream storagepb.StorageService_GetFileByKeyServer,
) error {
	r, err := s.manager.GetFileStream(polypb.Key(req.Key))
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

func (s *StorageService) PurgePrevBackup(
	ctx context.Context, req *storagepb.PurgePrevBackupRequest,
) (*storagepb.PurgePrevBackupResponse, error) {
	key, err := s.manager.GetKPastBackupKey(
		polypb.DatabaseID(req.Db), 2)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{
			Message: "There is no backup to purge.",
		}, nil
	}
	log.Printf("Purge key=%s\n", key)
	err = s.manager.RemoveBackups(key)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{}, err
	}
	return &storagepb.PurgePrevBackupResponse{
		Message: "Purge succeeds",
	}, nil
}

func (s *StorageService) TransferFullBackup(
	stream storagepb.StorageService_TransferFullBackupServer,
) error {
	var tempBackup appendCloser

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			meta, err := tempBackup.CloseTransfer()
			if err != nil {
				return err
			}
			key := keys.MakeBackupMetaKeyFromKey(meta.Key)
			if err := s.EtcdCli.PutBackupMeta(key, meta); err != nil {
				return err
			}
			return stream.SendAndClose(&storagepb.BackupReply{
				Message: "success",
				Key:     meta.Key,
			})
		}
		if err != nil {
			return err
		}
		if tempBackup == nil {
			if content.Db == nil {
				return errors.New("empty db is not acceptable")
			}
			tempBackup, err = s.tempMngr.openTempBackup(content.Db, &xtrabackupFullRequest{})
			if err != nil {
				return err
			}
			log.Printf("Start full-backup: db=%s\n", content.Db)
		}
		if err := tempBackup.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *StorageService) TransferIncBackup(
	stream storagepb.StorageService_TransferIncBackupServer,
) error {
	var tempBackup appendCloser

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			meta, err := tempBackup.CloseTransfer()
			if err != nil {
				return err
			}
			key := keys.MakeBackupMetaKeyFromKey(meta.Key)
			if err := s.EtcdCli.PutBackupMeta(key, meta); err != nil {
				return err
			}
			return stream.SendAndClose(&storagepb.BackupReply{
				Message: "success",
				Key:     meta.Key,
			})
		}
		if err != nil {
			return err
		}
		if tempBackup == nil {
			if content.Db == nil {
				return errors.New("empty db is not acceptable")
			}
			if content.Lsn == "" {
				return errors.New("empty lsn is not acceptable")
			}
			tempBackup, err = s.tempMngr.openTempBackup(
				content.Db,
				&xtrabackupIncRequest{
					LSN: content.Lsn,
				})
			if err != nil {
				return err
			}
			log.Printf("Start inc-backup: db=%s\n", content.Db)
		}
		if err := tempBackup.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *StorageService) TransferMysqldump(
	stream storagepb.StorageService_TransferMysqldumpServer,
) error {
	var tempBackup appendCloser

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			meta, err := tempBackup.CloseTransfer()
			if err != nil {
				return err
			}
			key := keys.MakeBackupMetaKeyFromKey(meta.Key)
			if err := s.EtcdCli.PutBackupMeta(key, meta); err != nil {
				return err
			}
			return stream.SendAndClose(&storagepb.BackupReply{
				Message: "success",
				Key:     meta.Key,
			})
		}
		if err != nil {
			return err
		}
		if tempBackup == nil {
			if content.Db == nil {
				return errors.New("empty db is not acceptable")
			}
			tempBackup, err = s.tempMngr.openTempBackup(
				content.Db, &mysqldumpRequest{})
			if err != nil {
				return err
			}
			log.Printf("Start mysqldump: db=%s\n", content.Db)
		}
		if err := tempBackup.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *StorageService) PostCheckpoints(
	ctx context.Context,
	req *storagepb.PostCheckpointsRequest,
) (*storagepb.PostCheckpointsResponse, error) {
	r := bytes.NewReader(req.Content)
	key := polypb.Key(req.Key)
	if err := s.manager.PostFile(key, "xtrabackup_checkpoints", r); err != nil {
		return &storagepb.PostCheckpointsResponse{}, err
	}
	cp := base.LoadXtrabackupCP(req.Content)
	if cp.ToLSN == "" {
		return nil, errors.New("failed to load")
	}
	err := s.EtcdCli.UpdateLSN(keys.MakeBackupMetaKeyFromKey(key), cp.ToLSN)
	if err != nil {
		return nil, err
	}
	return &storagepb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
