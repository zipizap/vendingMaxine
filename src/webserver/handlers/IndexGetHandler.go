package handlers

import (
	"github.com/gin-contrib/sessions"

	"net/http"

	"github.com/gin-gonic/gin"

	"vendingMachine/src/webserver/globals"
)

func IndexGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.Userkey)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"content": "",
			"user":    user,
		})
	}
}
