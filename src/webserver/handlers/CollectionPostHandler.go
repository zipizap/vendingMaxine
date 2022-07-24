package handlers

import (
	"fmt"
	"io/ioutil"

	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"

	"vendingMaxine/src/collection"
)

func CollectionPostHandler() gin.HandlerFunc {
	// a) Read consumerSelectionNewJson_string 		, from c.Request.Body
	// b.1) Read consumerSelectionPreviousJson_string
	// b.2) Read productsSchemaJson_string
	// c) Compose webdata and call col.NewRsf_from_WebconsumerSelection(webdata)
	return func(c *gin.Context) {
		// session := sessions.Default(c)
		// user := session.Get(globals.Userkey)

		theCollection_name := c.Param("collection_name")

		// a) Read consumerSelectionNewJson_string 		, from c.Request.Body
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
		//
		//----------------------------------------------------------------------------------------------
		//
		// c.JSON(http.StatusOK, gin.H{
		// 	"collection": theCollection_name,
		// })
		//
		// if err != nil {
		// 	log.Error(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		// b.1) Read consumerSelectionPreviousJson_string
		// b.2) Read productsSchemaJson_string
		var consumerSelectionPreviousJson_string string
		var productsSchemaJson_string string
		log.Info(fmt.Sprintf("Processing collection '%s'", theCollection_name))
		col, err := collection.CollectionGet(theCollection_name)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cantDo, consumerSelectionPreviousJson_string, productsSchemaJson_string, err := col.Get_selectionPrevious_and_prodSchema_from_collection()
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if cantDo {
			err = fmt.Errorf("this collection cannot be edited at this moment. Maybe someone edited already and its being processed")
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// c) Compose webdata and call col.NewRsf_from_WebconsumerSelection(webdata)
		webdata := map[string]string{
			"products.schema.json":             productsSchemaJson_string,
			"consumer-selection.previous.json": consumerSelectionPreviousJson_string,
			"consumer-selection.next.json":     consumerSelectionNewJson_string,
		}
		col, err = collection.CollectionGet(theCollection_name)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//cantDo, rsf_new, err = col.NewRsf_from_WebconsumerSelection(webdata)
		cantDo, _, err = col.NewRsf_from_WebconsumerSelection(webdata)
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
		// {
		// 	collection.RunnersOfProcEngs_wg.Wait()
		// 	spew.Dump(rsf_new)
		// }
	}
}
