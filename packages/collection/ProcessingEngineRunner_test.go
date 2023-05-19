package collection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Ex: reprepForTestProcessingEngineRunner(t, "../../tests/ProcessingEngineRunner/processingEngines")
func reprepForTestProcessingEngineRunner(t *testing.T, processingEnginesDirpath string) {
	dbFilepath := filepath.Dir(processingEnginesDirpath) + "/sqlite.db"
	_ = os.Remove(dbFilepath)

	f, _ := NewFacilitator()
	logger, _ := zap.NewProduction()
	slog = logger.Sugar()
	f.InitSetup(dbFilepath, processingEnginesDirpath, slog)
	db.Exec("DELETE FROM collections")
	db.Exec("DELETE FROM col_selections")
	db.Exec("DELETE FROM processing_engine_runners")
	db.Exec("DELETE FROM processing_engines")
	db.Exec("DELETE FROM schemas")
	db.Exec("VACUUM")
}

// creates processingEngines[] correctly and in order, with correct state
func TestNewProcessingEngineRunner(t *testing.T) {
	reprepForTestProcessingEngineRunner(t, "../../tests/ProcessingEngineRunner/processingEngines")
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, per)
	assert.Equal(t, "Pending", per.State)
	assert.Nil(t, per.error())
}

// test to create processingEngines[] successfully and in order, with correct state
func TestProcessingEngineRunner_Run(t *testing.T) {
	reprepForTestProcessingEngineRunner(t, "../../tests/ProcessingEngineRunner/processingEngines")
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, per)

	err = per.run()
	assert.Nil(t, err)

	if len(per.ProcessingEngines) != 3 {
		t.Fatalf("Did not find 3 processingEngines as expected")
	}
	if per.ProcessingEngines[0].BinPath != "../../tests/ProcessingEngineRunner/processingEngines/0000.helloworld.sh" &&
		per.ProcessingEngines[0].State != "Completed" {
		t.Fatalf("Did not find processingEngine[0].BinPath as expected")
	}
	if per.ProcessingEngines[1].BinPath != "../../tests/ProcessingEngineRunner/processingEngines/0100.echo_args_env.sh" &&
		per.ProcessingEngines[1].State != "Completed" {
		t.Fatalf("Did not find processingEngine[1].BinPath as expected")
	}
	if per.ProcessingEngines[2].BinPath != "../../tests/ProcessingEngineRunner/processingEngines/0200.exit_arg1.sh" &&
		per.ProcessingEngines[2].State != "Completed" {
		t.Fatalf("Did not find processingEngine[2].BinPath as expected")
	}

	assert.Equal(t, "Completed", per.State)
	assert.Nil(t, per.error())
}

// second pe fails, should verify per.ErrorString and per.State are correct
func TestProcessingEngineRunner_1PeFails(t *testing.T) {
	reprepForTestProcessingEngineRunner(t, "../../tests/ProcessingEngineRunner/1PeFails")
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, per)

	err = per.run()
	assert.NotNil(t, err)

	// first pe should be "Completed" without error
	first_pe := per.ProcessingEngines[0]
	assert.Equal(t, "Completed", first_pe.State)
	assert.Nil(t, first_pe.error())

	// second pe should be "Failed" with error
	second_pe := per.ProcessingEngines[1]
	assert.Equal(t, "Failed", second_pe.State)
	assert.NotNil(t, second_pe.error())

	if per.ErrorString != "ProcessingEngine ../../tests/ProcessingEngineRunner/1PeFails/0100.exit1.sh gave exit-code 1" {
		t.Fatalf("Did not find expected ErrorString")
	}
	assert.Equal(t, per.State, "Failed")

}

/*
func TestProcessingEngineRunner_recalculateStateAndError(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	pe, err := newProcessingEngine("test", []string{})
	if err != nil {
		t.Fatal(err)
	}
	per.ProcessingEngines = append(per.ProcessingEngines, pe)
	per.recalculateStateAndError(pe)
	assert.Equal(t, "Pending", per.State)
	assert.Nil(t, per.Error)

	pe.StateChange("Running", nil)
	per.recalculateStateAndError(pe)
	assert.Equal(t, "Running", per.State)
	assert.Nil(t, per.Error)

	pe.StateChange("Failed", errors.New("test error"))
	per.recalculateStateAndError(pe)
	assert.Equal(t, "Failed", per.State)
	assert.NotNil(t, per.Error)

	pe.StateChange("Completed", nil)
	per.recalculateStateAndError(pe)
	assert.Equal(t, "Completed", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_Run_Success(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pePath := filepath.Join(tmpDir, "test1")
	err = ioutil.WriteFile(pePath, []byte("#!/bin/bash\necho 'test'"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	initProcessingEngineRunner(tmpDir)
	err = per.run()
	assert.Nil(t, err)
	assert.NotNil(t, per)
	assert.Equal(t, "Completed", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_Run_Fail(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pePath := filepath.Join(tmpDir, "test1")
	err = ioutil.WriteFile(pePath, []byte("#!/bin/bash\nexit 1"), 0777)
	if err != nil {
		t.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	initProcessingEngineRunner(tmpDir)
	err = per.run()
	assert.NotNil(t, err)
	assert.NotNil(t, per)
	assert.Equal(t, "Failed", per.State)
	assert.NotNil(t, per.Error)
}

func TestProcessingEngineRunner_gormID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, per.ID, per.gormID())
}

func TestProcessingEngineRunner_initProcessingEngineRunner(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	initProcessingEngineRunner(tmpDir)
	assert.Equal(t, tmpDir, peDirpath)
}

func TestProcessingEngineRunner_newProcessingEngineRunner(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, per)
	assert.Equal(t, "Pending", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_newProcessingEngineRunner_Fail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	per.StateChange("Running", nil)
	_, err = newProcessingEngineRunner()
	assert.NotNil(t, err)
	assert.Equal(t, "Running", per.State)
}

func TestProcessingEngineRunner_reload(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	per.StateChange("Running", nil)
	err = per.reload(per)
	assert.Nil(t, err)
	assert.Equal(t, "Pending", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_Save(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	per.StateChange("Running", nil)
	err = per.save(per)
	assert.Nil(t, err)
	assert.Equal(t, "Running", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_Save_Fail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	per.StateChange("Running", nil)
	err = per.save(nil)
	assert.NotNil(t, err)
	assert.Equal(t, "Running", per.State)
}

func TestProcessingEngineRunner_StateChange(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	err = per.StateChange("Running", nil)
	assert.Nil(t, err)
	assert.Equal(t, "Running", per.State)
	assert.Nil(t, per.Error)
}

func TestProcessingEngineRunner_StateChange_Fail(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	err = per.StateChange("Running", nil)
	assert.Nil(t, err)
	assert.Equal(t, "Running", per.State)
	assert.Nil(t, per.Error)

	err = per.StateChange("Pending", errors.New("test error"))
	assert.NotNil(t, err)
	assert.Equal(t, "Running", per.State)
	assert.Nil(t, per.Error)
}

/*
func TestProcessingEngineRunner_Xstate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = db.AutoMigrate(&ProcessingEngineRunner{})
	if err != nil {
		t.Fatal(err)
	}
	per, err := newProcessingEngineRunner()
	if err != nil {
		t.Fatal(err)
	}
	xs := XState{State: "test", ErrorString: "test error"}
	per.Xstate = xs
	assert.Equal(t, xs, per.Xstate)
}
*/
