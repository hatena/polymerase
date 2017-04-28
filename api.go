package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

// API is a defined as struct bundle
// for api. Feel free to organize
// your app as you wish.
type API struct{}

type GetLastLSN struct {
	LastLSN string `json:"last_lsn"`
}

const ROOT_DIR = "/Users/taku_k/playground/mysql-backup-test/backups-dir"

const TIME_FORMAT = "2006-01-02-15-04-05"

// Bind attaches api routes
func (api *API) Bind(group *echo.Group) {
	group.GET("/v0/conf", api.ConfHandler)
	group.POST("/v0/:db/full-backup", api.fullBackupHandler)
	group.GET("/v0/:db/last-lsn/:from", api.getLastLSNHandler)
	group.POST("/v0/:db/inc-backup/:from", api.incBackupHandler)
}

// ConfHandler handle the app config, for example
func (api *API) ConfHandler(c echo.Context) error {
	app := c.Get("app").(*App)
	return c.JSON(200, app.Conf.Root)
}

func (api *API) fullBackupHandler(c echo.Context) error {
	db := c.Param("db")
	body := c.Request().Body

	tmpFile, err := ioutil.TempFile("", "mysql-backup")
	if err != nil {
		return err
	}

	_, err = io.Copy(tmpFile, body)
	if err != nil {
		return errors.Wrap(err, "Can't io.Copy(tmpFile, body)")
	}

	now := time.Now()
	fileDir := fmt.Sprintf("%s/%s/%s/%s", ROOT_DIR, db, now.Format("2006-01-02"), now.Format(TIME_FORMAT))
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		return errors.Wrap(err, "Can't mkdir")
	}
	fileName := fmt.Sprintf("%s/base.tar.gz", fileDir)
	if err := os.Rename(tmpFile.Name(), fileName); err != nil {
		return errors.Wrap(err, "Can't rename tmfile to filename")
	}

	if err := os.Chdir(fileDir); err != nil {
		return err
	}
	cmd := fmt.Sprintf("gunzip -c %s | tar xf - xtrabackup_checkpoints", fileName)
	if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
		return errors.Wrap(err, "cant extract xtrabackup_checkpoints")
	}

	return c.JSON(http.StatusCreated, "success")
}

func (api *API) getLastLSNHandler(c echo.Context) error {
	db := c.Param("db")
	from := c.Param("from")
	fileDir := fmt.Sprintf("%s/%s/%s", ROOT_DIR, db, from)

	latestBackupDir := ""
	var latestBackupTime time.Time
	files, err := ioutil.ReadDir(fileDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		curBackupTime, err := time.Parse(TIME_FORMAT, f.Name())
		if err != nil {
			return err
		}
		if latestBackupDir == "" {
			latestBackupDir = filepath.Join(fileDir, f.Name())
			latestBackupTime = curBackupTime
		} else {
			if !latestBackupTime.After(curBackupTime) {
				latestBackupDir = filepath.Join(fileDir, f.Name())
				latestBackupTime = curBackupTime
			}
		}
	}

	// Extract a LSN from a last checkpoint
	cpFile := fmt.Sprintf("%s/xtrabackup_checkpoints", latestBackupDir)
	fp, err := os.Open(cpFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	var lastLsn string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "to_lsn") {
			lastLsn = strings.TrimSpace(strings.Split(line, "=")[1])
		}
	}

	res := &GetLastLSN{
		LastLSN: lastLsn,
	}

	return c.JSON(http.StatusOK, res)
}

func (api *API) incBackupHandler(c echo.Context) error {
	db := c.Param("db")
	from := c.Param("from")
	body := c.Request().Body

	id, _ := uuid.NewV4()
	tmpFileDir := fmt.Sprintf("%s/tmp", ROOT_DIR)
	if err := os.MkdirAll(tmpFileDir, 0777); err != nil {
		return err
	}
	tmpFileName := fmt.Sprintf("%s/%s.tar.gz", tmpFileDir, id)
	outfile, err := os.Create(tmpFileName)
	if err != nil {
		return err
	}
	_, err = io.Copy(outfile, body)
	if err != nil {
		return err
	}
	outfile.Close()

	now := time.Now()
	fileDir := fmt.Sprintf("%s/%s/%s/%s", ROOT_DIR, db, from, now.Format(TIME_FORMAT))
	if err := os.MkdirAll(fileDir, 0777); err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s/inc.gz", fileDir)
	if err := os.Rename(tmpFileName, fileName); err != nil {
		return err
	}

	if err := exec.Command("sh", "-c", fmt.Sprintf("gunzip -c %s | tar xf - xtrabackup_checkpoints", fileName)).Run(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, "success")

	return nil
}
