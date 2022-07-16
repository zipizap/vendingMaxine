package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vendingMaxine/src/collection"

	log "github.com/sirupsen/logrus"
)

func CollectionsGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// user := session.Get(globals.Userkey)
		collections, err := collection.GetAllCollections()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}
		var collectionsOverallSblock []collection.StatusBlock
		for _, a_col := range collections {
			a_col_lastRsf, err := a_col.LastRsf()
			if err != nil {
				log.Error(err)
			}
			collectionsOverallSblock = append(collectionsOverallSblock, a_col_lastRsf.Status.Overall)
		}
		c.HTML(http.StatusOK, "collections.tmpl", gin.H{
			"collectionsOverallSblock": collectionsOverallSblock,
		})
	}
}
