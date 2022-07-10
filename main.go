package main

import (
	"github.com/labstack/echo"
	"jobScrapper/scrapper"
	"net/http"
)

func main() {
	scrapper.Scrape("term")
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		name := c.QueryParams().Get("name")
		return c.String(http.StatusOK, "Hello "+name)
	})
	e.Start(":9999")
}
