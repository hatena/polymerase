package storage

import (
	"io"
	"log"

	"github.com/looplab/fsm"
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
	var tempBackup *tempBackup
	var err error

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	backupFsm := fsm.NewFSM(
		"uninitialized",
		fsm.Events{
			{Name: "receive", Src: []string{"uninitialized", "receiving"}, Dst: "receiving"},
			{Name: "post_checkpoint", Src: []string{"receiving"}, Dst: "finished"},
			{Name: "client_error", Src: []string{"uninitialized", "receiving", "finished"}, Dst: "error"},
		},
		fsm.Callbacks{
			"leave_uninitialized": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.BackupRequest)
				if req.Db == nil {
					e.Cancel(errors.New("empty db is not acceptable"))
					return
				}
				tempBackup, err = s.tempMngr.openTempBackup(req.Db, &xtrabackupFullRequest{})
				if err != nil {
					e.Cancel(err)
					return
				}
				log.Printf("Start full-backup: db=%s\n", req.Db)
			},
			"enter_receiving": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.BackupRequest)
				if err := tempBackup.Append(req.Content); err != nil {
					e.Cancel(err)
				}
			},
			"after_post_checkpoint": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.CheckpointRequest)
				cp, err := polypb.LoadXtrabackupCP(req.Body)
				if err != nil {
					e.Cancel(err)
					return
				}
				backupMeta := tempBackup.meta.GetXtrabackup()
				if backupMeta == nil {
					e.Cancel(errors.Errorf("XtrabackupMeta is not initialized"))
					return
				}
				backupMeta.Checkpoints = cp
			},
			"enter_error": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.ClientErrorRequest)
				log.Print(req.Message)
				if tempBackup != nil {
					if err := tempBackup.remove(); err != nil {
						e.Cancel(err)
					}
				}
			},
		},
	)

	for {
		request, err := stream.Recv()
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
		if request.GetBackupRequest() != nil {
			if err := backupFsm.Event("receive", request.GetBackupRequest()); err != nil {
				switch err.(type) {
				case fsm.NoTransitionError:
				default:
					return err
				}
			}
		}
		if request.GetCheckpointRequest() != nil {
			if err := backupFsm.Event(
				"post_checkpoint", request.GetCheckpointRequest()); err != nil {
				return err
			}
		}
		if request.GetClientErrorRequest() != nil {
			if err := backupFsm.Event(
				"client_error", request.GetClientErrorRequest()); err != nil {
				return err
			}
		}
	}
}

func (s *Service) TransferIncBackup(
	stream storagepb.StorageService_TransferIncBackupServer,
) error {
	var tempBackup *tempBackup
	var err error

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	backupFsm := fsm.NewFSM(
		"uninitialized",
		fsm.Events{
			{Name: "receive", Src: []string{"uninitialized", "receiving"}, Dst: "receiving"},
			{Name: "post_checkpoint", Src: []string{"receiving"}, Dst: "finished"},
			{Name: "client_error", Src: []string{"uninitialized", "receiving", "finished"}, Dst: "error"},
		},
		fsm.Callbacks{
			"leave_uninitialized": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.BackupRequest)
				if req.Db == nil {
					e.Cancel(errors.New("empty db is not acceptable"))
					return
				}
				if req.Lsn == "" {
					e.Cancel(errors.New("empty lsn is not acceptable"))
					return
				}
				tempBackup, err = s.tempMngr.openTempBackup(
					req.Db,
					&xtrabackupIncRequest{
						LSN: req.Lsn,
					})
				if err != nil {
					e.Cancel(err)
					return
				}
				log.Printf("Start inc-backup: db=%s\n", req.Db)
			},
			"enter_receiving": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.BackupRequest)
				if err := tempBackup.Append(req.Content); err != nil {
					e.Cancel(err)
				}
			},
			"after_post_checkpoint": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.CheckpointRequest)
				cp, err := polypb.LoadXtrabackupCP(req.Body)
				if err != nil {
					e.Cancel(err)
					return
				}
				backupMeta := tempBackup.meta.GetXtrabackup()
				if backupMeta == nil {
					e.Cancel(errors.Errorf("XtrabackupMeta is not initialized"))
					return
				}
				backupMeta.Checkpoints = cp
			},
			"enter_error": func(e *fsm.Event) {
				req := e.Args[0].(*storagepb.ClientErrorRequest)
				log.Print(req.Message)
				if tempBackup != nil {
					if err := tempBackup.remove(); err != nil {
						e.Cancel(err)
					}
				}
			},
		},
	)

	for {
		request, err := stream.Recv()
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
		if request.GetBackupRequest() != nil {
			if err := backupFsm.Event("receive", request.GetBackupRequest()); err != nil {
				switch err.(type) {
				case fsm.NoTransitionError:
				default:
					return err
				}
			}
		}
		if request.GetCheckpointRequest() != nil {
			if err := backupFsm.Event(
				"post_checkpoint", request.GetCheckpointRequest()); err != nil {
				return err
			}
		}
		if request.GetClientErrorRequest() != nil {
			if err := backupFsm.Event(
				"client_error", request.GetClientErrorRequest()); err != nil {
				return err
			}
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
