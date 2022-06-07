package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

func main() {

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Define the route for the index page and display the index.html template
	// To start with, we'll use an inline route handler. Later on, we'll create
	// standalone functions that will be used as route handlers.
	router.GET("/userv/:userv_name", func(c *gin.Context) {

		// read userv_name from "parameter in path"  (ref: https://gin-gonic.com/docs/examples/param-in-path/ )
		// ex: "alpha"
		userv_name := c.Param("userv_name")
		if userv_name == "" {
			err := errors.New(fmt.Sprintf("unit-service '%s' not found!", userv_name))
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		}

		// Read consumerSelectionLatestJson_bytes from file
		consumerSelectionLatestJson_filepath := "unit-services/" + userv_name + "/consumer-selection.latest.json"
		consumerSelectionLatestJson_bytes, err := os.ReadFile(consumerSelectionLatestJson_filepath)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
			return
		}
		consumerSelectionLatestJson_string := string(consumerSelectionLatestJson_bytes)

		// Read productsSchemaJson_bytes from file
		productsSchemaJson_bytes, err := os.ReadFile("config/products.schema.json")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		}
		productsSchemaJson_string := string(productsSchemaJson_bytes)

		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"index.html",
			// Pass templating data
			gin.H{
				"productsSchemaJson":          productsSchemaJson_string,
				"consumerSelectionLatestJson": consumerSelectionLatestJson_string,
			},
		)

	})

	// Start serving the application
	router.Run()

}
