package routes

import (
	"github.com/gin-gonic/gin"

	"vendingMaxine/src/webserver/handlers"
)

func PublicRoutes(g *gin.RouterGroup) {

	g.GET("/", handlers.IndexGetHandler())
	g.GET("/login", handlers.LoginGetHandler())
	//g.POST("/login", handlers.LoginPostHandler())

}

func PrivateRoutes(g *gin.RouterGroup) {

	g.GET("/logout", handlers.LogoutGetHandler())
	g.GET("/collections", handlers.CollectionsGetHandler())
	g.GET("/collection/:collection_name", handlers.CollectionGetHandler())
	g.POST("/collection/:collection_name", handlers.CollectionPostHandler())
	g.GET("/engines/reassemble", handlers.EnginesReassembleGetHandler())

}
