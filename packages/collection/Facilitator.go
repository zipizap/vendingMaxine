package collection

import "fmt"

// Use the Facilitator functions, and avoid messing with anything else inside the package :)
// (Facilitator is just a dummy type, a placeholder of functions to be called by the API :))
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
//	colsInfo, err := CollectionsOverview()
//	for _, a_colInfo := range colsInfo {
//	  fmt.Println("Collection Name: " , a_colInfo["Name"])
//	  fmt.Println("Collection State: ", a_colInfo["State"])
//	  fmt.Println("Collection Error: ", a_colInfo["Error"])
//	}
func (f *Facilitator) CollectionsOverview() (colsInfo []map[string]string, err error) {
	var colList []*Collection
	err = db.Select("name", "state", "error_string").Find(&colList).Error
	if err != nil {
		return nil, err
	}
	for _, col := range colList {
		colsInfo = append(colsInfo, map[string]string{
			"Name":  col.Name,
			"State": col.State,
			"Error": col.ErrorString,
		})
	}
	return colsInfo, nil
}

// CollectionEditStart returns the necessary data to start-editing a collection
func (f *Facilitator) CollectionEditStart(colName string) (schemaLatest *Schema, jsonInput string, err error) {
	var col *Collection
	col, err = CollectionLoad(colName)
	if err != nil {
		return nil, "", err
	}
	if err = col.canBeUpdated(); err != nil {
		return nil, "", err
	}

	schemaLatest, err = SchemaLoadLatest()
	if err != nil {
		return nil, "", err
	}

	cselLatest, err := col.ColSelectionLatest()
	if err != nil {
		return nil, "", err
	}

	// thisEdit::jsonInput = formerEdit::jsonOutput
	jsonInput = cselLatest.JsonOutput

	// Safety validation: we dont accept schema-updates between cselLatest and now
	if cselLatest.SchemaID != schemaLatest.ID {
		return nil, "", fmt.Errorf("safety-protection, a schema-update is not accepted between last edit and now")
	}

	return schemaLatest, jsonInput, nil
}

// CollectionEditSave stores data after an edit-save, to update the collection
func (f *Facilitator) CollectionEditSave(colName string, schema *Schema, jsonInput string, jsonOutput string, requestingUser string) error {
	col, err := CollectionLoad(colName)
	if err != nil {
		return err
	}
	if err = col.canBeUpdated(); err != nil {
		return err
	}
	err = col.AppendAndRunColSelection(schema, jsonInput, jsonOutput, requestingUser)
	if err != nil {
		return err
	}
	return nil
}

// Create a new collection
//
//	`name` must be compliant with DNS label standard as defined in RFC 1123 (like pod label-names, see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names).
func (f *Facilitator) CollectionNew(colName string) error {
	_, err := CollectionNew(colName)
	return err
}
