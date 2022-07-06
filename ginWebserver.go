package main

import (
	"encoding/json"
	// "errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"proto-VD/collection"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var router *gin.Engine

func setup_gin_logger(router *gin.Engine) {
	// Disable Console Color, you don't need console color when writing the logs to file.
	// gin.DisableConsoleColor()

	// By default gin.DefaultWriter = os.Stdout
	// gin.DefaultWriter = io.MultiWriter(os.Stdout)

	// LoggerWithFormatter middleware will write the logs to gin.DefaultWriter
	// ex: ::1 - [Fri, 07 Dec 2018 17:04:38 JST] "GET /ping HTTP/1.1 200 122.767µs "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.80 Safari/537.36" "
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// your custom format
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())
}

func ginWebserver() {

	// Set the router as the default one provided by Gin
	router = gin.Default()

	setup_gin_logger(router)

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("vd-internal/templates/*")

	router.GET("/collection/:collection_name", func(c *gin.Context) {

		// read theCollection_name from "parameter in path"  (ref: https://gin-gonic.com/docs/examples/param-in-path/ )
		// ex: "alpha"
		theCollection_name := c.Param("collection_name")

		// if theCollection does not exist, reply with error
		theCollection, err := collection.GetCollection(theCollection_name)
		if err != nil {
			// collection_name not found
			err := fmt.Errorf("collection '%s' not found", theCollection_name)
			log.Error(err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}

		// Read consumerSelectionPreviousJson_string from theCollection.LastRsf
		last_rsf, err := theCollection.LastRsf()
		// NOTE: if lasT_rsf does not exist (ex: new collection) then err != nil
		// and new collectino will never work
		// todo: new collection will never work as it does not have a last_rsf for bootstraping
		if err != nil {
			log.Error(err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}
		consumerSelectionPreviousJson_gzB64 := last_rsf.Status.Overall.LatestUpdateData["consumer-selection.next.json"].(string)
		consumerSelectionPreviousJson_bytes, err := collection.Decode_gzB64_to_bytes(consumerSelectionPreviousJson_gzB64)
		if err != nil {
			log.Error(err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}
		consumerSelectionPreviousJson_string := string(consumerSelectionPreviousJson_bytes)

		// Read productsSchemaJson_string from file
		productsSchemaJson_bytes, err := os.ReadFile("config/products.schema.json")
		if err != nil {
			log.Error(err)
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
		}
		productsSchemaJson_string := string(productsSchemaJson_bytes)

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

	})

	router.POST("/collection/:collection_name", func(c *gin.Context) {
		theCollection_name := c.Param("collection_name")

		consumerSelectionNewJson_bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		var consumerSelectionNewJson interface{}
		err = json.Unmarshal([]byte(consumerSelectionNewJson_bytes), &consumerSelectionNewJson)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// ATP: consumerSelectionNewJson is ready to be read, for example with jmespath :)
		// collection_name, err := jmespath.Search(`" unit-services"[0].name`, consumerSelectionNewJson)
		// if err != nil {
		// 	log.Error(err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"status":     "posted",
			"collection": theCollection_name,
		})

		log.Info(fmt.Sprintf("Processing collection '%s'\n", theCollection_name))
		// TODO
	})
	// Start serving the application
	router.Run()

}
