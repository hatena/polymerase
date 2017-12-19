package storage

import (
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils"
)

type BackupManager struct {
	EtcdCli etcd.ClientAPI
	storage PhysicalStorage
	cfg     *base.ServerConfig
}

func NewBackupManager(cfg *base.ServerConfig) *BackupManager {
	return &BackupManager{
		cfg: cfg,
		storage: &DiskStorage{
			backupsDir: cfg.BackupsDir(),
		},
	}
}

func (m *BackupManager) GetLatestToLSN(db polypb.DatabaseID) (string, error) {
	metas, err := m.EtcdCli.GetBackupMeta(keys.MakeDBBackupMetaPrefix(db))
	if err != nil {
		return "", err
	}
	if len(metas) == 0 {
		return "", errors.New("not found any backups")
	}
	metas.Sort()
	meta := metas[len(metas)-1]
	details := meta.GetXtrabackup()
	if details == nil {
		return "", errors.Errorf("db %s is not Xtrabackup", db)
	}
	return details.ToLsn, nil
}

// SearchBaseTimePointByLSN finds base time point matching with a given lsn.
func (m *BackupManager) SearchBaseTimePointByLSN(db polypb.DatabaseID, lsn string) (polypb.TimePoint, error) {
	dbPrefix := keys.MakeDBBackupMetaPrefix(db)
	metas, err := m.EtcdCli.GetBackupMeta(dbPrefix)
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, errors.New("not found any full backups")
	}
	metas.Sort()
	for i := len(metas) - 1; i >= 0; i-- {
		m := metas[i]
		details := m.GetXtrabackup()
		if details == nil {
			return nil, errors.Errorf("db %s id not Xtrabackup", db)
		}
		if details.ToLsn == lsn {
			return m.BaseTimePoint, nil
		}
	}
	return nil, errors.New("backup matching with a given LSN is not found")
}

func (m *BackupManager) TransferTempBackup(tempDir string, key polypb.Key) error {
	return m.storage.Move(tempDir, key)
}

// SearchConsecutiveIncBackups
func (m *BackupManager) SearchConsecutiveIncBackups(
	db polypb.DatabaseID, from time.Time,
) ([]*storagepb.BackupFileInfo, error) {
	files := make([]*storagepb.BackupFileInfo, 0)
	metas, err := m.EtcdCli.GetBackupMeta(keys.MakeDBBackupMetaPrefix(db))
	if err != nil {
		return nil, err
	}
	if len(metas) == 0 {
		return nil, errors.New("not found any full backups")
	}
	metas.Sort()
	for i := len(metas) - 1; i >= 0; i-- {
		mi := metas[i]
		if (*mi.StoredTime).Before(from) {
			for j := i; j >= 0; j-- {
				mj := metas[j]
				files = append(files, &storagepb.BackupFileInfo{
					StorageType: m.storage.StorageType(),
					BackupType:  mj.BackupType,
					Key:         mj.Key,
					FileSize:    mj.FileSize,
				})
				if mj.BackupType == polypb.BackupType_XTRABACKUP_FULL {
					return files, nil
				}
			}
		}
	}
	return nil, errors.New("all backups are after a given time")
}

// GetFileStream returns a stream.
func (m *BackupManager) GetFileStream(key polypb.Key) (io.Reader, error) {
	metas, err := m.EtcdCli.GetBackupMeta(keys.MakeBackupMetaKeyFromKey(key))
	if err != nil {
		return nil, err
	}
	if len(metas) != 1 {
		return nil, errors.New("too many metadata to be fetched")
	}
	meta := metas[0]
	var r io.Reader
	switch meta.BackupType {
	case polypb.BackupType_XTRABACKUP_FULL:
		r, err = m.storage.FullBackupStream(meta.Key)
	case polypb.BackupType_XTRABACKUP_INC:
		r, err = m.storage.IncBackupStream(meta.Key)
	case polypb.BackupType_MYSQLDUMP:
		panic("implement me!")
	default:
		err = errors.New("not found such a backup type")
	}
	if err != nil {
		return nil, err
	}
	return r, nil
}

// PostFile creates a file.
func (m *BackupManager) PostFile(key polypb.Key, name string, r io.Reader) error {
	w, err := m.storage.CreateBackup(key, name)
	if err != nil {
		return err
	}
	defer w.Close()
	chunk := make([]byte, 1<<20)
	for {
		n, err := r.Read(chunk)
		if err == io.EOF {
			w.Write(chunk[:n])
			break
		}
		if err != nil {
			return err
		}
		w.Write(chunk[:n])
	}
	return nil
}

// RemoveBackups removes backups.
func (m *BackupManager) RemoveBackups(key polypb.Key) error {
	err := m.EtcdCli.RemoveBackupMeta(keys.MakeBackupMetaKeyFromKey(key))
	if err != nil {
		return err
	}
	return m.storage.DeleteBackup(key)
}

// GetKPastBackupKey returns a key.
func (m *BackupManager) GetKPastBackupKey(db polypb.DatabaseID, k int) (polypb.Key, error) {
	if k <= 0 {
		return nil, errors.Errorf("negative number %d is invalid", k)
	}
	metas, err := m.EtcdCli.GetBackupMeta(keys.MakeDBBackupMetaPrefix(db))
	if err != nil {
		return nil, err
	}
	fulls := make(polypb.BackupMetaSlice, 0)
	for _, meta := range metas {
		if meta.BackupType == polypb.BackupType_XTRABACKUP_FULL {
			fulls = append(fulls, meta)
		}
	}
	if len(fulls) < k {
		return nil, errors.New("not enough full backups to be removed")
	}
	fulls.Sort()
	return keys.MakeBackupPrefix(db, fulls[len(fulls)-k].BaseTimePoint), nil
}

func (m *BackupManager) RestoreBackupInfo() error {
	return m.storage.Walk(func(path string, info os.FileInfo, err error) error {
		if !isBackupedFile(path) {
			return nil
		}

		db, baseTP, backupTP, err := parseBackupPath(path)
		if err != nil {
			return err
		}
		key := keys.MakeBackupKey(db, baseTP, backupTP)
		storedTime, err := time.Parse(utils.TimeFormat, string(backupTP))
		if err != nil {
			return err
		}
		meta := &polypb.BackupMeta{
			StoredTime:    &storedTime,
			Host:          m.cfg.AdvertiseAddr,
			NodeId:        m.cfg.NodeID,
			Db:            db,
			FileSize:      info.Size(),
			Key:           key,
			BaseTimePoint: baseTP,
		}

		if err := m.restoreBackupMeta(path, meta); err != nil {
			return err
		}

		if err := m.EtcdCli.PutBackupMeta(
			keys.MakeBackupMetaKeyFromKey(key), meta); err != nil {
			return err
		}
		log.Printf("Restore %s backup: %s", meta.BackupType, path)
		return nil
	})
}

func parseBackupPath(
	path string,
) (polypb.DatabaseID, polypb.TimePoint, polypb.TimePoint, error) {
	sp := strings.Split(path, "/")
	if len(sp) < 4 {
		return nil, nil, nil, errors.New("not supported path")
	}
	db := polypb.DatabaseID(sp[len(sp)-4])
	base := polypb.TimePoint(sp[len(sp)-3])
	backup := polypb.TimePoint(sp[len(sp)-2])
	return db, base, backup, nil
}

func isBackupedFile(path string) bool {
	return strings.HasSuffix(path, "base.tar.gz") ||
		strings.HasSuffix(path, "inc.xb.gz") ||
		strings.HasSuffix(path, "dump.sql")
}

func (m *BackupManager) restoreBackupMeta(path string, meta *polypb.BackupMeta) error {
	if strings.HasSuffix(path, "base.tar.gz") {
		meta.BackupType = polypb.BackupType_XTRABACKUP_FULL
		return m.restoreXtrabackupMeta(meta)
	} else if strings.HasSuffix(path, "inc.xb.gz") {
		meta.BackupType = polypb.BackupType_XTRABACKUP_INC
		return m.restoreXtrabackupMeta(meta)
	} else if strings.HasSuffix(path, "dump.sql") {
		meta.BackupType = polypb.BackupType_MYSQLDUMP
		meta.Details = &polypb.BackupMeta_Mysqldump{
			Mysqldump: &polypb.MysqldumpMeta{},
		}
		return nil
	}
	return nil
}

func (m *BackupManager) restoreXtrabackupMeta(meta *polypb.BackupMeta) error {
	cp := m.storage.LoadXtrabackupCP(meta.Key)
	if cp.ToLSN == "" {
		return errors.New("xtrabackup_checkpoints file is not found")
	}
	meta.Details = &polypb.BackupMeta_Xtrabackup{
		Xtrabackup: &polypb.XtrabackupMeta{
			ToLsn: cp.ToLSN,
		},
	}
	return nil
}
