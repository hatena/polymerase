package storage

import (
	"bytes"
	"errors"
	"io"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type Service struct {
	manager   *BackupManager
	rateLimit float64
	EtcdCli   etcd.ClientAPI
	tempMngr  *TempBackupManager
	cfg       *base.ServerConfig
}

func NewService(
	manager *BackupManager,
	rateLimit uint64,
	tempMngr *TempBackupManager,
	cfg *base.ServerConfig,
) *Service {
	return &Service{
		manager:   manager,
		rateLimit: float64(rateLimit),
		tempMngr:  tempMngr,
		cfg:       cfg,
	}
}

func (s *Service) GetLatestToLSN(
	ctx context.Context, req *storagepb.GetLatestToLSNRequest,
) (*storagepb.GetLatestToLSNResponse, error) {
	db := polypb.DatabaseID(req.Db)
	lsn, err := s.manager.GetLatestToLSN(db)
	return &storagepb.GetLatestToLSNResponse{
		Lsn: lsn,
	}, err
}

func (s *Service) GetKeysAtPoint(
	ctx context.Context, req *storagepb.GetKeysAtPointRequest,
) (*storagepb.GetKeysAtPointResponse, error) {
	bfiles, err := s.manager.SearchConsecutiveIncBackups(req.Db, req.From)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal, "GetKeysAtPoint is failed: %s", err)
	}
	return &storagepb.GetKeysAtPointResponse{
		Keys: bfiles,
	}, nil
}

func (s *Service) GetFileByKey(
	req *storagepb.GetFileByKeyRequest, stream storagepb.StorageService_GetFileByKeyServer,
) error {
	r, err := s.manager.GetFileStream(req.Key)
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

func (s *Service) PurgePrevBackup(
	ctx context.Context, req *storagepb.PurgePrevBackupRequest,
) (*storagepb.PurgePrevBackupResponse, error) {
	key, err := s.manager.GetKPastBackupKey(req.Db, 2)
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

func (s *Service) TransferFullBackup(
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

func (s *Service) TransferIncBackup(
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

func (s *Service) TransferMysqldump(
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

func (s *Service) PostCheckpoints(
	ctx context.Context,
	req *storagepb.PostCheckpointsRequest,
) (*storagepb.PostCheckpointsResponse, error) {
	r := bytes.NewReader(req.Content)
	if err := s.manager.PostFile(req.Key, "xtrabackup_checkpoints", r); err != nil {
		return &storagepb.PostCheckpointsResponse{}, err
	}
	cp := base.LoadXtrabackupCP(req.Content)
	if cp.ToLSN == "" {
		return nil, errors.New("failed to load")
	}
	err := s.EtcdCli.UpdateLSN(keys.MakeBackupMetaKeyFromKey(req.Key), cp.ToLSN)
	if err != nil {
		return nil, err
	}
	return &storagepb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
