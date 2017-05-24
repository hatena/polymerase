package api

import (
	"net"
	"os/exec"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nu7hatch/gouuid"
	"github.com/taku-k/polymerase/pkg/storage"
)

// App struct.
// There is no singleton anti-pattern,
// all variables defined locally inside
// this struct.
type App struct {
	Engine *echo.Echo
	Conf   *Config
	API    *API
}

// NewApp returns initialized struct
// of main server application.
func NewApp(conf *Config) (*App, error) {
	// Make an engine
	engine := echo.New()

	// Regular middlewares
	engine.Use(middleware.Recover())

	engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${method} | ${status} | ${uri} -> ${latency_human}` + "\n",
	}))

	// FIXME: use option to choose a backup storage
	s, err := storage.NewLocalBackupStorage(conf.Config)
	if err != nil {
		return nil, err
	}
	bm := NewBackupManager(conf.Config)
	//p := NewNCPool(conf)

	// Initialize the application
	app := &App{
		Conf:   conf,
		Engine: engine,
		API: &API{
			storage: s,
			bm:      bm,
			//pool:    p,
		},
	}

	// Map app and uuid for every requests
	app.Engine.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("app", app)
			id, _ := uuid.NewV4()
			c.Set("uuid", id)
			return next(c)
		}
	})

	// Bind api handling for URL api.prefix
	app.API.Bind(
		app.Engine.Group(
			conf.HTTPApiPrefix,
		),
	)

	engine.Logger.SetLevel(log.DEBUG)

	return app, nil
}

func ensureExistXtrabackup() error {
	return exec.Command("which", "xtrabackup").Run()
}

// Run runs the app
func (app *App) Run(l net.Listener) {
	//if err := ensureExistXtrabackup(); err != nil {
	//	log.Error("xtrabackup command not found")
	//	panic(err)
	//}
	//err := app.Engine.Start(":" + strconv.Itoa(app.Conf.Port))
	app.Engine.Listener = l
	err := app.Engine.Start("")
	if err != nil {
		panic(err)
	}
}
