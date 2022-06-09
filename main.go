package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmespath/go-jmespath"
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

func main() {

	// Set the router as the default one provided by Gin
	router = gin.Default()

	setup_gin_logger(router)

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
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}

		// Read consumerSelectionPreviousJson_bytes from file
		consumerSelectionPreviousJson_filepath := "unit-services/" + userv_name + "/consumer-selection.latest.json"
		consumerSelectionPreviousJson_bytes, err := os.ReadFile(consumerSelectionPreviousJson_filepath)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
			return
		}
		consumerSelectionPreviousJson_string := string(consumerSelectionPreviousJson_bytes)

		// Read productsSchemaJson_bytes from file
		productsSchemaJson_bytes, err := os.ReadFile("config/products.schema.json")
		if err != nil {
			c.HTML(http.StatusInternalServerError, "error.tmpl", gin.H{"error": err.Error()})
		}
		productsSchemaJson_string := string(productsSchemaJson_bytes)

		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"index.tmpl",
			// Pass templating data
			gin.H{
				"productsSchemaJson":            productsSchemaJson_string,
				"consumerSelectionPreviousJson": consumerSelectionPreviousJson_string,
			},
		)

	})

	router.POST("/userv", func(c *gin.Context) {
		consumerSelectionNewJson_bytes, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// read json
		var consumerSelectionNewJson interface{}
		err = json.Unmarshal([]byte(consumerSelectionNewJson_bytes), &consumerSelectionNewJson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		unitService_name, err := jmespath.Search(`"unit-services"[0].name`, consumerSelectionNewJson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":       "posted",
			"unit-service": unitService_name,
		})

		fmt.Printf("Processing unit-service '%s'\n", unitService_name)
		// TODO
	})
	// Start serving the application
	router.Run()

}
