package collection

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type Collection struct {
	Name string // Name is always defined (since Collection creation)
}

var collectionBaseDir = filepath.Join("vd-internal", "collections")

// todo: address case when there is no last-rsf file
func (c *Collection) LastRsf() (rsf *RequestStatusFlow, err error) {
	rsf = &RequestStatusFlow{Collection: c.Name}
	err = rsf.i1_syncLoadFromLastFile()
	if err != nil {
		return nil, err
	}
	return rsf, nil
}

// If NewRsf can be created: yes == true
func (c *Collection) NewRsf_canBeCreated() (yes bool, err error) {
	// NewRsf can be created if
	//    a) LastRsf does not exist
	//  or
	//    b) LastRsf has .Status.Overall.LatestUpdateStatus ~= Error|Completed

	// Lets check a)
	lastFiles, err := c.lastFiles()
	if err != nil {
		return false, err
	}
	if len(lastFiles) == 0 {
		// LastRst does not exist
		return true, nil
	}

	// Lets check b)
	lastRsf, err := c.LastRsf()
	if err != nil {
		return false, err
	}
	matchFound, err := regexp.MatchString("Completed|Error", lastRsf.Status.Overall.LatestUpdateStatus)
	if err != nil {
		return false, err
	}
	if matchFound {
		return true, nil
	} else {
		return false, nil
	}
}

// Tries to create a new RequestStatusFlow (rsf) from webdata
//
// webdata must have following keys-values:
//		"products.schema.json":             []byte of the json (not b64gz, but raw json bytes)
//		"consumer-selection.previous.json": []byte of the json (not b64gz, but raw json bytes)
//		"consumer-selection.next.json":     []byte of the json (not b64gz, but raw json bytes)
//
// An implicit part of rsf creation, it that  rsf.runProcessingEngines() (async) will be launched asynchronously
// (but it will not wait for runProcessingEngines() to complete, that is left running async)
// If it cant do it =>  returns "cantDo == true"
func (c *Collection) NewRsf_from_WebconsumerSelection(webdata map[string]string) (cantDo bool, rsf *RequestStatusFlow, err error) {
	rsf = &RequestStatusFlow{
		Collection: c.Name,
	}
	t_now := time.Now()
	productsSchemaJson_gzB64, err := encode_bytes_to_gzB64([]byte(webdata["products.schema.json"]))
	if err != nil {
		return false, nil, err
	}
	consumerSelectionPreviousJson_gzB64, err := encode_bytes_to_gzB64([]byte(webdata["consumer-selection.previous.json"]))
	if err != nil {
		return false, nil, err
	}
	consumerSelectionNextJson_gzB64, err := encode_bytes_to_gzB64([]byte(webdata["consumer-selection.next.json"]))
	if err != nil {
		return false, nil, err
	}
	web_sblock := StatusBlock{
		Name:                   "WebConsumerSelection",
		StartTime:              t_now,
		LatestUpdateTime:       t_now,
		LatestUpdateStatus:     "Completed",
		LatestUpdateStatusInfo: "",
		LatestUpdateUml:        "",
		LatestUpdateData: map[string]interface{}{
			"products.schema.json":             productsSchemaJson_gzB64,
			"consumer-selection.previous.json": consumerSelectionPreviousJson_gzB64,
			"consumer-selection.next.json":     consumerSelectionNextJson_gzB64,
		},
	}
	cantDo, err = rsf.new_from_webConsumerSelection(web_sblock)
	if err != nil {
		return true, nil, err
	}
	if cantDo {
		return true, nil, nil
	}
	return false, rsf, nil
}

func (c *Collection) NewRsf_2start_WebconsumerSelection(collectionName string) (cantDo bool, consumerSelectionPreviousJson_string string, productsSchemaJson_string string, err error) {
	// This function will:
	//   a) Check if theCollection_name exists (and return error if not)
	//	 b) Check if theCollection.NewRsf_canBeCreated
	//   c) Read consumerSelectionPreviousJson_string from theCollection.LastRsf
	//   d) Read productsSchemaJson_string from file PRODUCT_SCHEMA_JSON_FILEPATH
	// 	 e) create newRsf with remaining fields, and Status.Overall.
	//			LatestUpdateStatus: "Ongoing_and_locked"
	//			LatestUpdateData:
	//				"consumer-selection.previous.json": <gz.b64>  >> copied from previous lastRsf.Status.Overall["consumer-selection.next.json"]
	//
	//	 f.1) create empty-file + syncSave (like Collection.go:LOC148-169)
	//	 f.2) call rsf.i1_runProcessingEngines("assemble",false) (in this same goroutine! blocking!)
	//				TODO: i1_runProcessingEngines needs improvement for arg "assemble|execute"
	//
	//   g) now that all processingengines-assembly were executed, return
	************************
	***** continue here ****
	rethink this: a get request would leave the rsf in "ongoing_and_blocked" indefinitely... 

	//
	// 		NOTE: if last_rsf does not exist (ex: new collection) then err != nil
	// 		and new collectino will never work
	// 		todo: a newly-created-collection might never work as it might not have a last_rsf for initial bootstraping

}

// Calculates path-to-dir of collection.
// That colDirpath might or not exist yet in fs! Its just calculated (not created or checked)
func (c *Collection) dirpath() (colDirpath string) {
	return filepath.Join(collectionBaseDir, c.Name)
}

// Returns []string of rel-path-and-filenames where first is newest-by-filename and last is oldest-by-filename
// The optional_filename_regexp filters filenames (without path). If any regexp matches, the file is included
// The optional_filename_regexp defaults to "RequestStatusFlow.*.yaml"
// If dirpath is not readable, or is empty, returns empty slice []string
//
//  fls, err := c.lastFiles()
//  fls, err := c.lastFiles("my-file*.*", "another-regexp$")
func (c *Collection) lastFiles(optional_filename_regexp ...string) (lastFilesRecentFirst []string, err error) {
	// The optional_filename_regexp defaults to "RequestStatusFlow.*.yaml"
	if len(optional_filename_regexp) == 0 {
		optional_filename_regexp = append(optional_filename_regexp, "^RequestStatusFlow.*.yaml$")
	}
	dirpath := c.dirpath()
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		// dirpath does not exist, return empty slice
		lastFilesRecentFirst = []string{}
		return lastFilesRecentFirst, nil
	}
	fileInfo, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, err
	}
	sort.Slice(fileInfo, func(i, j int) bool {
		return fileInfo[i].Name() > fileInfo[j].Name()
	})
	for _, file := range fileInfo {
		// skip if it's a dir
		if file.IsDir() {
			continue
		}
		if len(optional_filename_regexp) > 0 {
			// skip if file.Name() does not match any filename_regexp
			include_this_file := false
			for _, a_filename_regexp := range optional_filename_regexp {
				matchTrue, err := regexp.MatchString(a_filename_regexp, file.Name())
				if err != nil {
					return nil, err
				}
				if matchTrue {
					include_this_file = true
				}
			}
			if !include_this_file {
				continue
			}
		}
		lastFilesRecentFirst = append(lastFilesRecentFirst, filepath.Join(dirpath, file.Name()))
	}
	return lastFilesRecentFirst, nil
}
