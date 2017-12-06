package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"sort"
	"time"

	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
)

type StorageService struct {
	storage    BackupStorage
	rateLimit  float64
	EtcdCli    *clientv3.Client
	tempMngr   *TempBackupManager
	aggregator *status.WeeklyBackupAggregator
	cfg        *base.ServerConfig
}

func NewStorageService(
	storage BackupStorage,
	rateLimit uint64,
	tempMngr *TempBackupManager,
	aggregator *status.WeeklyBackupAggregator,
	cfg *base.ServerConfig,
) *StorageService {
	return &StorageService{
		storage:    storage,
		rateLimit:  float64(rateLimit),
		tempMngr:   tempMngr,
		aggregator: aggregator,
		cfg:        cfg,
	}
}

func (s *StorageService) GetLatestToLSN(
	ctx context.Context, req *storagepb.GetLatestToLSNRequest,
) (*storagepb.GetLatestToLSNResponse, error) {
	backups := status.GetBackupsInfo(s.EtcdCli, base.BackupDBKey(req.Db))
	if backups == nil {
		log.Printf("Not found db=%s\n", req.Db)
		return nil, errors.New("not found such a db")
	}
	sortedKeys := make([]*base.BackupKeyItem, len(backups))
	var i int
	for k := range backups {
		sortedKeys[i] = base.ParseBackupKey(k, s.cfg.TimeFormat)
		if sortedKeys[i] == nil {
			return nil, errors.New(fmt.Sprintf("found invalid backup key=%s", k))
		}
		i++
	}
	sort.Slice(sortedKeys, func(i, j int) bool {
		return sortedKeys[j].StoredTime.After(sortedKeys[i].StoredTime)
	})
	key := base.BackupBaseDBKey(req.Db, sortedKeys[0].StoredTime.Format(s.cfg.TimeFormat))
	lsn := backups[key].FullBackup.ToLsn
	if len(backups[key].IncBackups) != 0 {
		incs := backups[key].IncBackups
		lsn = incs[len(incs)-1].ToLsn
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
	key, err := s.storage.GetKPastBackupKey(req.Db, 2)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{
			Message: "There is no backup to purge.",
		}, nil
	}
	log.Printf("Purge key=%s\n", key)
	err = s.storage.RemoveBackups(s.EtcdCli, key)
	if err != nil {
		return &storagepb.PurgePrevBackupResponse{}, err
	}
	return &storagepb.PurgePrevBackupResponse{
		Message: "Purge succeeds",
	}, nil
}

func (s *StorageService) GetBestStartTime(
	ctx context.Context, req *storagepb.GetBestStartTimeRequest,
) (*storagepb.GetBestStartTimeResponse, error) {
	m, h := s.aggregator.BestStartTime(time.Weekday(0))
	return &storagepb.GetBestStartTimeResponse{
		Minute: int32(m),
		Hour:   int32(h),
	}, nil
}

func (s *StorageService) TransferFullBackup(
	stream storagepb.StorageService_TransferFullBackupServer,
) error {
	var state *TempBackupState

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&storagepb.BackupReply{
				Message: "success",
				Key:     state.key,
			})
		}
		if err != nil {
			return err
		}
		if state == nil {
			if content.Db == "" {
				return errors.New("empty db is not acceptable")
			}
			state, err = s.tempMngr.OpenFullBackup(content.Db)
			if err != nil {
				return err
			}
			log.Printf("Start full-backup: db=%s, temp_path=%s\n", content.Db, state.tempDir)
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *StorageService) TransferIncBackup(
	stream storagepb.StorageService_TransferIncBackupServer,
) error {
	var state *TempBackupState

	if p, ok := peer.FromContext(stream.Context()); ok {
		log.Printf("Established peer: %v\n", p.Addr)
	}

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&storagepb.BackupReply{
				Message: "success",
				Key:     state.key,
			})
		}
		if err != nil {
			return err
		}
		if state == nil {
			if content.Db == "" {
				return errors.New("empty db is not acceptable")
			}
			if content.Lsn == "" {
				return errors.New("empty lsn is not acceptable")
			}
			state, err = s.tempMngr.OpenIncBackup(content.Db, content.Lsn)
			if err != nil {
				return err
			}
			log.Printf("Start inc-backup: db=%s, temp_path=%s\n", content.Db, state.tempDir)
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *StorageService) PostCheckpoints(
	ctx context.Context,
	req *storagepb.PostCheckpointsRequest,
) (*storagepb.PostCheckpointsResponse, error) {
	r := bytes.NewReader(req.Content)
	if err := s.tempMngr.storage.PostFile(req.Key, "xtrabackup_checkpoints", r); err != nil {
		return &storagepb.PostCheckpointsResponse{}, err
	}
	cp := base.LoadXtrabackupCP(req.Content)
	if cp.ToLSN == "" {
		return nil, errors.New("failed to load")
	}
	err := status.UpdateCheckpoint(s.EtcdCli, req.Key, cp.ToLSN)
	if err != nil {
		return nil, err
	}
	return &storagepb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
