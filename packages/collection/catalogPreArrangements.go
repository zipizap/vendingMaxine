package collection

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Updates catalogDirTgzData by applying pre-arrangements
func catalogPreArrangements(catalogDirTgzDataPtr *[]byte) (err error) {
	// create tmpDir
	var tmpDir string
	{
		tmpDir, err = os.MkdirTemp("", "catPreArrangeTmp")
		if err != nil {
			return err
		}
		// Function to remove temporary directory
		defer func() {
			// Using os.RemoveAll to remove the directory and its contents
			err := os.RemoveAll(tmpDir)
			if err != nil {
				// Log the error if any occurs while removing the directory
				log.Printf("Error while removing temporary directory: %v", err)
			}
		}()
	}

	// extract catalogDirTgzData to tmpDir, so files there can be changed as needed
	{
		err = extractTgz2Dir(*catalogDirTgzDataPtr, tmpDir)
		if err != nil {
			return err
		}
	}

	// apply pre-arrengements (files in tmpDir can be changed as needed)
	{
		// catalogPreArrangement_SchemaYaml2SchemaJson:
		// If Schema.yaml exists, create (overwrite) Schema.json
		// In any case assure Schema.json exists  (even when Schema.yaml does not exist)
		{
			err := catalogPreArrangement_SchemaYaml2SchemaJson(tmpDir)
			if err != nil {
				return err
			}
		}
	}

	// recompress tmpDir into updated catalogDirTgzData
	{
		tmpBytes, err := compressDir2Tgz(tmpDir)
		if err != nil {
			return err
		}
		catalogDirTgzDataPtr = &tmpBytes
	}
	return nil
}

// catalogPreArrangement_SchemaYaml2SchemaJson function
// If Schema.yaml exists, create (overwrite) Schema.json
// In any case assure Schema.json exists  (even when Schema.yaml does not exist)
func catalogPreArrangement_SchemaYaml2SchemaJson(tmpDir string) (err error) {
	// If Schema.yaml exists, create (overwrite) Schema.json
	schemaYamlFullpath := filepath.Join(tmpDir, "Schema.yaml")
	schemaJsonFullpath := filepath.Join(tmpDir, "Schema.json")
	if fileExists(schemaYamlFullpath) {
		schemaYamlContent, err := os.ReadFile(schemaYamlFullpath)
		if err != nil {
			return err
		}
		schemaJsonContent, err := _generateJsonFromYaml(schemaYamlContent)
		if err != nil {
			return err
		}
		// create or overwrite
		err = os.WriteFile(schemaJsonFullpath, schemaJsonContent, 0400)
		if err != nil {
			return err
		}
	}

	// In any case assure Schema.json exists  (even when Schema.yaml does not exist)
	// At this point, the Schema.json file should exist, either generated from Schema.yaml or supplied directly from user
	if !fileExists(schemaJsonFullpath) {
		return fmt.Errorf("file catalogDir/Schema.json not found - unexpected error")
	}
	return nil
}

// _generateJsonFromYaml function converts yaml to json.
// The yamlInput must be a yaml-map (equivalent to golang map[string]interface)
// The json elements might have a different order than the original yaml elements. However each element of yamlInput will be added an additional json-field `"propertyOrder": 3`
// containing the element-1index order from the original yaml.
//
// Example:
//
//	--- inputYaml ---
//	prop9:
//	  type: string
//	prop8:
//	  type: string
//	prop6:
//	  type: string
//	prop3:
//	  type: string
//
//	--- outputJson ---
//	{
//		"prop3": {
//			"propertyOrder": 4,
//			"type": "string"
//		},
//		"prop6": {
//			"propertyOrder": 3,
//			"type": "string"
//		},
//		"prop8": {
//			"propertyOrder": 2,
//			"type": "string"
//		},
//		"prop9": {
//			"propertyOrder": 1,
//			"type": "string"
//		}
//	}
func _generateJsonFromYaml(yamlBytes []byte) (jsonBytes []byte, err error) {
	// Unmarshal the yaml into a map
	var yMap map[string]interface{}
	{
		err = yaml.Unmarshal(yamlBytes, &yMap)
		if err != nil {
			return nil, err
		}
	}

	// set jsonString
	var jsonString string
	{
		// Catalog_title_and_description_in_markdown
		{
			catalogTitleAndDescriptionInMarkdown, ok := yMap["Catalog_title_and_description_in_markdown"].(string)
			if !ok {
				return nil, fmt.Errorf("Error reading Catalog_title_and_description_in_markdown as string")
			}

			var title, description string
			{
				title, description, err = _parse_titleAndDescriptionInMarkdown(catalogTitleAndDescriptionInMarkdown)
				if err != nil {
					return nil, err
				}
			}

			jsonString = `
{
	"title": ` + title + `,
	"description": ` + description + `,
	"type": "object",
	"options": {
		"disable_collapse": true
	},
	"properties": {
`
		}

		// Catalog_products
		{

		}
	}

	// // Add a "propertyOrder" field to each top-level map element
	// order := 1
	// for k, v := range data {
	// 	// Check if the value is a map
	// 	if m, ok := v.(map[interface{}]interface{}); ok {
	// 		// Convert map[interface{}]interface{} to map[string]interface{}
	// 		newMap := make(map[string]interface{})
	// 		for mk, mv := range m {
	// 			if mkStr, ok := mk.(string); ok {
	// 				newMap[mkStr] = mv
	// 			}
	// 		}
	// 		// Add "propertyOrder" field
	// 		newMap["propertyOrder"] = order
	// 		order++
	// 		// Replace the old map with the new one
	// 		data[k] = newMap
	// 	}
	// }

	// Marshal the map back into a json byte slice
	jsonBytes, err = json.MarshalIndent(yMap, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}

// _convert_nrOfItems_to_minItems_maxItems is a function that takes a string representation of a range (nrOfItems)
// and returns the minimum and maximum values of that range.
// The input string can be in the form of "x-y", "x-inf", or a single number.
// It returns an error if the input string is not in the correct format.
//
//	+------------+----------+----------+
//	| nrOfItems  | minItems | maxItems |
//	+------------+----------+----------+
//	| 0-inf      | 0        | 1000     |
//	| 1          | 1        | 1        |
//	| 0-5        | 0        | 5        |
//	| 3          | 3        | 3        |
//	| 3-4        | 3        | 4        |
//	+------------+----------+----------+
func _convert_nrOfItems_to_minItems_maxItems(nrOfItems string) (minItems int, maxItems int, err error) {
	// Split the input string by the "-" character
	parts := strings.Split(nrOfItems, "-")

	// Check the number of parts after the split
	switch len(parts) {
	case 1:
		// If there is only one part, it means the input is a single number
		minItems, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, errors.New("invalid input")
		}
		maxItems = minItems
	case 2:
		// If there are two parts, it means the input is a range
		minItems, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, errors.New("invalid input")
		}
		// Check if the second part of the range is "inf"
		if parts[1] == "inf" {
			maxItems = 1000
		} else {
			maxItems, err = strconv.Atoi(parts[1])
			if err != nil {
				return 0, 0, errors.New("invalid input")
			}
		}
	default:
		// If there are more or less than 1 or 2 parts, it means the input is not in the correct format
		return 0, 0, errors.New("invalid input")
	}

	return minItems, maxItems, nil
}

// _parse_titleAndDescriptionInMarkdown function splits titleAndDescriptionMd into titleHtmlJsenc and descriptionHtmlJsenc
// first line becomes the title, third-and-remaining-lines become the description
func _parse_titleAndDescriptionInMarkdown(titleAndDescriptionMd string) (titleHtmlJsenc string, descriptionHtmlJsenc string, err error) {

	// titleHtmlJsenc
	{
		firstLineMd := strings.Split(titleAndDescriptionMd, "\n")[0]

		// remove leading ^# if they exist
		titleMd := regexp.MustCompile(`^#+\s?`).ReplaceAllString(firstLineMd, "")
		titleHtmlBytes := markdown.ToHTML([]byte(titleMd), nil, nil)

		titleHtmlJsencBytes, err := json.Marshal(string(titleHtmlBytes))
		if err != nil {
			return "", "", err
		}
		titleHtmlJsenc = string(titleHtmlJsencBytes)
	}

	// descriptionHtmlJsenc
	{
		// third-and-other-lines are taken as description
		ThirdAndOtherLinesMd := strings.Join(strings.Split(titleAndDescriptionMd, "\n")[2:], "\n")

		descHtlm := markdown.ToHTML([]byte(ThirdAndOtherLinesMd), nil, nil)

		descHtmlJsencBytes, err := json.Marshal(string(descHtlm))
		if err != nil {
			return "", "", err
		}
		descriptionHtmlJsenc = string(descHtmlJsencBytes)
	}

	return titleHtmlJsenc, descriptionHtmlJsenc, nil
}
