package collection

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"gorm.io/gorm"
)

// State transitions:
//   - State string:   "Pending" > "Running" > "Completed" or "Failed"
//   - Error() error:  set when State=="Failed"
type ProcessingEngineRunner struct {
	gorm.Model
	ColSelectionID                 uint
	ColSelection                   *ColSelection
	CollectionEditWorkdirTgzBlobId uint
	dbMethods
	XState `gorm:"embedded"`
}

func newProcessingEngineRunner() (*ProcessingEngineRunner, error) {
	o := &ProcessingEngineRunner{}
	err := o.stateChange(o, "Pending", nil)
	if err != nil {
		o.stateChange(o, "Failed", err)
		return nil, err
	}
	return o, nil
}
func (o *ProcessingEngineRunner) stateChangePostHandleXState(oldState string, oldError error, newXstate *XState) error {
	// db save
	err := o.save(o)
	if err != nil {
		return err
	}

	if o.ColSelection != nil {
		o.ColSelection._recalculateStateAndError(o)
	}
	return nil
}

func (per *ProcessingEngineRunner) run() (err error) {
	// reload csel from db
	{
		err := per.reload(per)
		if err != nil {
			return err
		}
	}

	// set per state to "Running"
	{
		per.stateChange(per, "Running", nil)
		err := per.reload(per) // reload object from db
		if err != nil {
			return err
		}
	}

	// create and prepare collectionEditWorkdir
	var collectionEditWorkdir string // fullpath
	var colSel *ColSelection
	var col *Collection
	var cat *Catalog
	{
		// get colSel > col > cat
		{
			colSel = per.ColSelection
			err := colSel.reload(colSel)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			col = colSel.Collection
			err = col.reload(col)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			cat = col.Catalog
			err = cat.reload(cat)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
		}

		// set collectionEditWorkdir and create its dir
		{
			// get collectionEditWorkdir_relPath
			colName := col.Name
			dateYYYYMMDD_hhmmss, err := col.getTimestampFormated()
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			collectionEditWorkdir_relPath := colName + "." + dateYYYYMMDD_hhmmss

			// get a temporary directory parentDirFullpath, to contain the collectionEditWorkdir
			parentDirFullpath, err := os.MkdirTemp("", "tmp")
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			// on DEBUG logleveldont delete parentdirs, to facilitate local debugging
			if slogGetLevel() != "DEBUG" {
				defer os.RemoveAll(parentDirFullpath)
			}

			// set collectionEditWorkdir (inside parentDirFullpath)
			collectionEditWorkdir = filepath.Join(parentDirFullpath, collectionEditWorkdir_relPath)

			// mkdir collectionEditWorkdir
			err = os.Mkdir(collectionEditWorkdir, os.ModePerm)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
		}

		// prepare files inside collectionEditWorkdir
		{
			// mkdir collectionEditFilesSubdir  (<collectionEditWorkdir>/CollectionEditFiles/)
			var collectionEditFilesSubdir string
			{
				collectionEditFilesSubdir = filepath.Join(collectionEditWorkdir, "CollectionEditFiles")
				err = os.Mkdir(collectionEditFilesSubdir, os.ModePerm)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
			}

			// helper func
			saveFile := func(filedir string, filename string, fileContent string, fileModePerm fs.FileMode) (err error) {
				aFileFullpath := filepath.Join(filedir, filename)
				err = os.WriteFile(aFileFullpath, []byte(fileContent), 0644)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
				return nil
			}
			// create <collectionEditWorkdir>/CollectionEditFiles/Schema.json
			{
				aFileName := "Schema.json"
				aFileContent, err := cat.schema()
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
				err = saveFile(collectionEditFilesSubdir, aFileName, aFileContent, 0400)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
			}
			// create <collectionEditWorkdir>/CollectionEditFiles/JsonInput.json
			{
				aFileName := "JsonInput.json"
				aFileContent := colSel.JsonInput
				err = saveFile(collectionEditFilesSubdir, aFileName, aFileContent, 0400)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
			}
			// create <collectionEditWorkdir>/CollectionEditFiles/JsonOutput.orig.json
			// create <collectionEditWorkdir>/CollectionEditFiles/JsonOutput.json
			{
				aFileName := "JsonOutput.orig.json"
				aFileContent := colSel.JsonOutput
				err = saveFile(collectionEditFilesSubdir, aFileName, aFileContent, 0400)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
				aFileName = "JsonOutput.json"
				err = saveFile(collectionEditFilesSubdir, aFileName, aFileContent, 0600)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
			}
			// create <collectionEditWorkdir>/CollectionEditFiles/PeConfig.json
			{
				aFileName := "PeConfig.json"
				aFileContent := `
{
	"catalog": {
		"name": "` + jsonEscape(cat.Name) + `"
	},
	"collection": {
		"name": "` + jsonEscape(col.Name) + `",
		"previousState": "` + jsonEscape(col.State) + `",
		"previousErrorStr": "` + jsonEscape(col.ErrorString) + `"
	},
	"collection-edit": {
		"schemaFilepath": "./CollectionEditFiles/Schema.json",
		"jsonInputFilepath": "./CollectionEditFiles/JsonInput.json",
		"jsonOutputFilepath": "./CollectionEditFiles/JsonOutput.json"
	}
}
`
				err = saveFile(collectionEditFilesSubdir, aFileName, aFileContent, 0400)
				if err != nil {
					per.stateChange(per, "Failed", err)
					return err
				}
			}
		}
	}

	// run  <catalogDirpath>/bin/internal/bash -c "<catalogDirpath>/bin/CollectionEdit.Launch.sh <collectionEditWorkdir>"
	// and save CollectionEditWorkdirTgzBlobId
	{
		catalogDirpath, err := cat.catalogDir()

		if err != nil {
			per.stateChange(per, "Failed", err)
			return err
		}
		launcherBinAndArgs := filepath.Join(catalogDirpath, "bin/CollectionEdit.Launch.sh") + "  " + collectionEditWorkdir
		cmdBin := catalogDirpath + "/bin/internal/bash"
		cmdBinAndArgs := []string{cmdBin, "-c", launcherBinAndArgs}
		slog.Infof("[PerID %d] Run starting> %s %s '%s'", per.ID, cmdBinAndArgs[0], cmdBinAndArgs[1], cmdBinAndArgs[2])
		exitCode, stdOutErr, errRun := per._runProcess(cmdBinAndArgs[0], cmdBinAndArgs[1:]...)
		slog.Infof("[PerID %d] Run completed> exit-code: %d ", per.ID, exitCode)
		fmt.Println(stdOutErr)

		// save per.CollectionEditWorkdirTgzBlobId before checking errRun
		{
			tgzData, err := compressDir2Tgz(collectionEditWorkdir)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			blob, err := blobNew(tgzData)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
			per.CollectionEditWorkdirTgzBlobId = blob.ID
			err = per.save(per)
			if err != nil {
				per.stateChange(per, "Failed", err)
				return err
			}
		}
		// check errRun
		if errRun != nil {
			per.stateChange(per, "Failed", errRun)
			return errRun
		}
	}

	per.stateChange(per, "Completed", nil)
	return nil
}

func (per *ProcessingEngineRunner) _runProcess(binPath string, args ...string) (exitCode int, stdOutErr string, err error) {
	var stdOutErrBuf bytes.Buffer
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = &stdOutErrBuf
	cmd.Stderr = &stdOutErrBuf
	err = cmd.Run()
	exitCode = cmd.ProcessState.ExitCode()
	stdOutErr = stdOutErrBuf.String()
	if exitCode != 0 {
		err = fmt.Errorf("ProcessingEngineLauncher %s gave exit-code %d", binPath, exitCode)
	}
	return
}

func (o *ProcessingEngineRunner) gormID() uint {
	return o.ID
}

func (per *ProcessingEngineRunner) getCollectionEditWorkdirTgz() (collectionEditWorkdirTgz []byte, err error) {
	return blobData(per.CollectionEditWorkdirTgzBlobId)
}
