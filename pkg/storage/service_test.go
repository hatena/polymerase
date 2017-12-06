package storage

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/taku-k/polymerase/pkg/base"
	"github.com/taku-k/polymerase/pkg/status"
	"github.com/taku-k/polymerase/pkg/status/statuspb"
	"github.com/taku-k/polymerase/pkg/storage/storagepb"
	"github.com/taku-k/polymerase/pkg/utils/etcd"
)

type testContext struct {
	tempdir string
	server  *etcd.EtcdServer
	cli     *clientv3.Client
	storage *StorageService
}

func (tc *testContext) Start(t *testing.T) {
	var err error
	tc.tempdir, err = ioutil.TempDir(os.TempDir(), "polymerase_test")
	if err != nil {
		t.Fatal(err)
	}
	cfg := embed.NewConfig()
	cfg.Dir = tc.tempdir
	tc.server, err = etcd.NewEtcdServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tc.cli, err = clientv3.New(clientv3.Config{
		Endpoints: []string{cfg.ACUrls[0].Host},
	})
	if err != nil {
		t.Fatal(err)
	}
	tc.storage = &StorageService{
		EtcdCli: tc.cli,
		cfg:     base.MakeServerConfig(),
	}
}

func (tc *testContext) Close() {
	tc.server.Close()
	os.RemoveAll(tc.tempdir)
}

func TestGetLatestToLSN(t *testing.T) {
	tc := testContext{}
	tc.Start(t)
	defer tc.Close()

	testCases := []struct {
		db   string
		data []struct {
			key  string
			info *statuspb.BackupInfo
		}
		expect string
	}{
		{
			db: "test-1",
			data: []struct {
				key  string
				info *statuspb.BackupInfo
			}{
				{
					key: base.BackupBaseDBKey("test-1", time.Now().Format(base.DefaultTimeFormat)),
					info: &statuspb.BackupInfo{
						FullBackup: &statuspb.BackupMetadata{
							StoredTime: &timestamp.Timestamp{
								Seconds: 1,
							},
							ToLsn: "100",
						},
					},
				},
			},
			expect: "100",
		},
		{
			db: "test-2",
			data: []struct {
				key  string
				info *statuspb.BackupInfo
			}{
				{
					key: base.BackupBaseDBKey("test-2", time.Now().Format(base.DefaultTimeFormat)),
					info: &statuspb.BackupInfo{
						FullBackup: &statuspb.BackupMetadata{
							ToLsn: "100",
						},
						IncBackups: []*statuspb.BackupMetadata{
							{
								ToLsn: "102",
							},
							{
								ToLsn: "103",
							},
						},
					},
				},
			},
			expect: "103",
		},
	}

	for i, ca := range testCases {
		for _, d := range ca.data {
			err := status.StoreBackupInfo(tc.cli, d.key, d.info)
			if err != nil {
				t.Fatalf("#%d: got error %v; want success", i, err)
			}
		}
		res, err := tc.storage.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{
			Db: ca.db,
		})
		if err != nil {
			t.Fatalf("#%d: got error %v; want success", i, err)
		}
		if res.Lsn != ca.expect {
			t.Fatalf("#%d: got wrong lsn(%s); want %s", i, res.Lsn, ca.expect)
		}
	}
}
