package handlers

import (
	"net/http"
	"vendingMaxine/src/collection"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func EnginesReassembleGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := collection.CollectionsAllAssembly()
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
