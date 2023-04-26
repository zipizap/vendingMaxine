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
// and then CollectionEditStart() should return correct data
// f.CollectionNew("col1") + f.CollectionEditStart() + f.CollectionEditSave
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

	// and then CollectionEditStart() should return correct data
	var schemaLatest *Schema
	var jsonInput string
	{
		schemaLatest, jsonInput, err = f.CollectionEditStart("col1")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaLatest.VersionName != "initial-empty-schema" {
			t.Errorf("Unexpected schemaLatest.VersionName, got %v", schemaLatest.VersionName)
		}
		if jsonInput != "{}" {
			t.Errorf("Unexpected jsonInput, got %v", jsonInput)
		}
	}

	// f.CollectionNew("col1") + f.CollectionEditStart() + f.CollectionEditSave
	// and then f.CollectionsOverview should show correct info
	// and then f.CollectionEditStart("col1") should show correct info
	{
		jsonOutput := `{"hi":"there"}`
		err = f.CollectionEditSave("col1", schemaLatest, jsonInput, jsonOutput, "requestinguser")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaLatest.VersionName != "initial-empty-schema" {
			t.Errorf("Unexpected schemaLatest.VersionName, got %v", schemaLatest.VersionName)
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

		schemaLatest2, jsonInput2, err := f.CollectionEditStart("col1")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if schemaLatest2.VersionName != "initial-empty-schema" {
			t.Errorf("Unexpected schemaLatest.VersionName, got %v", schemaLatest2.VersionName)
		}
		if jsonInput2 != jsonOutput {
			t.Errorf("Unexpected jsonInput, got %v", jsonInput2)
		}
	}

}
