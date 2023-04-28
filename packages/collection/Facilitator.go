package collection

import "fmt"

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
func (f *Facilitator) InitSetup(dbFilepath string, processingEnginesDirpath string) {
	initDb(dbFilepath)
	initProcessingEngineRunner(processingEnginesDirpath)
}

// CollectionsOverview returns list of maps with usefull info of all collections
//
//	colsInfo, err := f.CollectionsOverview()
//	for _, a_colInfo := range colsInfo {
//	  fmt.Println("Collection Name: " , a_colInfo["Name"])
//	  fmt.Println("Collection State: ", a_colInfo["State"])
//	  fmt.Println("Collection ErrorStr: ", a_colInfo["ErrorStr"])
//	}
func (f *Facilitator) CollectionsOverview() (colsInfo []map[string]string, err error) {
	return collectionsOverview()
}

// CollectionEdit_Prepinfo returns the preparatory-info necessary to start editing a collection
func (f *Facilitator) CollectionEdit_Prepinfo(colName string) (schemaJson string, jsonInput string, err error) {
	allowSchemaUpdate := false
	return f._collectionEdit_Prepinfo(colName, allowSchemaUpdate)
}
func (f *Facilitator) _collectionEdit_Prepinfo(colName string, allowSchemaUpdate bool) (schemaJson string, jsonInput string, err error) {
	var col *Collection
	col, err = collectionLoad(colName)
	if err != nil {
		return "", "", err
	}
	if err = col.canBeUpdated(); err != nil {
		return "", "", err
	}

	schemaLatest, err := schemaLoadLatest()
	if err != nil {
		return "", "", err
	}
	schemaJson = schemaLatest.Json

	cselLatest, err := col.colSelectionLatest()
	if err != nil {
		return "", "", err
	}

	// thisEdit::jsonInput = formerEdit::jsonOutput
	jsonInput = cselLatest.JsonOutput

	if !allowSchemaUpdate {
		// Safety validation: we dont accept schema-updates between cselLatest and now
		if cselLatest.SchemaID != schemaLatest.ID {
			return "", "", fmt.Errorf("safety-protection, a schema-update is not allowed between last edit and now")
		}
	}

	return schemaJson, jsonInput, nil
}

// CollectionEdit_Save updates the collection
func (f *Facilitator) CollectionEdit_Save(colName string, schemaJson string, jsonInput string, jsonOutput string, requestingUser string) error {
	col, err := collectionLoad(colName)
	if err != nil {
		return err
	}
	if err = col.canBeUpdated(); err != nil {
		return err
	}

	// Assure schemaJson == schemaLatest.Json
	schemaLatest, err := schemaLoadLatest()
	if err != nil {
		return err
	}
	if schemaJson != schemaLatest.Json {
		return fmt.Errorf("schemaJson != schemaLatest.Json which is unexpected - cannot save new collection with differing schemaJson")
	}

	err = col.appendAndRunColSelection(schemaLatest, jsonInput, jsonOutput, requestingUser)
	if err != nil {
		return err
	}
	return nil
}

// Create a new collection
//
//	`name` must be compliant with DNS label standard as defined in RFC 1123 (like pod label-names, see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names).
func (f *Facilitator) CollectionNew(colName string) error {
	_, err := collectionNew(colName)
	return err
}

// SchemaEdit_Prepinfo returns the preparatory-info necessary to prepare for a new schema
func (f *Facilitator) SchemaEdit_Prepinfo() (latestSchemaVersionName string, latestSchemaJsonStr string, err error) {
	latestSchema, err := schemaLoadLatest()
	if err != nil {
		return "", "", err
	}
	latestSchemaVersionName = latestSchema.VersionName
	latestSchemaJsonStr = latestSchema.Json
	err = nil
	return latestSchemaVersionName, latestSchemaJsonStr, err
}

// SchemaEdit_SaveAndApplyToAllCollections will create a new schema, and apply it to all existing Collections
func (f *Facilitator) SchemaEdit_SaveAndApplyToAllCollections(newSchemaVersionName string, newSchemaJsonStr string) error {
	// Save new schema latest
	_, err := schemaNew(newSchemaVersionName, newSchemaJsonStr)
	if err != nil {
		return err
	}

	// Apply to all Collections
	requestingUser := "system-apply-new-schema"
	colsInfo, err := f.CollectionsOverview()
	if err != nil {
		return err
	}
	for _, a_colInfo := range colsInfo {
		a_col_Name := a_colInfo["Name"]
		allowSchemaUpdate := true
		schemaLatest, jsonInput, err := f._collectionEdit_Prepinfo(a_col_Name, allowSchemaUpdate)
		if err != nil {
			//TODO: internally log this somehow, its important when a schema-update makes a collection fail!
			fmt.Printf("SchemaUpdate got error: %v \n", err)
		}
		jsonOutput := jsonInput
		err = f.CollectionEdit_Save(a_col_Name, schemaLatest, jsonInput, jsonOutput, requestingUser)
		if err != nil {
			//TODO: internally log this somehow, its important when a schema-update makes a collection fail!
			fmt.Printf("SchemaUpdate got error: %v \n", err)
		}
	}
	return nil
}
