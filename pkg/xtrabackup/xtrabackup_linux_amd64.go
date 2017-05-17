package xtrabackup

/*
#cgo LDFLAGS: -L./build/percona-xtrabackup/storage/innobase/xtrabackup/src -lxtrabackup_static
#cgo LDFLAGS: -L./build/percona-xtrabackup/libmysqld -lmysqld
#cgo LDFLAGS: -L./build/percona-xtrabackup/storage/innobase/xtrabackup/src -lxtrabackup_static
#cgo LDFLAGS: -L./build/percona-xtrabackup/storage/innobase/xtrabackup/src/libarchive/libarchive -larchive
#cgo LDFLAGS: -L./build/percona-xtrabackup/storage/innobase/xtrabackup/src/crc -lcrc
#cgo LDFLAGS: -L/usr/lib -L/usr/local/lib -L/usr/lib/x86_64-linux-gnu -lgcrypt -lz
#cgo LDFLAGS: -lstdc++ -lm -lpthread -lrt -ldl -laio -lgpg-error -lcrypt
extern int xtrabackup_main(int argc, char **argv);
*/
import "C"

func ExecXtrabackupCmd(args []string) {
	var argc C.int
	var argv []*C.char

	argc = C.int(len(args))
	argv = make([]*C.char, len(args))
	for i, arg := range args {
		argv[i] = C.CString(arg)
	}

	C.xtrabackup_main(argc, (**C.char)(&argv[0]))
}
