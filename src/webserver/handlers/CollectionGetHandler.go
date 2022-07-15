package handlers

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	"vendingMachine/src/webserver/helpers"
)

// Reads collection_name from urlpath
// Read consumerSelectionPreviousJson_string from theCollection.LastRsf()
//   NOTE: if last_rsf does not exist, this will fail badly... todo: improve this
// Read productsSchemaJson_string from file PRODUCT_SCHEMA_JSON_FILEPATH
// Send html-response composed by:
//		"collection.tmpl"
//		consumerSelectionPreviousJson_string
//		productsSchemaJson_string
//
func CollectionGetHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// user := session.Get(globals.Userkey)

		// read theCollection_name from "parameter in path"  (ref: https://gin-gonic.com/docs/examples/param-in-path/ )
		// ex: "alpha"
		theCollection_name := c.Param("collection_name")

		// Read consumerSelectionPreviousJson_string from theCollection.LastRsf
		// Read productsSchemaJson_string from file
		//
		// 		NOTE: if lasT_rsf does not exist (ex: new collection) then err != nil
		// 		and new collectino will never work
		// 		todo: a newly-created-collection will never work as it does not have a last_rsf for bootstraping
		consumerSelectionPreviousJson_string, productsSchemaJson_string, err := helpers.Get_selectionPrevious_and_prodSchema_from_collection(theCollection_name)
		if err != nil {
			log.Error(err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}

		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the collection.html template
			"collection.tmpl",
			// Pass templating data
			gin.H{
				"productsSchemaJson":            productsSchemaJson_string,
				"consumerSelectionPreviousJson": consumerSelectionPreviousJson_string,
			},
		)

	}
}
