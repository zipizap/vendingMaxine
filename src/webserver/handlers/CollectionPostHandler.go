package handlers

import (
	"fmt"
	"io/ioutil"

	"github.com/davecgh/go-spew/spew"

	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	"vendingMachine/src/collection"
	"vendingMachine/src/webserver/helpers"
)

// Read consumerSelectionNewJson_string 		, from c.Request.Body
// Read consumerSelectionPreviousJson_string
// Read productsSchemaJson_string
// Compose webdata and call col.NewRsf_from_WebconsumerSelection(webdata)
func CollectionPostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// user := session.Get(globals.Userkey)

		theCollection_name := c.Param("collection_name")

		// Read consumerSelectionNewJson_string 		, from c.Request.Body
		consumerSelectionNewJson_bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		consumerSelectionNewJson_string := string(consumerSelectionNewJson_bytes)
		// ATP are defined
		//	 - consumerSelectionNewJson_bytes
		//	 - consumerSelectionNewJson_string

		//	// NOTE: this is how we could make jmespath-queries on consumerSelectionNewJson_string:
		// var holder interface{}
		// err = json.Unmarshal([]byte(consumerSelectionNewJson_string), &holder)
		// if err != nil {
		// 	log.Error(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }
		// jmespath_query := `" unit-services"[0].name`
		// jmespath_result, err := jmespath.Search(jmespath_query, holder)
		// if err != nil {
		// 		log.Error(err)
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		// c.JSON(http.StatusOK, gin.H{
		// 	"collection": theCollection_name,
		// })
		//
		// if err != nil {
		// 	log.Error(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		// Read consumerSelectionPreviousJson_string
		// Read productsSchemaJson_string
		log.Info(fmt.Sprintf("Processing collection '%s'", theCollection_name))
		consumerSelectionPreviousJson_string, productsSchemaJson_string, err := helpers.Get_selectionPrevious_and_prodSchema_from_collection(theCollection_name)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Compose webdata and call col.NewRsf_from_WebconsumerSelection(webdata)
		webdata := map[string]string{
			"products.schema.json":             productsSchemaJson_string,
			"consumer-selection.previous.json": consumerSelectionPreviousJson_string,
			"consumer-selection.next.json":     consumerSelectionNewJson_string,
		}
		col, err := collection.GetCollection(theCollection_name)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cantDo, rsf_new, err := col.NewRsf_from_WebconsumerSelection(webdata)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if cantDo {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// ATP: runProcessingEngines() is running async
		//  If we want to wait for it to complete running, and then show the rsf_new updated
		// with the ProcEng final results, then we need to use collction.RunnersOfProcEngs_wg.Wait()
		{
			collection.RunnersOfProcEngs_wg.Wait()
			spew.Dump(rsf_new)
		}

	}
}
