package allocator

import (
	"github.com/pkg/errors"

	"github.com/hatena/polymerase/pkg/etcd"
	"github.com/hatena/polymerase/pkg/keys"
	"github.com/hatena/polymerase/pkg/polypb"
)

func SearchStoredAddr(cli etcd.ClientAPI, db polypb.DatabaseID) (string, error) {
	metas, err := cli.GetBackupMeta(keys.MakeDBBackupMetaPrefix(db))
	if err != nil {
		return "", err
	}
	if len(metas) == 0 {
		return "", errors.New("DB is not found.")
	}
	return metas[0].Host, nil
}
