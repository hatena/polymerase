package tempbackup

import (
	"bytes"
	"io"

	"github.com/pkg/errors"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
	"github.com/taku-k/polymerase/pkg/utils/log"
	"golang.org/x/net/context"
)

type TempBackupTransferService struct {
	manager *TempBackupManager
}

func NewBackupTransferService(m *TempBackupManager) *TempBackupTransferService {
	return &TempBackupTransferService{
		manager: m,
	}
}

func (s *TempBackupTransferService) TransferFullBackup(
	stream tempbackuppb.BackupTransferService_TransferFullBackupServer,
) error {
	var state *TempBackupState

	log.LogWithGRPC(stream.Context()).Info("Established peer")

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&tempbackuppb.BackupReply{
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
			state, err = s.manager.OpenFullBackup(content.Db)
			if err != nil {
				return err
			}
			log.WithField("db", content.Db).
				WithField("temp_path", state.tempDir).
				Info("Start full-backup")
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *TempBackupTransferService) TransferIncBackup(
	stream tempbackuppb.BackupTransferService_TransferIncBackupServer,
) error {
	var state *TempBackupState

	log.LogWithGRPC(stream.Context()).Info("Established peer")

	for {
		content, err := stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&tempbackuppb.BackupReply{
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
			state, err = s.manager.OpenIncBackup(content.Db, content.Lsn)
			if err != nil {
				return err
			}
			log.WithField("db", content.Db).
				WithField("temp_path", state.tempDir).
				Info("Start inc-backup")
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *TempBackupTransferService) PostCheckpoints(
	ctx context.Context,
	req *tempbackuppb.PostCheckpointsRequest,
) (*tempbackuppb.PostCheckpointsResponse, error) {
	r := bytes.NewReader(req.Content)
	if err := s.manager.storage.PostFile(req.Key, "xtrabackup_checkpoints", r); err != nil {
		return &tempbackuppb.PostCheckpointsResponse{}, err
	}
	return &tempbackuppb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
