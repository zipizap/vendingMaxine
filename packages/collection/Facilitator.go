package collection

import (
	"fmt"

	"go.uber.org/zap"
)

// Use the Facilitator functions, and avoid messing with anything else inside the package :)
//
// Facilitator is just a dummy type, a placeholder of functions to be called by the API :)
//
// f.InitSetup() function must be called first, before calling any other functions of this package
type Facilitator struct{}

func NewFacilitator() (*Facilitator, error) {
	return &Facilitator{}, nil
}

// InitSetup function must be called before using other functions of this package
func (f *Facilitator) InitSetup(dbFilepath string, catalogDefaultName string, catalogDefaultDirpath string, zapSugaredLogger *zap.SugaredLogger) {
	initSlog(zapSugaredLogger)
	initDb(dbFilepath)
	initCatalog(catalogDefaultName, catalogDefaultDirpath)
}

// Create a new collection
//
//	`name` must be compliant with DNS label standard as defined in RFC 1123 (like pod label-names, see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names).
func (f *Facilitator) CollectionNew(collectionName, catalogName string) error {
	_, err := collectionNew(collectionName, catalogName)
	return err
}

// CollectionEdit_Prepinfo returns the preparatory-info necessary to start editing a collection
func (f *Facilitator) CollectionEdit_Prepinfo(colName string) (schemaJson string, jsonInput string, err error) {
	allowCatalogRenewal := false
	return f._collectionEdit_Prepinfo(colName, allowCatalogRenewal)
}
func (f *Facilitator) _collectionEdit_Prepinfo(colName string, allowCatalogRenewal bool) (schemaJson string, jsonInput string, err error) {
	// load and validate col can be updated
	var col *Collection
	{
		col, err = collectionLoad(colName)
		if err != nil {
			return "", "", err
		}
		if err = col.canBeUpdated(); err != nil {
			return "", "", err
		}
	}

	// get schemaJson
	schemaJson, err = col.Catalog.schema()
	if err != nil {
		return "", "", err
	}

	// get jsonInput (jsonInput = cselLatest.JsonOutput)
	{
		cselLatest, err := col.colSelectionLatest()
		if err != nil {
			return "", "", err
		}

		// thisEdit::jsonInput = formerEdit::jsonOutput
		jsonInput = cselLatest.JsonOutput

		if !allowCatalogRenewal {
			// Safety validation: we dont accept catalog-renewal between cselLatest and now
			if cselLatest.CatalogID != col.CatalogID {
				return "", "", fmt.Errorf("safety-protection, a catalog-renewal is not allowed between last edit and now")
			}
		}
	}

	return schemaJson, jsonInput, nil
}

// CollectionEdit_Save updates the collection
func (f *Facilitator) CollectionEdit_Save(colName string, schemaJson string, jsonInput string, jsonOutput string, requestingUser string) error {
	// load and validate col can be updated
	var col *Collection
	{
		var err error
		col, err = collectionLoad(colName)
		if err != nil {
			return err
		}
		if err = col.canBeUpdated(); err != nil {
			return err
		}
	}

	// Assure schemaJson == catalogSchema, ie, since CollectionEdit_Prepinfo:
	// - the catalog schema has not changed (no catalog renewal happened in the meanwhile)
	// - the user has not changed the jsonSchema
	var catalogSchema string
	{
		var err error
		catalogSchema, err = col.Catalog.schema()
		if err != nil {
			return err
		}
		if schemaJson != catalogSchema {
			return fmt.Errorf("schemaJson != catalogSchema which is unexpected. From  collection-edit-start to collection-edit-save a change of jsonSchema is not supported. Aborting collection-edit-save")
		}
	}

	// Append and Run ColSelection
	err := col.appendAndRunColSelection(jsonInput, jsonOutput, requestingUser)
	if err != nil {
		return err
	}

	return nil
}

// CollectionsOverview returns list of maps with usefull info of all collections
//
//	colsInfo, err := f.CollectionsOverview()
//	for _, a_colInfo := range colsInfo {
//	  fmt.Println("Collection Name: " , a_colInfo.Name)
//	  fmt.Println("Collection CatalogName: " , a_colInfo.CatalogName)
//	  fmt.Println("Collection State: ", a_colInfo.State)
//	  fmt.Println("Collection ErrorStr: ", a_colInfo.ErrorStr)
//	}
func (f *Facilitator) CollectionsOverview() (colsInfo []CollectionInfo, err error) {
	return collectionsOverview()
}

// CollectionReplayable returns the replayableTgz for the colName collection
func (f *Facilitator) CollectionReplayable(colName string) (collectionReplayableTgz []byte, collectionReplayableTgzBasename string, err error) {
	col, err := collectionLoad(colName)
	if err != nil {
		return nil, "", err
	}
	return col.getCollectionReplayableDirTgz()
}

// CatalogsOverview returns list of maps with usefull info of all Catalogs
//
//	catsInfo, err := f.CatalogsOverview()
//	for _, a_catInfo := range catsInfo {
//	  fmt.Println("Catalog Name: " , a_catInfo.Name)
//	  fmt.Println("Catalog Deprecated?: " , a_catInfo.Deprecated)
//	}
func (f *Facilitator) CatalogsOverview() (colsInfo []CatalogInfo, err error) {
	return catalogsOverview()
}
