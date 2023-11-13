package collection

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"vendingMaxine/packages/yaml2json"

	jsoniter "github.com/json-iterator/go"
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
		*catalogDirTgzDataPtr = tmpBytes
	}
	return nil
}

// catalogPreArrangement_SchemaYaml2SchemaJson function:
//   - If Schema.yaml exists, create (overwrite) Schema.json
//   - In any case assure Schema.json exists  (even when Schema.yaml does not exist)
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
	return yaml2json.GenerateJsonFromYaml(yamlBytes)
}
