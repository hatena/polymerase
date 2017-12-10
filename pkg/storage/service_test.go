package storage

//
//import (
//	"context"
//	"testing"
//	"time"
//
//	"github.com/coreos/etcd/integration"
//	"github.com/golang/protobuf/ptypes/timestamp"
//
//	"github.com/taku-k/polymerase/pkg/base"
//	"github.com/taku-k/polymerase/pkg/etcd"
//	"github.com/taku-k/polymerase/pkg/polypb"
//	"github.com/taku-k/polymerase/pkg/storage/storagepb"
//)
//
//type testContext struct {
//	t       *testing.T
//	tempdir string
//	cluster *integration.ClusterV3
//	cli     etcd.ClientAPI
//	storage *StorageService
//}
//
//func (tc *testContext) Start(t *testing.T) {
//	var err error
//
//	tc.t = t
//
//	tc.cluster = integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
//
//	tc.cli, err = etcd.NewTestClient(tc.cluster.RandClient())
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	if err != nil {
//		t.Fatal(err)
//	}
//	tc.storage = &StorageService{
//		EtcdCli: tc.cli,
//		cfg:     base.MakeServerConfig(),
//	}
//}
//
//func (tc *testContext) Close() {
//	tc.cluster.Terminate(tc.t)
//}
//
//func TestGetLatestToLSN(t *testing.T) {
//	tc := testContext{}
//	tc.Start(t)
//	defer tc.Close()
//
//	testCases := []struct {
//		db   string
//		data []struct {
//			key  string
//			info *polypb.BackupInfo
//		}
//		expect string
//	}{
//		{
//			db: "test-1",
//			data: []struct {
//				key  string
//				info *polypb.BackupInfo
//			}{
//				{
//					key: base.BackupBaseDBKey("test-1", time.Now().Format(base.DefaultTimeFormat)),
//					info: &polypb.BackupInfo{
//						FullBackup: &polypb.BackupMetadata{
//							StoredTime: &timestamp.Timestamp{
//								Seconds: 1,
//							},
//							ToLsn: "100",
//						},
//					},
//				},
//			},
//			expect: "100",
//		},
//		{
//			db: "test-2",
//			data: []struct {
//				key  string
//				info *polypb.BackupInfo
//			}{
//				{
//					key: base.BackupBaseDBKey("test-2", time.Now().Format(base.DefaultTimeFormat)),
//					info: &polypb.BackupInfo{
//						FullBackup: &polypb.BackupMetadata{
//							ToLsn: "100",
//						},
//						IncBackups: []*polypb.BackupMetadata{
//							{
//								ToLsn: "102",
//							},
//							{
//								ToLsn: "103",
//							},
//						},
//					},
//				},
//			},
//			expect: "103",
//		},
//	}
//
//	for i, ca := range testCases {
//		for _, d := range ca.data {
//			err := etcd.StoreBackupInfo(tc.cli, d.key, d.info)
//			if err != nil {
//				t.Fatalf("#%d: got error %v; want success", i, err)
//			}
//		}
//		res, err := tc.storage.GetLatestToLSN(context.Background(), &storagepb.GetLatestToLSNRequest{
//			Db: ca.db,
//		})
//		if err != nil {
//			t.Fatalf("#%d: got error %v; want success", i, err)
//		}
//		if res.Lsn != ca.expect {
//			t.Fatalf("#%d: got wrong lsn(%s); want %s", i, res.Lsn, ca.expect)
//		}
//	}
//}
