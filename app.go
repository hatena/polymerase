package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/nu7hatch/gouuid"
	"github.com/olebedev/config"
)

// App struct.
// There is no singleton anti-pattern,
// all variables defined locally inside
// this struct.
type App struct {
	Engine *echo.Echo
	Conf   *config.Config
	API    *API
}

// NewApp returns initialized struct
// of main server application.
func NewApp(opts ...AppOptions) *App {
	options := AppOptions{}
	for _, i := range opts {
		options = i
		break
	}

	options.init()

	// Parse config yaml string from ./conf.go
	conf, err := config.ParseYaml(confString)
	if err != nil {
		panic(err)
	}

	// Set config variables delivered from main.go:11
	// Variables defined as ./conf.go:3
	conf.Set("debug", debug)
	conf.Set("commitHash", commitHash)

	// Parse environ variables for defined
	// in config constants
	conf.Env()

	// Make an engine
	engine := echo.New()

	// Set up echo debug level
	engine.Debug = conf.UBool("debug")

	// Regular middlewares
	engine.Use(middleware.Recover())

	engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${method} | ${status} | ${uri} -> ${latency_human}` + "\n",
	}))

	// Initialize the application
	app := &App{
		Conf:   conf,
		Engine: engine,
		API:    &API{},
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

	// Bind api hadling for URL api.prefix
	app.API.Bind(
		app.Engine.Group(
			app.Conf.UString("api.prefix"),
		),
	)

	engine.Logger.SetLevel(log.DEBUG)

	return app
}

// Run runs the app
func (app *App) Run() {
	err := app.Engine.Start(":" + app.Conf.UString("port"))
	if err != nil {
		panic(err)
	}
}

// AppOptions is options struct
type AppOptions struct{}

func (ao *AppOptions) init() { /* write your own*/ }
