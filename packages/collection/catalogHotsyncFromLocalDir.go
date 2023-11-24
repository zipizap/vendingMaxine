package collection

import (
	"io"
	"os"
	"path/filepath"
)

// NOTE: please understand that this had to be done as quickly as possible ;)
// If there was more time, there would certainly be a better way to it, like using interfaces or something ;)

var catalogHotsyncFromLocalDir_Dirpath string

// set_catalogHotsyncFromLocalDir_Dirpath function is just a facade to the var set_catalogHotsyncFromLocalDir_Dirpath
// to make it easier to separate-and-track which other files called this func. If in the future this is removed
func set_catalogHotsyncFromLocalDir_Dirpath(val string) {
	catalogHotsyncFromLocalDir_Dirpath = val
}

// catalogHotsyncFromLocalDir_catalogDir function is a ?temporary? trick to always hot-read and return the local-dir of default catalog
// It will return the <catalogDefaultDirpath> that was passed to initCatalog()
// It was done as an alternative to ca.catalogDir(), to always-hot-reload from local-dir (instead of using saved-static-db-data) so that changes in local-dir are returned inmediately
// This was done to speed-up iterative debugging of catalog files ;)
func catalogHotsyncFromLocalDir_catalogDir() (catalogDirPath string, err error) {
	return catalogHotsyncFromLocalDir_Dirpath, nil
}

// catalogHotsyncFromLocalDir_schema function is a trick to replace ca.schema(), so it hot-reloads from local-dir-files
func catalogHotsyncFromLocalDir_schema() (schemaJson string, err error) {
	catalogDirPath := catalogHotsyncFromLocalDir_Dirpath
	fileInCatalogRelativeFilepath := "Schema.json"
	var fileInCatalogContent []byte

	{
		fullFilepath := filepath.Join(catalogDirPath, fileInCatalogRelativeFilepath)
		file, err := os.Open(fullFilepath)
		if err != nil {
			return "", err
		}
		defer file.Close()

		fileInCatalogContent, err = io.ReadAll(file)
		if err != nil {
			return "", err
		}
	}
	schemaJson = string(fileInCatalogContent)

	return schemaJson, nil
}
