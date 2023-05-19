package collection

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

var prepForTestProcessingEngineOnlyOnceFlag bool

func prepForTestProcessingEngine(t *testing.T) {
	if prepForTestProcessingEngineOnlyOnceFlag == true {
		return
	}
	if err := os.Chdir("../../tests/ProcessingEngine/"); err != nil {
		t.Fatalf("Error changing working directory: %v", err)
	}
	dbFilepath := "./sqlite.db"
	processingEnginesDirpath := "./processingEngines"
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
	prepForTestProcessingEngineOnlyOnceFlag = true
}

func TestNewProcessingEngine(t *testing.T) {
	prepForTestProcessingEngine(t)

	binPath := "processingEngines/0200.echo_args_env.sh"
	runArgs := []string{"arg1", "arg2"}

	pe, err := newProcessingEngine(binPath, runArgs)

	if err != nil {
		t.Errorf("Error creating ProcessingEngine: %s", err.Error())
	}

	if pe.BinPath != binPath {
		t.Errorf("Expected BinPath to be %s, but got %s", binPath, pe.BinPath)
	}

	if len(pe._runArgs()) != len(runArgs) {
		t.Errorf("Expected RunArgs to have length %d, but got %d", len(runArgs), len(pe._runArgs()))
	}
}

func TestProcessingEngine_Run(t *testing.T) {
	prepForTestProcessingEngine(t)

	binPath := "processingEngines/0200.echo_args.sh"
	runArgs := []string{"arg1", "arg2"}

	pe, err := newProcessingEngine(binPath, runArgs)
	if err != nil {
		t.Errorf("Error creating ProcessingEngine: %s", err.Error())
	}

	err = pe.run()
	if err != nil {
		t.Errorf("Error running ProcessingEngine: %s", err.Error())
	}

	fileInfo, err := os.Stat(pe.BinPath)
	if err != nil {
		t.Errorf("Error getting file info for %s: %s", pe.BinPath, err.Error())
	}

	if pe.BinLastModTime != fileInfo.ModTime() {
		t.Errorf("Expected BinLastModTime to be %s, but got %s", fileInfo.ModTime(), pe.BinLastModTime)
	}

	if pe.RunExitcode != 0 {
		t.Errorf("Expected RunExitcode to be 0, but got %d", pe.RunExitcode)
	}

	if pe.State != "Completed" {
		t.Errorf("Expected State to be Completed, but got %s", pe.State)
	}

	if pe.Error() != nil {
		t.Errorf("Expected Error to be nil, but got %s", pe.Error().Error())
	}
}

func TestProcessingEngine_Run_Failed(t *testing.T) {
	prepForTestProcessingEngine(t)
	binPath := "processingEngines/0200.echo_args_env.sh"
	binPath = "nonexistentfile"
	runArgs := []string{"arg1", "arg2"}

	pe, err := newProcessingEngine(binPath, runArgs)

	if err != nil {
		t.Errorf("Error creating ProcessingEngine: %s", err.Error())
	}

	// Change the BinPath to a non-existent file to force an error

	err = pe.run()

	if err == nil {
		t.Errorf("Expected an error running ProcessingEngine, but got nil")
	}

	if pe.State != "Failed" {
		t.Errorf("Expected State to be Failed, but got %s", pe.State)
	}

	if pe.Error() == nil {
		t.Errorf("Expected Error to be non-nil, but got nil")
	}
}

func TestProcessingEngine_Run_Failed_ExitCode(t *testing.T) {
	prepForTestProcessingEngine(t)
	binPath := "processingEngines/0100.exit_arg1.sh"
	runArgs := []string{"3"}

	pe, err := newProcessingEngine(binPath, runArgs)

	if err != nil {
		t.Errorf("Error creating ProcessingEngine: %s", err.Error())
	}
	err = pe.run()
	if err == nil {
		t.Errorf("Expected an error running ProcessingEngine, but got nil")
	}

	if pe.State != "Failed" {
		t.Errorf("Expected State to be Failed, but got %s", pe.State)
	}

	if pe.Error() == nil {
		t.Errorf("Expected Error to be non-nil, but got nil")
	}

	if pe.RunExitcode != 3 {
		t.Errorf("Expected RunExitcode to be 1, but got %d", pe.RunExitcode)
	}
}
