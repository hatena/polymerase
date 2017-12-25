package storage

import (
	"io"
	"log"

	"github.com/pkg/errors"
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
	cfg       *base.ServerConfig
}

func NewService(
	manager *BackupManager,
	rateLimit uint64,
	cfg *base.ServerConfig,
) *Service {
	return &Service{
		manager:   manager,
		rateLimit: float64(rateLimit),
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

func (s *Service) TransferBackup(
	stream storagepb.StorageService_TransferBackupServer,
) error {
	var backup *inBackup

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			meta, err := backup.close()
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
		if request.GetInitializeRequest() != nil {
			req := request.GetInitializeRequest()
			if req.Db == nil {
				return errors.New("empty db is not acceptable")
			}
			switch req.BackupType {
			case polypb.BackupType_XTRABACKUP_FULL:
				backup, err = s.manager.openBackup(req.Db, &xtrabackupFullRequest{})
			case polypb.BackupType_XTRABACKUP_INC:
				backup, err = s.manager.openBackup(req.Db, &xtrabackupIncRequest{
					LSN: req.Lsn,
				})
			case polypb.BackupType_MYSQLDUMP:
				backup, err = s.manager.openBackup(req.Db, &mysqldumpRequest{})
			default:
				return errors.Errorf("unknown backup type %s", req.BackupType)
			}
			if err != nil {
				return err
			}
			log.Printf("Start %s: db=%s\n", req.BackupType, req.Db)
		}
		if request.GetBackupRequest() != nil {
			req := request.GetBackupRequest()
			if err := backup.append(req.Content); err != nil {
				return err
			}
		}
		if request.GetCheckpointRequest() != nil {
			req := request.GetCheckpointRequest()
			cp, err := polypb.LoadXtrabackupCP(req.Body)
			if err != nil {
				return err
			}
			backupMeta := backup.meta.GetXtrabackupMeta()
			if backupMeta == nil {
				return errors.Errorf("XtrabackupMeta is not initialized")
			}
			backupMeta.Checkpoints = cp
		}
		if request.GetClientErrorRequest() != nil {
			req := request.GetClientErrorRequest()
			log.Print(req.Message)
			if backup != nil {
				if err := backup.remove(); err != nil {
					return err
				}
			}
		}
	}
}
