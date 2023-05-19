package web

import (
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"gopkg.in/fsnotify.v1"

	echoSwagger "github.com/swaggo/echo-swagger" // echo-swagger middleware
)

var e *echo.Echo
var slog *zap.SugaredLogger
var jsonPrettyIdent = "  "

// Ex: go startServer(":8080")
func startServer(listeningAddress string) {
	slog.Infoln("Http-web-server - Starting")
	e = echo.New()

	slog.Infoln("Http-web-server - setup zap logger")
	{
		e.Use(middleware.BodyDump(
			func(c echo.Context, reqBody, resBody []byte) {
				if slog.Level().CapitalString() == "INFO" {
					slog.Infow("request",
						"ReqRemoteAddr", c.Request().RemoteAddr,
						"ReqMethod", c.Request().Method,
						"ReqURI", c.Request().URL,
						"ResStatus", c.Response().Status,
					)
				} else if slog.Level().CapitalString() == "DEBUG" {
					var resBodyToLog string
					if strings.Contains(strings.ToLower(c.Request().URL.Path), "swagger") {
						// swagger request - no log
						resBodyToLog = "...swagger ignore..."
					} else if strings.Contains(strings.ToLower(c.Response().Header().Get(echo.HeaderContentType)), "json") {
						// json response-body - log it
						resBodyToLog = string(resBody)
					} else {
						// non-json response-body - no log
						resBodyToLog = "...non-json-content..."
					}
					slog.Debugw("request",
						"ReqRemoteAddr", c.Request().RemoteAddr,
						"ReqMethod", c.Request().Method,
						"ReqURI", c.Request().URL,
						"ReqBody", string(reqBody),
						"ResStatus", c.Response().Status,
						"ResBody", resBodyToLog,
					)
				}
			}))
	}

	slog.Infoln("Http-web-server - setup template renderer (with hot-reload)")
	{
		e.Renderer = &TemplateRenderer{
			templates: template.Must(template.ParseGlob("web/templates/*.html")),
		}
		// hot-reload of templates changed on disk
		{
			// Initialize and start the watcher
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				slog.Fatal(err)
			}
			defer watcher.Close()
			// Watch the templates directory for changes
			go watchTemplates(watcher, "web/templates", e.Renderer.(*TemplateRenderer))
		}
	}

	slog.Infoln("Http-web-server - setup routes")
	{

		slog.Infoln("Enabling API routes")
		{
			slog.Infoln("Enabling API route: GET /api/v1/catalogs")
			e.GET("/api/v1/catalogs", getCatalogsOverview)

			slog.Infoln("Enabling API route: GET /api/v1/collections")
			e.GET("/api/v1/collections", getCollectionsOverview)

			slog.Infoln("Enabling API route: POST /api/v1/collections")
			e.POST("/api/v1/collections", postCollectionNew)

			slog.Infoln("Enabling API route: GET /api/v1/collections/:collection-name")
			e.GET("/api/v1/collections/:collection-name", getCollectionEditPrepinfo)

			slog.Infoln("Enabling API route: PUT /api/v1/collections/:collection-name")
			e.PUT("/api/v1/collections/:collection-name", putCollectionEditSave)

			slog.Infoln("Enabling API route: GET /api/v1/collections/:collection-name/replayable")
			e.GET("/api/v1/collections/:collection-name/replayable", getCollectionReplayable)

		}

		slog.Infoln("Enabling WEB routes")
		{
			slog.Infoln("Enabling WEB route:  /swagger/index.html")
			e.GET("/swagger/*", echoSwagger.WrapHandler)

			// http://zzz/static --> ./web/static
			slog.Infoln("Enabling WEB route:  GET /static")
			e.Static("/static", "web/static")

			slog.Infoln("Enabling WEB route:  GET /")
			e.GET("/", getWebHome)

			slog.Infoln("Enabling WEB route:  GET /catalogs")
			e.GET("/catalogs", getWebCatalogs)

			slog.Infoln("Enabling WEB route:  GET /collections")
			e.GET("/collections", getWebCollections)

			slog.Infoln("Enabling WEB route:  GET /collections/:collection-name")
			e.GET("/collections/:collection-name", getWebCollection)

		}

	}

	slog.Infoln("Http-web-server - listening forever-and-ever in goroutine")
	e.Logger.Fatal(e.Start(listeningAddress))
}

// Stops server gracefully within 10s (or ungracefully after 10sec)
func closeServer() error {
	slog.Infoln("Closing Http-web-server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	return e.Close()
}
