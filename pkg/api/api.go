package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/taku-k/polymerase/pkg/storage"
)

type API struct {
	storage storage.BackupStorage
	bm      *BackupManager
	//pool    *NCPool
}

type StartFullBackupRes struct {
	Port int `json:"port"`
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

type RestoreSearchRes struct {
	Keys []*storage.BackupFile `json:"keys"`
}

// Bind attaches api routes
func (api *API) Bind(group *echo.Group) {
	group.GET("/conf", api.ConfHandler)
	group.POST("/full-backup/:db", api.fullBackupHandler)
	group.GET("/last-lsn/:db", api.getLastLSNHandler)
	group.POST("/inc-backup/:db/:last-lsn", api.incBackupHandler)
	//group.GET("/full-backup/start/:db", api.startFullBackupHandler)

	// Restore
	group.GET("/restore/search/:db/:from", api.restoreSearchHandler)
	group.GET("/restore/file/:stype", api.restoreFileHandler)
}

// ConfHandler handle the app config, for example
func (api *API) ConfHandler(c echo.Context) error {
	app := c.Get("app").(*App)
	return c.JSON(200, fmt.Sprintf("%#v", app.Conf))
}

//func (api *API) startFullBackupHandler(c echo.Context) error {
//	db := c.Param("db")
//
//	port, err := api.pool.CreateConn(db, "base.tar.gz")
//	if err != nil {
//		return err
//	}
//
//	return c.JSON(http.StatusOK, &StartFullBackupRes{
//		Port: port,
//	})
//}

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

	lastLsn, err := api.storage.GetLatestToLSN(db)
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

func (api *API) restoreSearchHandler(c echo.Context) error {
	db := c.Param("db")
	from := c.Param("from")

	t, err := time.Parse("2006-01-02", from)
	if err != nil {
		return err
	}
	t = t.AddDate(0, 0, 1)
	bfiles, _ := api.storage.SearchConsecutiveIncBackups(db, t)
	return c.JSON(http.StatusOK, &RestoreSearchRes{
		Keys: bfiles,
	})
}

func (api *API) restoreFileHandler(c echo.Context) error {
	//stype := c.Param("stype")
	key := c.QueryParam("key")

	r, err := api.storage.GetFileStream(key)
	if err != nil {
		return err
	}
	return c.Stream(http.StatusOK, "application/gzip", r)
}
