package main

import (
	"embed"
	"net/http"

	"github.com/labstack/echo/v5"
)

//go:embed dist/*
var assets embed.FS

func main() {
	e := echo.New()
	e.StaticFS("/", echo.MustSubFS(assets, "dist"))
	e.GET("/api/token/refresh", func(c *echo.Context) error {
		// TODO: get the header from the req.
		return c.JSON(http.StatusOK, map[string]string{"token": "new-token"})
	})

	if err := e.Start(":8080"); err != nil {
		panic(err)
	}
}
