package collection

import (
	"os"
	"path/filepath"
	"testing"

	"gorm.io/gorm/clause"
)

// Ex: 	reprepForTestCollection(t, "../../tests/Collection/processingEngines")
func reprepForTestCollection(t *testing.T, processingEnginesDirpath string) {
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

// 1. cr collection with ok-name works
// 2. cr collection with bad-name fails
func TestCollectionNew(t *testing.T) {
	reprepForTestCollection(t, "../../tests/Collection/processingEngines")
	// Test creating a collection with a valid name
	collection, err := collectionNew("valid-name")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if collection.Name != "valid-name" {
		t.Errorf("Expected collection name to be 'valid-name', but got %v", collection.Name)
	}

	// Test creating a collection with an invalid name
	invalid_names := []string{"", "CamelCase", "under_score", "999"}
	for _, a_invalid_name := range invalid_names {
		_, err = collectionNew(a_invalid_name)
		if err == nil {
			t.Errorf("With invalid-name '%s' expected an error, but got no error", a_invalid_name)
		}
	}
}

// 3. cr collection1 and then try to create collection1 and verify it fails as it already exists
// 5. collectionLoad("existingcol") should work
// 6. collectionLoad("nonexistingcol") should fail
func TestCollectionLoad(t *testing.T) {
	reprepForTestCollection(t, "../../tests/Collection/processingEngines")
	_, err := collectionNew("existingcol")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	// Test loading an existing collection
	collection, err := collectionLoad("existingcol")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if collection.Name != "existingcol" {
		t.Errorf("Expected collection name to be 'existingcol', but got %v", collection.Name)
	}

	// Test loading a non-existing collection
	_, err = collectionLoad("nonexistingcol")
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}
}

// 7. c.appendAndRunColSelection(...) shuold work
// 8. c.appendAndRunColSelection("not-newest-schema",...) shuold fail
// 9. when c.State != "Completed", then c.appendAndRunColSelection() should fail
func TestCollectionAppendAndRunColSelection(t *testing.T) {
	reprepForTestCollection(t, "../../tests/Collection/processingEngines")
	// Test appending and running a ColSelection with the newest schema
	collection, _ := collectionNew("valid-name")
	schema1, err := schemaNew("schema1", "{}")
	if err != nil {
		t.Errorf("Expected no error on creating new schema, but got %v", err)
	}
	err = collection.appendAndRunColSelection(schema1, "jsonInput", "jsonOutput", "requestingUser")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test appending and running a ColSelection with a non-newest schema
	_, err = schemaNew("schema2", "{}")
	if err != nil {
		t.Errorf("Expected no error on creating new schema, but got %v", err)
	}
	err = collection.appendAndRunColSelection(schema1, "jsonInput", "jsonOutput", "requestingUser")
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}

	reprepForTestCollection(t, "../../tests/Collection/PeFail")
	// Test appending and running a ColSelection when the collection is not in the "Completed" state
	col, err := collectionNew("valid-name")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	schema, err := schemaNew("schema", "{}")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	err = col.appendAndRunColSelection(schema, "jsonInput", "jsonOutput", "requestingUser")
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}
	if col.State != "Failed" {
		t.Errorf("Expected state failed, but got state %v", col.State)
	}
	err = col.appendAndRunColSelection(schema, "jsonInput", "jsonOutput", "requestingUser")
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}
}

/*
check db tables

	Create new collection c
	then c.appendAndRunColSelection
	then c.appendAndRunColSelection a second time
	Verify that db table of Collection has expected new row
	Verify that db table of ColSelection has expected new row
	Verify that db table of ProcessingEngineRunner has expected new row
	Verify that db table of ProcessingEngine has expected new rows
	Verify the State field in all the tables has the expected value
*/
func TestCollectionDbTables(t *testing.T) {
	reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Create new collection c
	// then c.appendAndRunColSelection
	// then c.appendAndRunColSelection a second time
	collection, _ := collectionNew("valid-name")
	schema1, err := schemaNew("schema1", "{}")
	if err != nil {
		t.Errorf("Expected no error on creating new schema, but got %v", err)
	}
	err = collection.appendAndRunColSelection(schema1, "jsonInput", "jsonOutput", "requestingUser")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	err = collection.appendAndRunColSelection(schema1, "jsonInput2", "jsonOutput2", "requestingUser")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Verify that db table of Collection has expected new row
	var dbCol Collection
	/*
		GORM Nested Preloading
		Collection
			ColSelections.Schema
			ColSelections.ProcessingEngineRunner
			ColSelections.ProcessingEngineRunner.ProcessingEngines
	*/
	err = db.
		Preload(clause.Associations).    // direct-1-level-deep-fields are loaded
		Preload("ColSelections.Schema"). // 2orMore-level-deep-fields need explicit "nested preloading" for each deep association
		Preload("ColSelections.ProcessingEngineRunner").
		Preload("ColSelections.ProcessingEngineRunner.ProcessingEngines").
		Last(&dbCol).Error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if dbCol.CreatedAt != collection.CreatedAt {
		t.Errorf("Unexpected last row")
	}

	// Verify that db table of ColSelection has expected new row
	var dbColSel ColSelection
	err = db.
		Preload(clause.Associations). // direct-1-level-deep-fields are loaded
		Preload("ProcessingEngineRunner").
		Preload("ProcessingEngineRunner.ProcessingEngines").
		Last(&dbColSel).Error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	colSelLatest, err := collection.colSelectionLatest()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if dbColSel.CreatedAt != colSelLatest.CreatedAt {
		t.Errorf("Unexpected last row")
	}

	// Verify that db table of ProcessingEngineRunner has expected new row
	var dbPer ProcessingEngineRunner
	err = db.
		Preload(clause.Associations). // direct-1-level-deep-fields are loaded
		Last(&dbPer).Error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if dbPer.CreatedAt != colSelLatest.ProcessingEngineRunner.CreatedAt {
		t.Errorf("Unexpected last row")
	}

	// Verify that db table of ProcessingEngine has expected new rows
	var dbPe ProcessingEngine
	err = db.
		Preload(clause.Associations).
		Last(&dbPe).Error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if dbPe.CreatedAt != colSelLatest.ProcessingEngineRunner.ProcessingEngines[len(colSelLatest.ProcessingEngineRunner.ProcessingEngines)-1].CreatedAt {
		t.Errorf("Unexpected last row")
	}

	// Verify the State field in all the tables has the expected value
	// with "processingEngines/" -> "Completed"
	{
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")
		sch, err := schemaNew("schema1", "{}")
		if err != nil {
			t.Errorf("Unexpected error, %v", err)
		}
		col, err := collectionNew("col-b")
		if err != nil {
			t.Errorf("Unexpected error, %v", err)
		}
		err = col.appendAndRunColSelection(sch, "jsonInput", "jsonOutput", "requestingUser")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		// ColSelection.State == "Completed"
		colSelLatest, err := col.colSelectionLatest()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if colSelLatest.State != "Completed" {
			t.Errorf("Detected unexpected state")
		}

		// ProcessineEngineRunner.State == "Completed"
		if colSelLatest.ProcessingEngineRunner.State != "Completed" {
			t.Errorf("Detected unexpected state")
		}

		// ProcessingEngine[-1].State == "Completed"
		if colSelLatest.ProcessingEngineRunner.ProcessingEngines[len(colSelLatest.ProcessingEngineRunner.ProcessingEngines)-1].State != "Completed" {
			t.Errorf("Detected unexpected state")
		}
		// ProcessingEngine[-1].RunExitCode == 0
		if colSelLatest.ProcessingEngineRunner.ProcessingEngines[len(colSelLatest.ProcessingEngineRunner.ProcessingEngines)-1].RunExitcode != 0 {
			t.Errorf("Detected unexpected RunExitCode")
		}
	}
	// with "PeFail/" -> "Failed"
	{
		reprepForTestCollection(t, "../../tests/Collection/PeFail")
		sch, err := schemaNew("schema1", "{}")
		if err != nil {
			t.Errorf("Unexpected error, %v", err)
		}
		col, err := collectionNew("col-b")
		if err != nil {
			t.Errorf("Unexpected error, %v", err)
		}
		err = col.appendAndRunColSelection(sch, "jsonInput", "jsonOutput", "requestingUser")
		if err == nil {
			t.Errorf("Expected error, but got %v", err)
		}
		// ColSelection.State == "Failed"
		colSelLatest, err := col.colSelectionLatest()
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if colSelLatest.State != "Failed" {
			t.Errorf("Detected unexpected state")
		}

		// ProcessineEngineRunner.State == "Failed"
		if colSelLatest.ProcessingEngineRunner.State != "Failed" {
			t.Errorf("Detected unexpected state")
		}

		// ProcessingEngine[-1].State == "Failed"
		if colSelLatest.ProcessingEngineRunner.ProcessingEngines[len(colSelLatest.ProcessingEngineRunner.ProcessingEngines)-1].State != "Failed" {
			t.Errorf("Detected unexpected state")
		}
		// ProcessingEngine[-1].RunExitCode != 0
		if colSelLatest.ProcessingEngineRunner.ProcessingEngines[len(colSelLatest.ProcessingEngineRunner.ProcessingEngines)-1].RunExitcode == 0 {
			t.Errorf("Detected unexpected RunExitCode")
		}
	}

}
