package handlers

import (
	"github.com/gin-contrib/sessions"

	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	"vendingMachine/src/webserver/globals"
)

func LogoutGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get(globals.Userkey)
		log.Println("logging out user:", user)
		if user == nil {
			// user is not logged-in
			log.Println("Invalid session token")
			return
		}
		session.Delete(globals.Userkey)
		if err := session.Save(); err != nil {
			log.Println("Failed to save session:", err)
			return
		}

		c.Redirect(http.StatusMovedPermanently, "/")
	}
}
