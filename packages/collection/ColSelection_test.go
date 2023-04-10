package collection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ex: reprepForTestCelSelection(t, "../../tests/ProcessingEngineRunner/processingEngines")
func reprepForTestCelSelection(t *testing.T, processingEnginesDirpath string) {
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

// run cs and verify correct cs.State and cs.Error() with per.State
func TestRunColSelection_PeAllOk(t *testing.T) {
	reprepForTestCelSelection(t, "../../tests/ColSelection/PeAllOk")

	// create a new ColSelection object
	schema, err := schemaLoadLatest()
	assert.Nil(t, err)

	jsonInput := "{}"
	jsonOutput := "{}"
	requestingUser := "test"
	csel, err := newColSelection(schema, jsonInput, jsonOutput, requestingUser)
	assert.Nil(t, err)

	// run ColSelection object
	err = csel.run()
	assert.Nil(t, err)

	// verify correct cs.State and cs.Error() with per.State
	assert.Equal(t, csel.State, csel.ProcessingEngineRunner.State)
	assert.Equal(t, csel.ErrorString, csel.ProcessingEngineRunner.ErrorString)
}

// run cs and verify correct cs.State and cs.Error() and per.State when 1pe-failed
func TestRunColSelection_1PeFails(t *testing.T) {
	reprepForTestCelSelection(t, "../../tests/ColSelection/1PeFails")

	// create a new ColSelection object
	schema, err := schemaLoadLatest()
	assert.Nil(t, err)

	jsonInput := "{}"
	jsonOutput := "{}"
	requestingUser := "test"
	csel, err := newColSelection(schema, jsonInput, jsonOutput, requestingUser)
	assert.Nil(t, err)

	// run ColSelection object
	err = csel.run()
	assert.NotNil(t, err)

	// verify correct cs.State and cs.Error() with per.State
	assert.Equal(t, csel.State, csel.ProcessingEngineRunner.State)
	assert.Equal(t, csel.ErrorString, csel.ProcessingEngineRunner.ErrorString)
}

// run cs with a schema that is not latest-schema, should fail with error
func TestRunColSelection_BadSchema(t *testing.T) {
	reprepForTestCelSelection(t, "../../tests/ColSelection/1PeFails")

	// create a new ColSelection object
	schema, err := schemaLoadLatest()
	assert.Nil(t, err)

	// create a newer schema
	_, err = schemaNew("newer schema", "{}")
	assert.Nil(t, err)

	jsonInput := "{}"
	jsonOutput := "{}"
	requestingUser := "test"
	_, err = newColSelection(schema, jsonInput, jsonOutput, requestingUser)
	assert.NotNil(t, err)

}
