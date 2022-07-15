package middleware

import (
	"vendingMachine/src/webserver/globals"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"net/http"

	log "github.com/sirupsen/logrus"
)

func AuthRequired(c *gin.Context) {
	if globals.DebugDisableLogin {
		c.Next()
		return
	}
	session := sessions.Default(c)
	user := session.Get(globals.Userkey)
	if user == nil {
		log.Println("User not logged in")
		c.Redirect(http.StatusMovedPermanently, "/login")
		c.Abort()
		return
	}
	c.Next()
}
