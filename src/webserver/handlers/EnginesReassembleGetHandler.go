package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func EnginesReassembleGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {

		// if err != nil {
		// 	log.Error(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
