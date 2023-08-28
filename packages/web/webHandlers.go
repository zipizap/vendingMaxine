package web

import (
	"net/http"

	"github.com/labstack/echo/v4"
	// echo-swagger middleware
)

func getWebHome(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func getWebCatalogs(c echo.Context) error {
	return c.Render(http.StatusOK, "catalogs.html", nil)
}

func getWebCollections(c echo.Context) error {
	return c.Render(http.StatusOK, "collections.html", nil)
}

func getWebCollection(c echo.Context) error {
	templateVals := map[string]string{
		"ColName": c.Param("collection-name"),
	}
	return c.Render(http.StatusOK, "collection.html", templateVals)
}
