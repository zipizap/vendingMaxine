package collection

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

// TODO
func CollectionNew(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// create collection dir
	// create first valid rsf
	// create anything else necessary for the collection to be valid and used from now on
	// NOTE: once a collection-dir exists, its assumed everything needed for the collection to properly work is ready too

	// At the end of this function, everything necessary for the collection to work must be ready
}

// returns error if collection dirpath does not exist
func CollectionGet(cname string) (col *Collection, err error) {
	c := &Collection{Name: cname}
	if _, err := os.Stat(c.dirpath()); os.IsNotExist(err) {
		// c.dirpath() does not exist => collection does not exist
		return nil, fmt.Errorf("collection '%s' does not exist", cname)
	}
	return c, nil
}

// Ex:
//
//    allCols, err := CollectionsAllGet()
//
func CollectionsAllGet(optional_collectionName_regexp ...string) (cols []*Collection, err error) {
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
		a_col, err := CollectionGet(a_colName)
		if err != nil {
			return nil, err
		}
		cols = append(cols, a_col)
	}
	return cols, nil
}

// TODO
func CollectionDelete(cname string) (col *Collection, err error) {
	log.Fatal("TODO: CODE ME")
	return nil, nil
	// delete collection dir, just by doing this shold be enough
}

// In all collections, in async paralel go-routines, calls
//	 a_col.NewRsf_for_ProcEngAssembly(data map[string]string)
//      and if they cantDo then retry every 5s (infinite loop)
// 		and if they error-out, return that error
//		and if all launches were ok, then return quickly
//
// This function launches go-routines (that will run async in paralell), and quickly returns from itself.
// Ie, this function returns quickly, does not wait for runners execution to complete
//
func CollectionsAllAssembly() (err error) {
	// This function should:
	// a) get allCols []*Collections
	// b) for each a_col, launch in paralell go-routines
	//	b.1) cantDo, rsf, err := a_col.NewRsf_for_ProcEngAssembly(data map[string]string)
	//			cantDo => retry after 5s
	//			err => log error and return
	//			else => NewRsf_for_ProcEngAssembly went well, so just return
	// c) return
	//
	// This function launches go-routines (that will run async in paralell), and quickly returns from itself.
	// Ie, this function returns quickly, does not wait for runner execution
	//

	// a) get allCols []*Collections
	allCols, err := CollectionsAllGet()
	if err != nil {
		return err
	}

	// b) for each a_col, launch in paralell go-routines
	//	b.1) cantDo, rsf, err := a_col.NewRsf_for_ProcEngAssembly(data map[string]string)
	//			cantDo => retry after 5s
	//			err => log error and return
	//			else => NewRsf_for_ProcEngAssembly went well, so just return
	for _, a_col := range allCols {
		go func(a_col *Collection) {
			for {

				data := make(map[string]string)
				cantDo, _, err := a_col.NewRsf_for_ProcEngAssembly(data)
				if err != nil {
					log.Error(err)
					return
				} else if cantDo {
					// retry after 5s (continue in infinite-for-loop)
					time.Sleep(5 * time.Second)
					continue
				} else {
					return
				} // if
			} // while
		}(a_col) // go func()
	} // for

	// c) return
	return
}
