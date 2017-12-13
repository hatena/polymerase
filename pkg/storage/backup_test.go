package storage

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils"
	"github.com/taku-k/polymerase/pkg/utils/testutil"
)

func toPtr(s time.Time) *time.Time {
	return &s
}

type fakeEtcdCli struct {
	etcd.ClientAPI
	FakeGetBackupMeta func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)
	FakePutBackupMeta func(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error

	ts []time.Time
}

func (c *fakeEtcdCli) GetBackupMeta(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
	return c.FakeGetBackupMeta(key)
}

func (c *fakeEtcdCli) PutBackupMeta(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error {
	return c.FakePutBackupMeta(key, meta)
}

func (c *fakeEtcdCli) tpAt(i int) polypb.TimePoint {
	return polypb.NewTimePoint(c.ts[i])
}

func (c *fakeEtcdCli) tAt(i int) time.Time {
	return c.ts[i]
}

func newFakeClient(t time.Time) *fakeEtcdCli {
	c := &fakeEtcdCli{}
	c.ts = make([]time.Time, 7)
	for i := 0; i < 6; i++ {
		c.ts[i] = t.Add(time.Duration(i-6) * time.Hour)
	}
	/*
	 Time order: t0 < t1 < t2 < t3 < t4 < t5
	 db
	 ├── t0
	 │   ├── t0 (FULL)
	 │   ├── t1 (INC)
	 │   ├── t2 (INC)
	 │   └── t4 (INC)
	 └── t3
	     ├── t3 (FULL)
	     └── t5 (INC)
	*/
	c.FakeGetBackupMeta = func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
		db, _, _, _ := keys.DecodeMetaKey(key)
		if !bytes.Equal(db, []byte("db")) {
			return make(polypb.BackupMetaSlice, 0), nil
		}
		return []*polypb.BackupMeta{
			{
				StoredTime:    toPtr(c.tAt(0)),
				BaseTimePoint: c.tpAt(0),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "10",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_FULL,
				Key:        keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(0)),
			},
			{
				StoredTime:    toPtr(c.tAt(1)),
				BaseTimePoint: c.tpAt(0),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "20",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_INC,
				Key:        keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(1)),
			},
			{
				StoredTime:    toPtr(c.tAt(2)),
				BaseTimePoint: c.tpAt(0),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "30",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_INC,
				Key:        keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(2)),
			},
			{
				StoredTime:    toPtr(c.tAt(4)),
				BaseTimePoint: c.tpAt(0),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "110",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_INC,
				Key:        keys.MakeBackupKey(db, c.tpAt(0), c.tpAt(4)),
			},
			{
				StoredTime:    toPtr(c.tAt(3)),
				BaseTimePoint: c.tpAt(3),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "100",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_FULL,
				Key:        keys.MakeBackupKey(db, c.tpAt(3), c.tpAt(3)),
			},
			{
				StoredTime:    toPtr(c.tAt(5)),
				BaseTimePoint: c.tpAt(3),
				Details: &polypb.BackupMeta_Xtrabackup{
					Xtrabackup: &polypb.XtrabackupMeta{
						ToLsn: "110",
					},
				},
				BackupType: polypb.BackupType_XTRABACKUP_INC,
				Key:        keys.MakeBackupKey(db, c.tpAt(3), c.tpAt(5)),
			},
		}, nil
	}
	return c
}

func TestBackupManager_GetLatestToLSN(t *testing.T) {
	cli := newFakeClient(time.Now())
	mngr := &BackupManager{
		EtcdCli: cli,
	}

	testCases := []struct {
		db       polypb.DatabaseID
		expected string
		errStr   string
	}{
		{
			db:       polypb.DatabaseID("db"),
			expected: "110",
		},
		{
			db:     polypb.DatabaseID("db-nothing"),
			errStr: "not found any backups",
		},
	}

	for i, tc := range testCases {
		lsn, err := mngr.GetLatestToLSN(tc.db)
		if tc.errStr == "" {
			if err != nil {
				t.Errorf("#%d: GetLatestToLSN(%q): got error %q; want success",
					i, tc.db, err)
			}
			if lsn != tc.expected {
				t.Errorf("#%d: GetLatestToLSN(%q): got wrong lsn %s; want %s",
					i, tc.db, lsn, tc.expected)
			}
		} else {
			if !testutil.IsError(err, tc.errStr) {
				t.Errorf("#%d: GetLatestToLSN(%q): got wrong error %q; want %q",
					i, tc.db, err, tc.errStr)
			}
		}
	}

}

func TestBackupManager_SearchBaseTimePointByLSN(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")
	fakeClient := newFakeClient(tn)

	testCases := []struct {
		db       polypb.DatabaseID
		lsn      string
		expected polypb.TimePoint
	}{
		{
			db:       db,
			lsn:      "30",
			expected: fakeClient.tpAt(0),
		},
		{
			db:       db,
			lsn:      "100",
			expected: fakeClient.tpAt(3),
		},
		{
			db:       db,
			lsn:      "110",
			expected: fakeClient.tpAt(3),
		},
	}

	mngr := &BackupManager{
		EtcdCli: fakeClient,
	}

	for i, tc := range testCases {
		tp, err := mngr.SearchBaseTimePointByLSN(tc.db, tc.lsn)
		if err != nil {
			t.Errorf("#%d: got error %q; want success", i, err)
		}
		if !tp.Equal(tc.expected) {
			t.Errorf("#%d: got wrong timepoint %q; want timepoint %q",
				i, tp, tc.expected)
		}
	}
}

func TestBackupManager_SearchConsecutiveIncBackups(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")
	cli := newFakeClient(tn)

	testCases := []struct {
		from     time.Time
		expected []*storagepb.BackupFileInfo
	}{
		{
			from: time.Now(),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(3), cli.tpAt(5)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(3), cli.tpAt(3)),
				},
			},
		},
		{
			from: cli.tAt(2).Add(time.Minute),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(2)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_INC,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(1)),
				},
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(0)),
				},
			},
		},
		{
			from: cli.tAt(0).Add(time.Minute),
			expected: []*storagepb.BackupFileInfo{
				{
					StorageType: polypb.StorageType_LOCAL_MEM,
					BackupType:  polypb.BackupType_XTRABACKUP_FULL,
					Key:         keys.MakeBackupKey(db, cli.tpAt(0), cli.tpAt(0)),
				},
			},
		},
	}

	mngr := &BackupManager{
		EtcdCli: cli,
		storage: &MemStorage{},
	}
	for i, tc := range testCases {
		res, err := mngr.SearchConsecutiveIncBackups(db, tc.from)
		if err != nil {
			t.Errorf("#%d: got error %q; want success", i, err)
		}
		if diff := pretty.Compare(tc.expected, res); diff != "" {
			t.Errorf("#%d: got wrong BackupFileInfo\n%s",
				i, diff)
		}
		//if !reflect.DeepEqual(res, tc.expected) {
		//	t.Errorf("#%d: got wrong BackupFileInfo %q; want %q",
		//		i, res, tc.expected)
		//}
	}
}

func TestBackupManager_GetFileStream(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")

	testCases := []struct {
		getBackupMeta func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error)
		storage       *fakePhysicalStorage
		expected      string
	}{
		{
			getBackupMeta: func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
				return []*polypb.BackupMeta{
					{
						BackupType: polypb.BackupType_XTRABACKUP_FULL,
					},
				}, nil
			},
			storage: &fakePhysicalStorage{
				FakeFullBackupStream: func(key polypb.Key) (io.Reader, error) {
					return bytes.NewBufferString("full"), nil
				},
			},
			expected: "full",
		},
		{
			getBackupMeta: func(key polypb.BackupMetaKey) (polypb.BackupMetaSlice, error) {
				return []*polypb.BackupMeta{
					{
						BackupType: polypb.BackupType_XTRABACKUP_INC,
					},
				}, nil
			},
			storage: &fakePhysicalStorage{
				FakeIncBackupStream: func(key polypb.Key) (io.Reader, error) {
					return bytes.NewBufferString("inc"), nil
				},
			},
			expected: "inc",
		},
	}

	for i, tc := range testCases {
		mngr := &BackupManager{
			EtcdCli: &fakeEtcdCli{
				FakeGetBackupMeta: tc.getBackupMeta,
			},
			storage: tc.storage,
		}

		stream, err := mngr.GetFileStream(keys.MakeBackupKey(db, polypb.NewTimePoint(tn), polypb.NewTimePoint(tn)))
		if err != nil {
			t.Errorf("#%d: Got error %q; want success", i, err)
		}
		buf, err := ioutil.ReadAll(stream)
		if err != nil {
			t.Errorf("#%d: ioutil.ReadAll returns error %q; want success", i, err)
		}
		if string(buf) != tc.expected {
			t.Errorf("#%d: Got wrong stream %q; want %s", i, buf, tc.expected)
		}
	}
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (cb *ClosingBuffer) Close() (err error) {
	return
}

func TestBackupManager_PostFile(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")

	buf := &bytes.Buffer{}
	storage := &fakePhysicalStorage{
		FakeCreateBackup: func(key polypb.Key, name string) (io.WriteCloser, error) {
			return &ClosingBuffer{buf}, nil
		},
	}

	mngr := &BackupManager{
		storage: storage,
	}
	input := "content"
	inbuf := bytes.NewBufferString(input)
	err := mngr.PostFile(keys.MakeBackupKey(db, polypb.NewTimePoint(tn), polypb.NewTimePoint(tn)), "", inbuf)
	if err != nil {
		t.Errorf("Got error %q; want success", err)
	}
	if buf.String() != input {
		t.Errorf("Got wrong content %q; want %q", buf.String(), input)
	}
}

func TestBackupManager_GetKPastBackupKey(t *testing.T) {
	tn := time.Now()
	db := polypb.DatabaseID("db")
	cli := newFakeClient(tn)

	mngr := &BackupManager{
		EtcdCli: cli,
	}

	testCases := []struct {
		past     int
		expected polypb.Key
		errStr   string
	}{
		{
			past:     1,
			expected: keys.MakeBackupPrefix(db, cli.tpAt(3)),
		},
		{
			past:     2,
			expected: keys.MakeBackupPrefix(db, cli.tpAt(0)),
		},
		{
			past:   0,
			errStr: "negative number 0 is invalid",
		},
		{
			past:   3,
			errStr: "not enough full backups to be removed",
		},
	}

	for i, tc := range testCases {
		key, err := mngr.GetKPastBackupKey(db, tc.past)
		if tc.errStr == "" {
			if err != nil {
				t.Errorf("#%d: got error %q; want success", i, err)
			}
			if !bytes.Equal(key, tc.expected) {
				t.Errorf("#%d: got wrong key %q; want key %q",
					i, key, tc.expected)
			}
		} else {
			if !testutil.IsError(err, tc.errStr) {
				t.Errorf("#%d: get wrong error %q; want error string %q",
					i, err, tc.errStr)
			}
		}
	}
}

type fakeFileInfo struct {
	os.FileInfo
	size int64
}

func (f fakeFileInfo) Size() int64 { return f.size }

func TestBackupManager_RestoreBackupInfo(t *testing.T) {
	tn, _ := time.Parse(utils.TimeFormat, "2017-12-13_17-22-44_+0000")
	db := polypb.DatabaseID("db")
	pt := polypb.NewTimePoint(tn)
	cfg := base.MakeServerConfig()

	golden := []struct {
		lsn     string
		key     polypb.Key
		genPath func(key polypb.Key) string
		info    os.FileInfo
		genMeta func(key polypb.Key, info os.FileInfo, lsn string) *polypb.BackupMeta
	}{
		{
			lsn: "100",
			key: keys.MakeBackupKey(db, pt, pt),
			genPath: func(key polypb.Key) string {
				return string(key) + "/base.tar.gz"
			},
			info: fakeFileInfo{size: 100},
			genMeta: func(key polypb.Key, info os.FileInfo, lsn string) *polypb.BackupMeta {
				return &polypb.BackupMeta{
					StoredTime: &tn,
					Host:       cfg.AdvertiseAddr,
					NodeId:     cfg.NodeID,
					BackupType: polypb.BackupType_XTRABACKUP_FULL,
					Db:         db,
					Details: &polypb.BackupMeta_Xtrabackup{
						Xtrabackup: &polypb.XtrabackupMeta{
							ToLsn: lsn,
						},
					},
					FileSize:      info.Size(),
					Key:           key,
					BaseTimePoint: pt,
				}
			},
		},
		{
			lsn: "110",
			key: keys.MakeBackupKey(db, pt, polypb.NewTimePoint(tn.Add(time.Hour))),
			genPath: func(key polypb.Key) string {
				return string(key) + "/inc.xb.gz"
			},
			info: fakeFileInfo{size: 110},
			genMeta: func(key polypb.Key, info os.FileInfo, lsn string) *polypb.BackupMeta {
				st := tn.Add(time.Hour)
				return &polypb.BackupMeta{
					StoredTime: &st,
					Host:       cfg.AdvertiseAddr,
					NodeId:     cfg.NodeID,
					BackupType: polypb.BackupType_XTRABACKUP_INC,
					Db:         db,
					Details: &polypb.BackupMeta_Xtrabackup{
						Xtrabackup: &polypb.XtrabackupMeta{
							ToLsn: lsn,
						},
					},
					FileSize:      info.Size(),
					Key:           key,
					BaseTimePoint: pt,
				}
			},
		},
	}

	walkIndex := 0
	storage := &fakePhysicalStorage{
		FakeLoadXtrabackupCP: func(key polypb.Key) base.XtrabackupCheckpoints {
			return base.XtrabackupCheckpoints{
				ToLSN: golden[walkIndex].lsn,
			}
		},
		FakeWalk: func(f func(path string, info os.FileInfo, err error) error) error {
			for ; walkIndex < len(golden); walkIndex++ {
				g := golden[walkIndex]
				err := f(g.genPath(g.key), g.info, nil)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}

	result := make([]*polypb.BackupMeta, 0)
	cli := &fakeEtcdCli{
		FakePutBackupMeta: func(key polypb.BackupMetaKey, meta *polypb.BackupMeta) error {
			result = append(result, meta)
			return nil
		},
	}

	mngr := &BackupManager{
		EtcdCli: cli,
		storage: storage,
		cfg:     cfg,
	}

	err := mngr.RestoreBackupInfo()
	if err != nil {
		t.Errorf("Got error %q; want success", err)
	}
	if len(result) != len(golden) {
		t.Errorf("Got the wrong number of metadata %d", len(result))
	}

	for i := 0; i < len(golden); i++ {
		g := golden[i]
		expected := g.genMeta(g.key, g.info, g.lsn)
		if diff := pretty.Compare(expected, result[i]); diff != "" {
			t.Errorf("#%d: got wrong metadata\n%s",
				i, diff)
		}
	}
}
