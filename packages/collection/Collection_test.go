package collection

import (
	"os"
	"path/filepath"
	"testing"
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
	err = db.Last(&dbCol).Error
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if dbCol.CreatedAt != collection.CreatedAt {
		t.Errorf("Unexpected last row")
	}

	// Verify that db table of ColSelection has expected new row
	var dbColSel ColSelection
	err = db.Last(&dbColSel).Error
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
	// Verify that db table of ProcessingEngine has expected new rows
	// Verify the State field in all the tables has the expected value

}

/*
func TestCollectionRecalculateStateAndError(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test recalculation of state and error when the latest ColSelection is in the "Pending" state
	collection, _ := collectionNew("valid-name")
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Pending"})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Pending" || collection.Error != nil {
		t.Errorf("Expected collection state to be 'Pending' and error to be nil, but got state '%v' and error '%v'", collection.State, collection.Error)
	}

	// Test recalculation of state and error when the latest ColSelection is in the "Running" state
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Running"})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Running" || collection.Error != nil {
		t.Errorf("Expected collection state to be 'Running' and error to be nil, but got state '%v' and error '%v'", collection.State, collection.Error)
	}

	// Test recalculation of state and error when the latest ColSelection is in the "Completed" state
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Completed"})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Completed" || collection.Error != nil {
		t.Errorf("Expected collection state to be 'Completed' and error to be nil, but got state '%v' and error '%v'", collection.State, collection.Error)
	}

	// Test recalculation of state and error when the latest ColSelection is in the "Failed" state
	expectedError := errors.New("test error")
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Failed", Error: expectedError})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Failed" || collection.Error.Error() != expectedError.Error() {
		t.Errorf("Expected collection state to be 'Failed' and error to be '%v', but got state '%v' and error '%v'", expectedError, collection.State, collection.Error)
	}
}

func TestCollectionCanBeUpdated(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test that _canBeUpdated returns an error when the collection is in the "Completed" state
	collection, _ := collectionNew("valid-name")
	collection.State = "Completed"
	err := collection._canBeUpdated()
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}

	// Test that _canBeUpdated returns no error when the collection is not in the "Completed" state
	collection.State = "Running"
	err = collection._canBeUpdated()
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

func TestCollectionColSelectionLatest(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test that colSelectionLatest returns the latest ColSelection
	collection, _ := collectionNew("valid-name")
	collection.ColSelections = append(collection.ColSelections, &ColSelection{ID: 1})
	collection.ColSelections = append(collection.ColSelections, &ColSelection{ID: 2})
	latest, _ := collection.colSelectionLatest()
	if latest.ID != 2 {
		t.Errorf("Expected latest ColSelection ID to be 2, but got %v", latest.ID)
	}

	// Test that colSelectionLatest returns an error when there are no ColSelections
	collection.ColSelections = []*ColSelection{}
	_, err := collection.colSelectionLatest()
	if err == nil {
		t.Errorf("Expected an error, but got no error")
	}
}

func TestCollectionSaveAndDelete(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test saving and deleting a collection
	collection, _ := collectionNew("valid-name")
	err := collection.save(&collection)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	err = collection.delete(&collection)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

func TestCollectionGormID(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test that gormID returns the ID of the collection
	collection, _ := collectionNew("valid-name")
	if collection.gormID() != collection.ID {
		t.Errorf("Expected gormID to return %v, but got %v", collection.ID, collection.gormID())
	}
}

func TestCollectionReload(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test reloading a collection
	collection, _ := collectionNew("valid-name")
	collection.Name = "newName"
	err := collection.save(&collection)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	err = collection.reload(&collection)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if collection.Name != "valid-name" {
		t.Errorf("Expected collection name to be 'valid-name', but got %v", collection.Name)
	}
}

func TestCollectionStateAndColSelectionState(t *testing.T) {
		reprepForTestCollection(t, "../../tests/Collection/processingEngines")

	// Test that c.State == c.ColSelections.State when c.ColSelections.State == "Completed"
	collection, _ := collectionNew("valid-name")
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Completed"})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Completed" || collection.ColSelections.State != "Completed" {
		t.Errorf("Expected collection state and ColSelections state to be 'Completed', but got collection state '%v' and ColSelections state '%v'", collection.State, collection.ColSelections.State)
	}

	// Test that c.State == c.ColSelections.State when c.ColSelections.State == "Failed"
	expectedError := errors.New("test error")
	collection.ColSelections = append(collection.ColSelections, &ColSelection{State: "Failed", Error: expectedError})
	collection._recalculateStateAndError(collection.ColSelectionLatest())
	if collection.State != "Failed" || collection.ColSelections.State != "Failed" {
		t.Errorf("Expected collection state and ColSelections state to be 'Failed', but got collection state '%v' and ColSelections state '%v'", collection.State, collection.ColSelections.State)
	}
}
*/
