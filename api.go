package main

import (
	"net/http"

	"github.com/labstack/echo"
)

// API is a defined as struct bundle
// for api. Feel free to organize
// your app as you wish.
type API struct {
	storage BackupStorage
}

type FullBackupRes struct {
	Message string `json:"message"`
	Key     string `json:"key"`
}

type IncBackupRes struct {
	Message string `json:"message"`
	Key     string `json:"key"`
}

type GetLastLSNRes struct {
	LastLSN string `json:"last_lsn"`
}

const ROOT_DIR = "/Users/taku_k/playground/mysql-backup-test/backups-dir"

const TIME_FORMAT = "2006-01-02-15-04-05"

// Bind attaches api routes
func (api *API) Bind(group *echo.Group) {
	group.GET("/v0/conf", api.ConfHandler)
	group.POST("/v0/:db/full-backup", api.fullBackupHandler)
	group.GET("/v0/:db/last-lsn", api.getLastLSNHandler)
	group.POST("/v0/:db/inc-backup/:last-lsn", api.incBackupHandler)
}

// ConfHandler handle the app config, for example
func (api *API) ConfHandler(c echo.Context) error {
	app := c.Get("app").(*App)
	return c.JSON(200, app.Conf.Root)
}

func (api *API) fullBackupHandler(c echo.Context) error {
	db := c.Param("db")
	body := c.Request().Body

	key, err := saveFullBackupFromReq(api.storage, body, db)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &FullBackupRes{
		Message: "success",
		Key:     key,
	})
}

func (api *API) getLastLSNHandler(c echo.Context) error {
	db := c.Param("db")

	lastLsn, err := api.storage.GetLastLSN(db)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &GetLastLSNRes{
		LastLSN: lastLsn,
	})
}

func (api *API) incBackupHandler(c echo.Context) error {
	db := c.Param("db")
	from := c.Param("from")
	body := c.Request().Body

	key, err := saveIncBackupFromReq(api.storage, body, db, from)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &IncBackupRes{
		Message: "success",
		Key:     key,
	})
}
