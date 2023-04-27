package collection

import (
	"os"
	"path/filepath"
	"testing"
)

// Ex: 	reprepForTestFacilitator(t, "../../tests/Facilitator/processingEngines")
func reprepForTestFacilitator(t *testing.T, processingEnginesDirpath string) {
	dbFilepath := filepath.Dir(processingEnginesDirpath) + "/sqlite.db"
	_ = os.Remove(dbFilepath)

	f, _ := NewFacilitator()
	f.InitSetup(dbFilepath, processingEnginesDirpath)
	db.Exec("DELETE FROM collections")
	db.Exec("DELETE FROM col_selections")
	db.Exec("DELETE FROM processing_engine_runners")
	db.Exec("DELETE FROM processing_engines")
	db.Exec("DELETE FROM schemas")
	db.Exec("VACUUM")
}

// f.CollectionNew("col1")
// f.CollectionNew("col2")
// and then f.CollectionsOverview should show correct info of both
// f.CollectionNew("col1")
// and then f.CollectionNew("col1") should fail
// f.CollectionNew("col1")
// and then CollectionEdit_Prepinfo() should return correct data
// f.CollectionNew("col1") + f.CollectionEdit_Prepinfo() + f.CollectionEdit_Save
// and then f.CollectionsOverview should show correct info
func TestFacilitator(t *testing.T) {
	reprepForTestFacilitator(t, "../../tests/Facilitator/processingEngines")

	// f.CollectionNew("col1")
	// f.CollectionNew("col2")
	// and then f.CollectionsOverview should show correct info of both
	var f *Facilitator
	var err error
	{
		f, err = NewFacilitator()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		err = f.CollectionNew("col1")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		err = f.CollectionNew("col2")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		colsInfo, err := f.CollectionsOverview()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		for _, a_colInfo := range colsInfo {
			if !((a_colInfo["Name"] == "col1" || a_colInfo["Name"] == "col2") &&
				a_colInfo["State"] == "Completed" &&
				a_colInfo["ErrorStr"] == "") {
				t.Errorf("Expected colsInfo with different content, but got %v", a_colInfo)
			}
		}
	}

	// and then f.CollectionNew("col1") should fail
	{
		err = f.CollectionNew("col1")
		if err == nil {
			t.Errorf("Expected error, but got %v", err)
		}
	}

	// and then CollectionEdit_Prepinfo() should return correct data
	var schemaJson string
	var jsonInput string
	{
		schemaJson, jsonInput, err = f.CollectionEdit_Prepinfo("col1")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaJson != "{}" {
			t.Errorf("Unexpected schemaJson, got %v", schemaJson)
		}
		if jsonInput != "{}" {
			t.Errorf("Unexpected jsonInput, got %v", jsonInput)
		}
	}

	// f.CollectionNew("col1") + f.CollectionEdit_Prepinfo() + f.CollectionEdit_Save
	// and then f.CollectionsOverview should show correct info
	// and then f.CollectionEdit_Prepinfo("col1") should show correct info
	{
		jsonOutput := `{"hi":"there"}`
		err = f.CollectionEdit_Save("col1", schemaJson, jsonInput, jsonOutput, "requestinguser")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaJson != "{}" {
			t.Errorf("Unexpected schemaJson, got %v", schemaJson)
		}
		if jsonInput != "{}" {
			t.Errorf("Unexpected jsonInput, got %v", jsonInput)
		}
		colsInfo, err := f.CollectionsOverview()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		for _, a_colInfo := range colsInfo {
			if !((a_colInfo["Name"] == "col1" || a_colInfo["Name"] == "col2") &&
				a_colInfo["State"] == "Completed" &&
				a_colInfo["ErrorStr"] == "") {
				t.Errorf("Expected colsInfo with different content, but got %v", a_colInfo)
			}
		}

		schemaJson2, jsonInput2, err := f.CollectionEdit_Prepinfo("col1")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaJson2 != "{}" {
			t.Errorf("Unexpected schemaJson2, got %v", schemaJson2)
		}
		if jsonInput2 != jsonOutput {
			t.Errorf("Unexpected jsonInput, got %v", jsonInput2)
		}
	}

}

// Call f.SchemaEdit_Prepinfo and verify return values are correct
// Call f.SchemaEdit_SaveAndApplyToAllCollections + f.SchemaEdit_Prepinfo and verify return values are correct after the save
func TestFacilitatorSchemaEdit(t *testing.T) {
	reprepForTestFacilitator(t, "../../tests/Facilitator/processingEngines")
	var f *Facilitator
	var err error
	{
		f, err = NewFacilitator()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
	}

	// Call f.SchemaEdit_Prepinfo and verify return values are correct
	// Call f.SchemaEdit_SaveAndApplyToAllCollections + f.SchemaEdit_Prepinfo and verify return values are correct after the save
	{
		latestSchemaVersionName, latestSchemaJsonStr, err := f.SchemaEdit_Prepinfo()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if latestSchemaVersionName != "initial-empty-schema" {
			t.Errorf("Unexpected outcome from comparison")
		}
		if latestSchemaJsonStr != "{}" {
			t.Errorf("Unexpected outcome from comparison")
		}
		newSchemaVersionName := "my-schema2"
		newSchemaJsonStr := `{"changed":"new-schema2"}`
		err = f.SchemaEdit_SaveAndApplyToAllCollections(newSchemaVersionName, newSchemaJsonStr)
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		latestSchemaVersionName, latestSchemaJsonStr, err = f.SchemaEdit_Prepinfo()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if latestSchemaVersionName != newSchemaVersionName {
			t.Errorf("Unexpected outcome from comparison")
		}
		if latestSchemaJsonStr != newSchemaJsonStr {
			t.Errorf("Unexpected outcome from comparison")
		}

	}
}
