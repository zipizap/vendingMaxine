package collection

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
)

// TODO
func NewCollection(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// create collection dir
	// create first valid rsf
	// create anything else necessary for the collection to be valid and used from now on
	// NOTE: once a collection-dir exists, its assumed everything needed for the collection to properly work is ready too

	// At the end of this function, everything necessary for the collection to work must be ready
}

// returns error if collection dirpath does not exist
func GetCollection(cname string) (col *Collection, err error) {
	c := &Collection{Name: cname}
	if _, err := os.Stat(c.dirpath()); os.IsNotExist(err) {
		// c.dirpath() does not exist => collection does not exist
		return nil, fmt.Errorf("collection '%s' does not exist", cname)
	}
	return c, nil
}

// Ex:
//
//    allCols, err := GetAllCollections()
//
func GetAllCollections(optional_collectionName_regexp ...string) (cols []*Collection, err error) {
	// The optional_collectionName_regexp defaults to "*"
	var colNames []string
	if len(optional_collectionName_regexp) == 0 {
		optional_collectionName_regexp = append(optional_collectionName_regexp, ".*")
	}
	if _, err := os.Stat(collectionBaseDir); os.IsNotExist(err) {
		// collectionBaseDir does not exist ?!?!
		return nil, fmt.Errorf("wut? could not read directory collectionBaseDir='%s'", collectionBaseDir)
	}
	fileInfo, err := ioutil.ReadDir(collectionBaseDir)
	if err != nil {
		return nil, err
	}
	sort.Slice(fileInfo, func(i, j int) bool {
		return fileInfo[i].Name() < fileInfo[j].Name()
	})
	for _, fileinfo := range fileInfo {
		if !fileinfo.IsDir() {
			// it's not a dir, skip
			continue
		}

		// skip if fileinfo.Name() does not match any a_colname_regexp
		include_this_fileinfo := false
		for _, a_colname_regexp := range optional_collectionName_regexp {
			matchTrue, err := regexp.MatchString(a_colname_regexp, fileinfo.Name())
			if err != nil {
				return nil, err
			}
			if matchTrue {
				include_this_fileinfo = true
			}
		}
		if !include_this_fileinfo {
			continue
		}
		colNames = append(colNames, fileinfo.Name())
	}
	for _, a_colName := range colNames {
		a_col, err := GetCollection(a_colName)
		if err != nil {
			return nil, err
		}
		cols = append(cols, a_col)
	}
	return cols, nil
}

// TODO
func DeleteCollection(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// delete collection dir, just by doing this shold be enough
}
