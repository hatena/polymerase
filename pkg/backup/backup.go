package backup

import (
	pb "github.com/taku-k/xtralab/pkg/backup/proto"
	"io"
	"fmt"
)

type BackupTransferService struct {
	manager *TempBackupManager
}

func NewBackupTransferService(m *TempBackupManager) *BackupTransferService {
	return &BackupTransferService{
		manager: m,
	}
}

func (s *BackupTransferService) TransferFullBackup(stream pb.BackupTransferService_TransferFullBackupServer) error {
	var state *TempBackupState
	for {
		content, err := stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&pb.BackupReply{
				Message: "success",
			})
		}
		if err != nil {
			return err
		}
		if state == nil {
			state, err = s.manager.OpenFullBackup(content.Db)
			fmt.Printf("DB: %v\n", content.Db)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Received bytes: %v\n", len(content.Content))
		if err := state.Append(content.Content); err != nil {
			return err
		}
	}
}

func (s *BackupTransferService) TransferIncBackup(stream pb.BackupTransferService_TransferIncBackupServer) error {
	var state *TempBackupState
	var content *pb.IncBackupContentStream
	var err error
	for {
		content, err = stream.Recv()
		if err == io.EOF {
			if err := state.Close(); err != nil {
				return err
			}
			return stream.SendAndClose(&pb.BackupReply{
				Message: "success",
			})
		}
		if err != nil {
			return err
		}
		if state == nil {
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