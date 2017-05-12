package api

import (
	"fmt"
	"net/http"

	"time"

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
	group.GET("/conf", api.ConfHandler)
	group.POST("/full-backup/:db", api.fullBackupHandler)
	group.GET("/last-lsn/:db", api.getLastLSNHandler)
	group.POST("/inc-backup/:db/:last-lsn", api.incBackupHandler)

	// Restore
	group.GET("/restore/:db/:from", api.restoreHandler)
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

func (api *API) restoreHandler(c echo.Context) error {
	db := c.Param("db")
	from := c.Param("from")

	t, err := time.Parse("2006-01-02", from)
	if err != nil {
		return err
	}
	t = t.AddDate(0, 0, 1)
	keys, _ := api.storage.SearchConsecutiveIncBackups(db, t)
	fmt.Println(keys)
	return c.JSON(http.StatusOK, "aaa")
}
