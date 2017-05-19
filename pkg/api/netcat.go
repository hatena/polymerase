package api

import (
	"fmt"
	"io/ioutil"
	"net"
	"os/exec"

	"github.com/mattn/go-shellwords"
	"github.com/taku-k/xtralab/pkg/config"
)

type NCPool struct {
	maxConn  int
	dbToPort map[string]int
	tempDir  string
}

func NewNCPool(conf *config.Config) *NCPool {
	return &NCPool{
		maxConn:  10,
		dbToPort: make(map[string]int),
		tempDir:  conf.TempDir,
	}
}

func (p *NCPool) CreateConn(db, output string) (int, error) {
	port := getFreePort()
	go func(port int) {
		tempDir, err := ioutil.TempDir(p.tempDir, "mysql-backup")
		if err != nil {
			return
		}
		c, err := shellwords.Parse(fmt.Sprintf("nc -l -p %d > %s/%s", port, tempDir, output))
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Start()
		cmd.Wait()
	}(port)
	return port, nil
}

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
