package tempbackup

import (
	"io"

	"bytes"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/taku-k/polymerase/pkg/tempbackup/tempbackuppb"
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

func (s *TempBackupTransferService) TransferFullBackup(stream tempbackuppb.BackupTransferService_TransferFullBackupServer) error {
	var state *TempBackupState
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
			log.WithField("db", content.Db).Info("Connected from db")
			log.WithField("path", state.tempDir).Info("Create temporary directory")
			if err != nil {
				return err
			}
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *TempBackupTransferService) TransferIncBackup(stream tempbackuppb.BackupTransferService_TransferIncBackupServer) error {
	var state *TempBackupState
	var content *tempbackuppb.IncBackupContentStream
	var err error
	for {
		content, err = stream.Recv()
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
		}
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *TempBackupTransferService) PostCheckpoints(
	ctx context.Context,
	req *tempbackuppb.PostCheckpointsRequest) (*tempbackuppb.PostCheckpointsResponse, error) {
	r := bytes.NewReader(req.Content)
	if err := s.manager.storage.PostFile(req.Key, "xtrabackup_checkpoints", r); err != nil {
		return &tempbackuppb.PostCheckpointsResponse{}, err
	}
	return &tempbackuppb.PostCheckpointsResponse{
		Message: "success",
	}, nil
}
