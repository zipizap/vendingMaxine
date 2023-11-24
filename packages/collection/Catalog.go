package collection

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

var cachedCatalogDirs = map[string]string{}

type CatalogInfo struct {
	Name       string `json:"name"`
	Deprecated bool   `json:"deprecated"`
}

type Catalog struct {
	gorm.Model

	Name                string `gorm:"unique,uniqueIndex,not null"`
	CatalogDirTgzBlobId uint
	Deprecated          bool

	dbMethods
}

func initCatalog(catalogDefaultName string, catalogDefaultDirpath string) {
	set_catalogHotsyncFromLocalDir_Dirpath(catalogDefaultDirpath)

	_, err := catalogLoad(catalogDefaultName)
	if err == gorm.ErrRecordNotFound {
		// default category does not exist yet, lets create it
		_, err = _catalogCreateInitial(catalogDefaultName, catalogDefaultDirpath)
		if err != nil {
			slog.Fatalf("error creating default category %v", err)
		}
	} else if err != nil {
		slog.Fatalf("Unexpected error %v", err)
	}
}

func catalogNew(catalogName string, catalogDirTgzData []byte) (*Catalog, error) {

	// validate name
	{
		if err := _isValidDNSLabel(catalogName); err != nil {
			return nil, err
		}
	}

	// if Catalog already exists, return error
	{
		if _, err := catalogLoad(catalogName); err == nil {
			return nil, fmt.Errorf("Catalog %v already exists", catalogName)
		}
	}

	// improvement: validate catalogDirTgzData

	// preArrangements on catalogDirTgzData (catalogDirTgzData might be changed)
	{
		err := catalogPreArrangements(&catalogDirTgzData)
		if err != nil {
			return nil, err
		}
	}
	// Create and set new object
	o := &Catalog{}
	{
		o.Name = catalogName

		// o.CatalogDirTgzBlobId
		{
			catalogDirTgzBlob, err := blobNew(catalogDirTgzData)
			if err != nil {
				return nil, err
			}
			catalogDirTgzBlobId := catalogDirTgzBlob.ID
			o.CatalogDirTgzBlobId = catalogDirTgzBlobId
		}
	}

	// save object to db
	{
		err := o.save(o)
		if err != nil {
			return nil, err
		}
	}

	// return object
	return o, nil
}

func catalogNewFromLocaldir(catalogName string, localDirPath string) (*Catalog, error) {
	// define catalogDirTgz from localDirPath
	var catalogDirTgz []byte
	{
		var err error
		catalogDirTgz, err = compressDir2Tgz(localDirPath)
		if err != nil {
			return nil, err
		}
	}
	return catalogNew(catalogName, catalogDirTgz)
}

// func catalogNewFromGitref(catalogName string, gitref string) (*Catalog, error) {
// }

// catalogsOverview returns list of maps with usefull info of all Catalogs
//
//	catsInfo, err := f.CatalogsOverview()
//	for _, a_catInfo := range catsInfo {
//	  fmt.Println("Catalog Name: " , a_catInfo.Name)
//	  fmt.Println("Catalog Deprecated?: " , a_catInfo.Deprecated)
//	}
func catalogsOverview() (catsInfo []CatalogInfo, err error) {
	catList := []*Catalog{}
	err = db.Find(&catList).Error
	if err != nil {
		return nil, err
	}
	for _, cat := range catList {
		catsInfo = append(catsInfo, CatalogInfo{
			Name:       cat.Name,
			Deprecated: cat.Deprecated,
		})
	}
	if catsInfo == nil {
		catsInfo = []CatalogInfo{}
	}
	return catsInfo, nil
}

// catalogLoad loads from db
func catalogLoad(name string) (*Catalog, error) {
	if err := _isValidDNSLabel(name); err != nil {
		return nil, err
	}

	o := &Catalog{}
	// The following db.Where... will not do nested-preloading, that will be done latter
	// with the o.reload(o) call
	err := db.Where("name = ?", name).First(o).Error
	if err != nil {
		return nil, err
	}
	err = o.reload(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// +++ changed for temporary quick debug
func (ca *Catalog) catalogDir() (catalogDirPath string, err error) {
	// temporary quick debugging
	return catalogHotsyncFromLocalDir_catalogDir()

	// reload from db
	{
		err := ca.reload(ca)
		if err != nil {
			return "", err
		}
	}

	// return from cache if exists
	{
		if cachedCatalogDir, ok := cachedCatalogDirs[ca.Name]; ok {
			return cachedCatalogDir, nil
		}
	}

	// create temp-dir cachedCatalogDir
	var cachedCatalogDir string
	{
		cachedCatalogDir, err = os.MkdirTemp("", ca.getCatalogDirBasenameString())
		if err != nil {
			return "", err
		}
	}

	// extract ca.CatalogDirTgz into cachedCatalogDir
	{
		catalogDirTgzData, err := ca.getCatalogDirTgz()
		if err != nil {
			return "", err
		}
		err = extractTgz2Dir(catalogDirTgzData, cachedCatalogDir)
		if err != nil {
			return "", err
		}
	}

	// add to cache
	cachedCatalogDirs[ca.Name] = cachedCatalogDir

	// return
	return cachedCatalogDir, nil
}

// "catalog-my-catalog"
func (ca *Catalog) getCatalogDirBasenameString() string {
	return "catalog-" + ca.Name
}

// +++ changed for temporary quick debug
func (ca *Catalog) schema() (schemaJson string, err error) {
	// temp quick debug
	return catalogHotsyncFromLocalDir_schema()

	schemaJsonContent, err := ca.readFile("Schema.json")
	if err != nil {
		return "", err
	}
	schemaJson = string(schemaJsonContent)
	return schemaJson, nil
}

// schemaYamlContent, err := ca.readFile("Schema.yaml")
func (ca *Catalog) readFile(fileInCatalogRelativeFilepath string) (fileInCatalogContent []byte, err error) {
	// reload from db
	{
		err = ca.reload(ca)
		if err != nil {
			return fileInCatalogContent, err
		}
	}

	// get catalogDirPath
	var catalogDirPath string
	{
		catalogDirPath, err = ca.catalogDir()
		if err != nil {
			return fileInCatalogContent, err
		}
	}

	// read schemaJson from catalogDirPath/Schema.json
	{
		fullFilepath := filepath.Join(catalogDirPath, fileInCatalogRelativeFilepath)
		file, err := os.Open(fullFilepath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fileInCatalogContent, err = io.ReadAll(file)
		if err != nil {
			return nil, err
		}
	}
	return fileInCatalogContent, nil
}

func (ca *Catalog) deprecate(deprecateBool bool) error {
	// reload from db
	{
		err := ca.reload(ca)
		if err != nil {
			return err
		}
	}
	ca.Deprecated = deprecateBool
	return nil
}

func (ca *Catalog) getCatalogDirTgz() (catalogDirTgzData []byte, err error) {
	return blobData(ca.CatalogDirTgzBlobId)
}

func (o *Catalog) gormID() uint {
	return o.ID
}

func _catalogCreateInitial(catalogDefaultName string, catalogDefaultDirpath string) (*Catalog, error) {
	var catDefault *Catalog
	{
		var err error
		catDefault, err = catalogNewFromLocaldir(catalogDefaultName, catalogDefaultDirpath)
		if err != nil {
			return nil, err
		}
	}
	return catDefault, nil
}
