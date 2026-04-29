package main

import (
	"cmp"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	charmlog "charm.land/log/v2"
	"github.com/alecthomas/kong"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

//go:embed dist/*
var assets embed.FS

type cli struct {
	DevMode bool     `name:"dev" short:"d"`
	DevURL  *url.URL `name:"dev-url" default:"http://localhost:5173" help:"URL of the development server"`
}

func main() {
	var cl cli
	kong.Parse(&cl)

	if err := cl.Start(); err != nil {
		panic(err)
	}
}

func (cl cli) Start() error {
	e := echo.New()

	e.Use(
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogLatency:   true,
			LogStatus:    true,
			LogMethod:    true,
			LogRequestID: true,
			LogRoutePath: true,
			LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
				msg := fmt.Sprintf("%v %v (%v - %v)",
					v.Method,
					cmp.Or(v.RoutePath, c.Path()),
					v.Status,
					http.StatusText(v.Status),
				)
				l := c.Logger().With("id", v.RequestID, "duration", v.Latency)
				if v.Error != nil {
					l = l.With("error", v.Error)
				}

				switch {
				case v.Status >= 500:
					l.Error(msg)
				case v.Status >= 400:
					l.Warn(msg)
				case v.Status >= 300:
					l.Debug(msg)
				case v.Status >= 200:
					l.Info(msg)
				default:
					l.Debug(msg)
				}
				return nil
			},
		}),
	)

	if cl.DevMode {
		e.Logger = slog.New(charmlog.NewWithOptions(os.Stdout, charmlog.Options{
			Formatter: charmlog.TextFormatter,
		}))

		e.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
			Balancer: middleware.NewRoundRobinBalancer([]*middleware.ProxyTarget{{URL: cl.DevURL}}),
			Skipper:  func(c *echo.Context) bool { return strings.HasPrefix(c.Path(), "/api") },
		}))
	} else {
		e.StaticFS("/*", echo.MustSubFS(assets, "dist"))
	}

	e.GET("/api/token/refresh", func(c *echo.Context) error {
		// TODO: get the header from the req.
		return c.JSON(http.StatusOK, map[string]string{"token": "new-token"})
	})

	return e.Start(":8080")
}
