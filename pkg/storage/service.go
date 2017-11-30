package storage

import (
	"bytes"
	"io"
	"log"
	"time"

	"fmt"

	"sort"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"golang.org/x/net/context"
	"google.golang.org/grpc/peer"
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
	lsn, err := s.storage.GetLatestToLSN(req.Db)
	if err != nil {
		log.Printf("Not found db=%s\n", req.Db)
		return &storagepb.GetLatestToLSNResponse{Lsn: ""}, errors.New("Not found such a db")
	}
	resp, err := s.EtcdCli.KV.Get(context.Background(), req.Db, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend), clientv3.WithKeysOnly())
	if err != nil {
		panic(err)
	}
	if len(resp.Kvs) == 0 {
		return nil, errors.New("not found any base backup")
	}
	sort.Slice(resp.Kvs, func(i, j int) bool {
		spi := strings.Split(string(resp.Kvs[i].Key), "/")
		spj := strings.Split(string(resp.Kvs[j].Key), "/")
		ti, err := time.Parse(s.cfg.TimeFormat, spi[1])
		if err != nil {
			panic(err)
		}
		tj, err := time.Parse(s.cfg.TimeFormat, spj[1])
		if err != nil {
			panic(err)
		}
		if ti.Equal(tj) {
			ti, err := time.Parse(s.cfg.TimeFormat, spi[2])
			if err != nil {
				panic(err)
			}
			tj, err := time.Parse(s.cfg.TimeFormat, spj[2])
			if err != nil {
				panic(err)
			}
			return tj.After(ti)
		}
		return tj.After(ti)
	})
	for _, kv := range resp.Kvs {
		fmt.Printf("%s\n", kv.Key)
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
	return &storagepb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
