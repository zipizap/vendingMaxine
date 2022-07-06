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

// todo: address case when there is no last-rsf file
func (c *Collection) LastRsf() (rsf *RequestStatusFlow, err error) {
	rsf = &RequestStatusFlow{Collection: c.Name}
	err = rsf.i1_syncLoadFromLastFile()
	if err != nil {
		return nil, err
	}
	return rsf, nil
}

// Tries to create a new RequestStatusFlow (rsf) from webdata
// webdata must have following keys-values:
//		"products.schema.json":             []byte of the json
//		"consumer-selection.previous.json": []byte of the json
//		"consumer-selection.next.json":     []byte of the json
//
// An implicit part of rsf creation, it that  rsf.runProcessingEngines() (async) will be launched asynchronously
// (but it will not wait for runProcessingEngines() to complete, that is left running async)
// If it cant do it =>  returns "cantDo == true"
func (c *Collection) NewRsf_from_WebconsumerSelection(webdata map[string]interface{}) (cantDo bool, rsf *RequestStatusFlow, err error) {
	rsf = &RequestStatusFlow{
		Collection: c.Name,
	}
	t_now := time.Now()
	web_sblock := StatusBlock{
		Name:                   "WebConsumerSelection",
		StartTime:              t_now,
		LatestUpdateTime:       t_now,
		LatestUpdateStatus:     "Completed",
		LatestUpdateStatusInfo: "",
		LatestUpdateUml:        "",
		LatestUpdateData: map[string]interface{}{
			"products.schema.json":             webdata["products.schema.json"],
			"consumer-selection.previous.json": webdata["consumer-selection.previous.json"],
			"consumer-selection.next.json":     webdata["consumer-selection.next.json"],
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

// Calculates path-to-dir of collection.
// Dirpath might or not exist yet in fs! Its just calculated (not created or checked)
func (c *Collection) dirpath() (colDirpath string) {
	return filepath.Join("vd-internal", "collections", c.Name)
}

// Returns []string of rel-path-and-filenames where first is newest-modified-file and last is oldest-modified-file
// The optional_filename_regexp filters filenames (without path). If any regexp matches, the file is included
// The optional_filename_regexp defaults to "RequestStatusFlow.*.yaml"
// If dirpath is not readable, or is empty, returns empty slice []string
//
//  fls, err := c.lastFiles()
//  fls, err := c.lastFiles("my-file*.*", "another-regexp$")
func (c *Collection) lastFiles(optional_filename_regexp ...string) (lastFilesRecentFirst []string, err error) {
	// The optional_filename_regexp defaults to "RequestStatusFlow.*.yaml"
	if len(optional_filename_regexp) == 0 {
		optional_filename_regexp = append(optional_filename_regexp, "RequestStatusFlow.*.yaml")
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
		t1 := fileInfo[i].ModTime()
		t2 := fileInfo[j].ModTime()
		//return t1.Before(t2)
		return t1.After(t2)
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
