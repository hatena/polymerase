package allocator

import (
	"github.com/pkg/errors"

	"github.com/taku-k/polymerase/pkg/etcd"
	"github.com/taku-k/polymerase/pkg/keys"
	"github.com/taku-k/polymerase/pkg/polypb"
)

func SearchStoredAddr(cli etcd.ClientAPI, db string) (string, error) {
	metas, err := cli.GetBackupMeta(keys.MakeDBBackupMetaPrefixKey(polypb.DatabaseID(db)))
	if err != nil {
		return "", err
	}
	if len(metas) == 0 {
		return "", errors.New("DB is not found.")
	}
	return metas[0].Host, nil
}
