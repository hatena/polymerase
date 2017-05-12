package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/taku-k/xtralab/pkg/storage"
)

type API struct {
	storage storage.BackupStorage
	bm      *BackupManager
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

// Bind attaches api routes
func (api *API) Bind(group *echo.Group) {
	group.GET("/v0/conf", api.ConfHandler)
	group.POST("/v0/:db/full-backup", api.fullBackupHandler)
	group.GET("/v0/:db/last-lsn", api.getLastLSNHandler)
	group.POST("/v0/:db/inc-backup/:last-lsn", api.incBackupHandler)

	//group.GET("/v0/:db/restore", api.)
}

// ConfHandler handle the app config, for example
func (api *API) ConfHandler(c echo.Context) error {
	app := c.Get("app").(*App)
	return c.JSON(200, fmt.Sprintf("%#v", app.Conf))
}

func (api *API) fullBackupHandler(c echo.Context) error {
	db := c.Param("db")
	body := c.Request().Body

	key, err := api.bm.saveFullBackupFromReq(api.storage, body, db)
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
	lastLsn := c.Param("last-lsn")
	body := c.Request().Body

	key, err := api.bm.saveIncBackupFromReq(api.storage, body, db, lastLsn)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &IncBackupRes{
		Message: "success",
		Key:     key,
	})
}
