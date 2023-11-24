package yaml2json

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gomarkdown/markdown"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v2"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var commonRegexpPatterns = map[string]string{
	"alphanumUnderscoreHyphen": "^[A-Za-z0-9_-]+$",
}

// GenerateJsonFromYaml function converts the Yaml to a Json equivalent, with special handling of certain special fields.
// The yamlInput must be a yaml-map (equivalent to golang map[string]interface)
//
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
func GenerateJsonFromYaml(yamlBytes []byte) (jsonBytes []byte, err error) {
	/*
		============ json =======

		jMap	{
					"headerTemplate": "Select products",
					"description": "See our wiki <b>bold</b>",
					"title": " "
					"type": "object",
					"options": {
						"disable_collapse": true
					},
			O1		"properties": {
						"additional_members_of_AADGroup_DataClients": {
							"propertyOrder": 1001 or 1002 or 1003 ... each product with its ordernumber
							"headerTemplate": "AAD-Group 'mycollection_DataClients' members",
							"title": " ",
							"type": "object",
							"options": {
								"collapsed": true
							},
				O2			"properties": {
								"info": {
									"propertyOrder": 1,
									"description": "The 'mycollection_DataClients' AAD-Group members can access data from resources: blob/files from storage accounts, secrets from keyvaults, etc",
									"type": "info"
								},
								"elements": {
									"propertyOrder": 2,
									"title": " ",

									"type": "array",
									"options": {
										"collapsed": false, "disable_collapse": true
									},
									"format": "tabs",
						[opt]     "minItems": 1, "maxItems": 1,     << for 1mandatory element: minItems=maxItems=1
				O3					"items": {

	*/

	// Unmarshal the yaml into yMap
	var yMap map[string]interface{}
	{
		err = yaml.Unmarshal(yamlBytes, &yMap)
		if err != nil {
			return nil, err
		}
	}

	// jMap
	var jMap map[string]interface{}

	var JheaderTemplate, Jdescription string
	{
		YcatalogTitleAndDescriptionInMarkdown, ok := yMap["Catalog_title_and_description_in_markdown"].(string)
		if !ok {
			return nil, fmt.Errorf("error reading Catalog_title_and_description_in_markdown as string")
		}

		JheaderTemplate, Jdescription = parse_titleAndDescriptionInMarkdown(YcatalogTitleAndDescriptionInMarkdown)
	}
	jMap = map[string]interface{}{
		"headerTemplate": JheaderTemplate,
		"description":    Jdescription,
		"title":          " ",
		"type":           "object",
		"options": map[string]interface{}{
			"disable_collapse": true,
		},
		"properties": map[string]interface{}{},
	}

	// O1 level which sets jMap["properties"]
	var catalogProducts map[string]interface{}
	{
		if yMap["Catalog_products"] == nil {
			return nil, fmt.Errorf(`error "Catalog_products" is unexpectedly empty :/`)
		}
		var err error
		catalogProducts, err = mapIfcIfc_2_mapStrIfc(yMap["Catalog_products"].(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}
	}
	a_prod_counter := 0
	for a_prod_varname, a_prod_iface := range catalogProducts {
		// a_prod_counter     1,2,...   counts each product
		// a_prod_varname     "additional_members_of_AADGroup_DataClients"
		// a_prod_map         map[string]interface{}

		// assure a_prod_varname only contains valid-chars
		{
			patternValidChars := commonRegexpPatterns["alphanumUnderscoreHyphen"]
			match, err := regexp.MatchString(patternValidChars, a_prod_varname)
			if err != nil {
				return nil, err
			}
			if !match {
				return nil, fmt.Errorf("error, product-name '%s' does not match valid regexp '%s'", a_prod_varname, patternValidChars)
			}
		}
		a_prod_counter += 1
		a_prod_map, err := mapIfcIfc_2_mapStrIfc(a_prod_iface.(map[interface{}]interface{}))
		if err != nil {
			return nil, err
		}

		var JpropertyVarname, JheaderTemplate, Jdescription string
		var JpropertyOrder, JminItems, JmaxItems int
		var Jitems map[string]interface{}
		{
			var ok bool
			JpropertyVarname = a_prod_varname
			JpropertyOrder = 1000 + a_prod_counter

			var YproductTitleAndDescriptionInMarkdown string
			{
				YproductTitleAndDescriptionInMarkdown, ok = a_prod_map["Product_title_and_description_in_markdown"].(string)
				if !ok {
					return nil, fmt.Errorf("error reading Product_title_and_description_in_markdown")
				}
			}

			JheaderTemplate, Jdescription = parse_titleAndDescriptionInMarkdown(YproductTitleAndDescriptionInMarkdown)

			var YproductNrOfItems string
			{
				YproductNrOfItems, ok = a_prod_map["Product_nr_of_items"].(string)
				if !ok {
					YproductNrOfItems = "0-inf"
				}
			}

			JminItems, JmaxItems, err = convert_nrOfItems_to_minItems_maxItems(YproductNrOfItems)
			if err != nil {
				return nil, err
			}

			// Jitems
			{
				var YproductItem map[string]interface{}
				{
					mapIfcIfc, ok := a_prod_map["Product_item"].(map[interface{}]interface{})
					if !ok {
						return nil, fmt.Errorf(`Catalog_products["%s"].Product_items could not be read - aborting`, a_prod_varname)
					}
					YproductItem, err = mapIfcIfc_2_mapStrIfc(mapIfcIfc)
					if err != nil {
						return nil, err
					}
				}
				Jitems, err = yM2jM(YproductItem)
				if err != nil {
					return nil, err
				}
			}

		}

		jMapProperties := jMap["properties"].(map[string]interface{})
		jMapProperties[JpropertyVarname] = map[string]interface{}{
			"propertyOrder":  JpropertyOrder,
			"headerTemplate": JheaderTemplate,
			"title":          " ",
			"type":           "object",
			"options": map[string]interface{}{
				"collapsed": true,
			},
			"properties": map[string]interface{}{
				"info": map[string]interface{}{
					"propertyOrder": 1,
					"description":   Jdescription,
					"type":          "info",
					"title":         " ",
				},
				"elements": map[string]interface{}{
					"propertyOrder": 2,
					"title":         " ",
					"type":          "array",
					"options": map[string]interface{}{
						"collapsed": false, "disable_collapse": true,
					},
					"format":   "tabs",
					"minItems": JminItems, "maxItems": JmaxItems,
					"items": Jitems,
				},
			},
		}

	} // end: for a_prod_varname, a_prod_iface

	// marshall jMap into jsonBytes
	{
		jsonBytes, err = json.Marshal(jMap)
		if err != nil {
			return nil, err
		}
		jsonPrettyPrintString, err := jsonPrettyPrinter(string(jsonBytes))
		if err != nil {
			return nil, err
		}
		jsonBytes = []byte(jsonPrettyPrintString)
	}
	return jsonBytes, nil
}

// convert_nrOfItems_to_minItems_maxItems is a function that takes a string representation of a range (nrOfItems)
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
func convert_nrOfItems_to_minItems_maxItems(nrOfItems string) (minItemsInt int, maxItemsInt int, err error) {
	// Split the input string by the "-" character
	parts := strings.Split(nrOfItems, "-")

	// Check the number of parts after the split
	switch len(parts) {
	case 1:
		// If there is only one part, it means the input is a single number
		minItemsInt, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, errors.New("invalid input")
		}
		maxItemsInt = minItemsInt
		return minItemsInt, maxItemsInt, nil
	case 2:
		// If there are two parts, it means the input is a range
		minItemsInt, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, errors.New("invalid input")
		}
		// Check if the second part of the range is "inf"
		if parts[1] == "inf" {
			maxItemsInt = 1000
		} else {
			maxItemsInt, err = strconv.Atoi(parts[1])
			if err != nil {
				return 0, 0, errors.New("invalid input")
			}
		}
		return minItemsInt, maxItemsInt, nil
	default:
		// If there are more or less than 1 or 2 parts, it means the input is not in the correct format
		return 0, 0, errors.New("invalid input")
	}
}

func jsonPrettyPrinter(jsonIn string) (jsonOut string, err error) {
	if jsonIn == "" {
		return "", nil
	}
	var m map[string]interface{}
	err = json.Unmarshal([]byte(jsonIn), &m)
	if err != nil {
		return "", err
	}
	jsonOutBytes, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonOutBytes), nil
}

// parse_titleAndDescriptionInMarkdown function splits titleAndDescriptionMd into titleText and descriptionHtml
// The first line becomes the title, with its content copied as-is (left original text, not parsed as md)
// The third-and-remaining-lines become the description, with its content parsed from md to Html
func parse_titleAndDescriptionInMarkdown(titleAndDescriptionMd string) (titleText string, descriptionHtml string) {

	strSlice := strings.Split(titleAndDescriptionMd, "\n")

	// titleText
	{
		firstLineMd := ""
		if len(strSlice) > 0 {
			firstLineMd = strSlice[0]
		}

		// remove leading ^# if they exist
		titleMd := regexp.MustCompile(`^#+\s?`).ReplaceAllString(firstLineMd, "")
		titleText = titleMd
	}

	// descriptionHtml
	{
		// third-and-other-lines are taken as description
		ThirdAndOtherLinesMd := ""
		if len(strSlice) > 2 {
			ThirdAndOtherLinesMd = strings.Join(strSlice[2:], "\n")
		}
		descHtlm := markdown.ToHTML([]byte(ThirdAndOtherLinesMd), nil, nil)
		descriptionHtml = string(descHtlm)
	}

	return titleText, descriptionHtml
}

func mapIfcIfc_2_mapStrIfc(mapIfcIfc map[interface{}]interface{}) (mapStrIfc map[string]interface{}, err error) {
	mapStrIfc = make(map[string]interface{})
	for kIfc, vIfc := range mapIfcIfc {
		kStr, ok := kIfc.(string)
		if !ok {
			return nil, fmt.Errorf("string casting failed")
		}
		mapStrIfc[kStr] = vIfc
	}
	return
}

// yM2jM converts a yMap to a jMap with converted field-names for a json representation
func yM2jM(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	Ytype, ok := yMap["_type"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to read _type - aborting")
	}

	switch Ytype {
	case "string":
		return yM2jM_string(yMap)
	case "number":
		return yM2jM_number(yMap)
	case "integer":
		return yM2jM_integer(yMap)
	case "boolean":
		return yM2jM_boolean(yMap)
	case "object":
		return yM2jM_object(yMap)
	case "array":
		return yM2jM_array(yMap)
	default:
		return nil, fmt.Errorf("unrecognized _type: '%s' - aborting", Ytype)
	}
}

func yM2jM_object(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	// /*
	// 			   Expected input yaml structure (in a map)
	//
	// 					_type: object
	// 					_title_and_description_in_markdown: |-                 #[opt]
	// 						Sample obj to pick - {{ self.mypropX }}            #[opt] Raw-tex, no md; {{i0,i1,self,self.childprop...}} usefull when parent is array
	// 						--------------------
	// 						An sample object description                       # Use md in the description
	// 					_format: grid                                          #[opt] grid, prefer grid when content is small and fits well in same row
	// 					"mypropX":
	// 				    	<<generic type>>
	// 					"mypropY":
	// 						<<generic type>>
	//
	// 			   Produced output json (in a map)
	//
	// 				{
	// 	                "headerTemplate": "Sample obj to pick - {{ self.mypropX }}",   #[opt] Raw-tex, no HTML; {{i0,i1,self,self.childprop...}} usefull when parent is array
	// 	                "title": " ",                                                  #NOTE: "headerTemplate" replaces "title" (which is set to " " on purpose)
	// 	                "description": "A sample object description",
	// 	                "type": "object",
	// 	                "format": "grid",                                              #[opt] grid or null, prefer grid when content is small and fits well in same row
	// 	                "properties": {
	// 	                  "mypropX": {
	// 	                     <<generic type>>
	// 	                  },
	// 	                  "mypropY": {
	// 	                     <<generic type>>
	// 	                  },
	// 	                  ...
	// 	                }
	// 	            }
	// */

	// set default value for yMap["_format"]=nil (which will get mapped to json-null jMap["format"]=json-null)
	handle_format_defaultValue := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		if _, format_already_exists := yMap["_format"]; !format_already_exists {
			// yMap["_format"] does not exist
			// So lets set it to default value yMap["_format"]=nil which will get mapped to json-null
			yMap["_format"] = nil
		}
		return nil
	}

	// yMap["mypropX"|"mypropY"] -> jMap["properties"]["mypropX"|"mypropY"]
	handle_myprops := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		Jproperties := map[string]interface{}{}
		// YpropsKeys ex: ["mypropX", "mypropY"]
		var YpropsKeys []string
		{
			for k := range yMap {
				if regexp.MustCompile(`^[^_]`).MatchString(k) {
					*yHandledFieldsPtr = append(*yHandledFieldsPtr, k)
					YpropsKeys = append(YpropsKeys, k)
				}
			}
		}

		for _, a_YpropsKeys := range YpropsKeys {
			// a_YpropsKeys    "mypropX"
			// a_YpropsBlock   map[string]{interface}
			a_YpropsBlock, err := mapIfcIfc_2_mapStrIfc(yMap[a_YpropsKeys].(map[interface{}]interface{}))
			if err != nil {
				return err
			}
			a_JpropsBlock, err := yM2jM(a_YpropsBlock)
			if err != nil {
				return err
			}
			Jproperties[a_YpropsKeys] = a_JpropsBlock
		}
		jMap["properties"] = Jproperties
		return nil
	}

	return yM2jM_template_with_middlefuncs(yMap, handle_format_defaultValue, handle_myprops)
}

func yM2jM_string(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	return yM2jM_template_with_middlefuncs(yMap)
}

func yM2jM_number(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	return yM2jM_template_with_middlefuncs(yMap)
}

func yM2jM_integer(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	return yM2jM_template_with_middlefuncs(yMap)
}

func yM2jM_boolean(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	// set default value for yMap["_format"]="checkbox"
	handle_format_defaultValue := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		if _, format_already_exists := yMap["_format"]; !format_already_exists {
			// yMap["_format"] does not exist
			// So lets set it to default value yMap["_format"]="checkbox"
			yMap["_format"] = "checkbox"
		}
		return nil
	}

	return yM2jM_template_with_middlefuncs(yMap, handle_format_defaultValue)
}

func yM2jM_array(yMap map[string]interface{}) (jMap map[string]interface{}, err error) {
	// set default value for yMap["_options"]=map[string]interface{}{ "collapsed": false }
	handle_options_defaultValue := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		if _, options_already_exists := yMap["_options"]; !options_already_exists {
			// yMap["_options"] does not exist
			// So lets set it to default value yMap["_options"]=map[string]interface{}{ "collapsed": false }
			yMap["_options"] = map[string]interface{}{
				"collapsed": false,
			}
		}
		return nil
	}

	// yMap["_nr_of_items"] (default "0-inf") -> jMap["minItems"] and jMap["maxItems"]
	handle_nrOfItems := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		// if yMap["_nr_of_items"] is not set by user, then put default value yMap["_nr_of_items"]="0-inf"
		if _, nrOfItems_already_exists := yMap["_nr_of_items"]; !nrOfItems_already_exists {
			// yMap["_nr_of_items"] does not exist
			// So lets set it to default value yMap["_nr_of_items"]="0-inf"
			yMap["_nr_of_items"] = "0-inf"
		}
		// ATP: yMap["_nr_of_items"] exists with a value (either default or user-set)

		// calculate JminItems, JmaxItems and set jMap["minItems"], jMap["maxItems"]
		YnrOfItems := yMap["_nr_of_items"].(string)
		JminItems, JmaxItems, err := convert_nrOfItems_to_minItems_maxItems(YnrOfItems)
		if err != nil {
			return err
		}
		jMap["minItems"] = JminItems
		jMap["maxItems"] = JmaxItems

		// yHandledFieldsPtr appended with "_nr_of_items"
		*yHandledFieldsPtr = append(*yHandledFieldsPtr, "_nr_of_items")
		return nil
	}

	// yMap["_items"] -> jMap["items"]
	handle_items := func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
		// NOTE: yItems represents one single map, and not a slice-of-maps. The plural name yItem*s* should be read as a single map from yMap["_items"] into jMap["items"]
		if yItemsIfcIfc, yItems_exists := yMap["_items"].(map[interface{}]interface{}); yItems_exists {
			yItems, err := mapIfcIfc_2_mapStrIfc(yItemsIfcIfc)
			if err != nil {
				return err
			}
			jMap["items"], err = yM2jM(yItems)
			if err != nil {
				return err
			}

			// yHandledFieldsPtr appended with "_items"
			*yHandledFieldsPtr = append(*yHandledFieldsPtr, "_items")
		} else {
			return fmt.Errorf("array missing mandatory key '_items'")
		}

		return nil
	}

	return yM2jM_template_with_middlefuncs(yMap, handle_options_defaultValue, handle_nrOfItems, handle_items)
}

type middleFunc func(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error)

// yM2jM_template_with_middlefuncs will in this order:
//   - yMap["_type"] -> jMap["type"]
//   - yMap "_title_and_description_in_markdown" -> jMap "headerTemplate" "title" "description"
//   - execute all middleFuncs (if any)
//   - all remaining yUnderscoreUnhandledFields yMap["_zzzz"] not present in yHandledFields -> jMap["zzzz"]
//
// A middleFunc that processes a yMap["yfield"] into an jMap["jfield"], should take care of appending the string "yfield" into *yHandledFieldsPtr
// to assure yMap["yfield"] does not get re-processed afterwards
func yM2jM_template_with_middlefuncs(yMap map[string]interface{}, middleFuncs ...middleFunc) (jMap map[string]interface{}, err error) {

	jMap = map[string]interface{}{}
	// yHandledFields is list of _yamlfields that are already handled and should be skipped in further automatic-processing (like JfieldsAutomappedFromYunderscoreFields())
	yHandledFields := []string{}
	yHandledFieldsPtr := &yHandledFields

	// yMap["_type"] -> jMap["type"]
	setJmap_type(jMap, yMap, &yHandledFields)

	// yMap "_title_and_description_in_markdown" -> jMap "headerTemplate" "title" "description"
	setJmap_headerTemplate_title_description(jMap, yMap, &yHandledFields)

	// execute middleFuncs
	for _, aMiddleFunc := range middleFuncs {
		err = aMiddleFunc(jMap, yMap, yHandledFieldsPtr)
		if err != nil {
			return nil, err
		}
	}

	// all remaining yUnderscoreUnhandledFields yMap["_zzzz"] not present in yHandledFields -> jMap["zzzz"]
	{
		err := setJmap_yUnderscodeUnhandledFields(jMap, yMap, &yHandledFields)
		if err != nil {
			return nil, err
		}
	}

	return jMap, nil
}

// setJmap_type will :
//   - read yMap["_type"] and set jMap["type"]
//   - append into yHandledFields "_type"
//
// Assumes yMap["type"] exists (must exist!)
func setJmap_type(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) {
	yMap_type := yMap["_type"]
	*yHandledFieldsPtr = append(*yHandledFieldsPtr, "_type")
	jMap["type"] = yMap_type
}

// setJmap_headerTemplate_title_description will:
//   - read yMap["_title_and_description_in_markdown"]
//   - set jMap "headerTemplate", "title", "description"
//   - append into yHandledFields "_title_and_description_in_markdown"
func setJmap_headerTemplate_title_description(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) {
	var JheaderTemplate, Jtitle, Jdescription string
	{
		yFieldName := "_title_and_description_in_markdown"
		*yHandledFieldsPtr = append(*yHandledFieldsPtr, yFieldName)
		YtitleAndDescriptionInMarkdown := ""
		if _, ok := yMap[yFieldName].(string); ok {
			YtitleAndDescriptionInMarkdown = yMap[yFieldName].(string)
		}
		JheaderTemplate, Jdescription = parse_titleAndDescriptionInMarkdown(YtitleAndDescriptionInMarkdown)
		Jtitle = " "
	}

	jMap["headerTemplate"] = JheaderTemplate
	jMap["title"] = Jtitle
	jMap["description"] = Jdescription
}

// setJmap_yUnderscodeUnhandledFields will do:
// yMap["_zzzz"] not present in yHandledFields -> jMap["zzzz"]
func setJmap_yUnderscodeUnhandledFields(jMap map[string]interface{}, yMap map[string]interface{}, yHandledFieldsPtr *[]string) (err error) {
	for yk, yv := range yMap {
		yk_is_zzzz := regexp.MustCompile(`^_`).MatchString(yk)
		if yk_is_zzzz {
			// yk is _zzzz
			yk_not_present_in_yHandledFields := (slices.IndexFunc(*yHandledFieldsPtr, func(yhk string) bool { return yhk == yk }) == -1)
			if yk_not_present_in_yHandledFields {
				// yk is not present in yHandledFields

				// yk    "_zzzz"
				// jk    "zzzz"
				jk := regexp.MustCompile("^_").ReplaceAllString(yk, "")
				jv := yv
				jMap[jk] = jv
			}
		}
	}
	return
}
