package webserver

import (

	// "errors"
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"vendingMaxine/src/webserver/globals"
	"vendingMaxine/src/webserver/middleware"
	"vendingMaxine/src/webserver/routes"
)

var (
	Router *gin.Engine
)

func setup_gin_logger() {
	// Disable Console Color, you don't need console color when writing the logs to file.
	// gin.DisableConsoleColor()

	// By default gin.DefaultWriter = os.Stdout
	// gin.DefaultWriter = io.MultiWriter(os.Stdout)

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// ex: ::1 - [Fri, 07 Dec 2018 17:04:38 JST] "GET /ping HTTP/1.1 200 122.767µs "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36" "
	Router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	Router.Use(gin.Recovery())
}

func GinWebserver() {

	// :)
	// webserver file-structure and code organization based on
	// https://betterprogramming.pub/how-to-create-a-simple-web-login-using-gin-for-golang-9ac46a5b0f89

	// Set the router as the default one provided by Gin
	Router = gin.Default()
	setup_gin_logger()
	Router.Static("/assets", "src/webserver/assets")
	Router.LoadHTMLGlob("src/webserver/templates/*")
	Router.Use(sessions.Sessions("session", cookie.NewStore(globals.Secret)))

	public := Router.Group("/")
	routes.PublicRoutes(public)

	private := Router.Group("/")
	private.Use(middleware.AuthRequired)
	routes.PrivateRoutes(private)

	// Start serving the application
	Router.Run("localhost:8081")

}
