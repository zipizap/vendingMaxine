package collection

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CollectionInfo struct {
	Name        string `json:"name"`
	CatalogName string `json:"catalogName"`
	State       string `json:"state"`
	ErrorStr    string `json:"errorStr"`
}

// State transitions:
//   - State string:                  (never "Pending") "Running" > "Completed" or "Failed"
//   - Error() error:                 set when State=="Failed"
type Collection struct {
	gorm.Model
	Name          string          `gorm:"unique,uniqueIndex,not null"`
	Catalog       *Catalog        // relationship manyCollection-to-1Catalog
	CatalogID     uint            // relationship manyCollection-to-1Catalog
	ColSelections []*ColSelection // relationship 1Collection-to-manyColSelections
	dbMethods
	XState `gorm:"embedded"`
}

func collectionNew(newCollectionName, catalogName string) (*Collection, error) {
	// validate newCollectionName
	{
		if err := _isValidDNSLabel(newCollectionName); err != nil {
			return nil, err
		}
	}

	// if collection already exists, return error
	if _, err := collectionLoad(newCollectionName); err == nil {
		return nil, fmt.Errorf("Collection %v already exists", newCollectionName)
	}

	// validate catalogName
	if catalogName == "" {
		return nil, fmt.Errorf("CatalogName '%v' invalid", catalogName)
	}

	// load catalog from catalogName
	var catalog *Catalog
	{
		var err error
		catalog, err = catalogLoad(catalogName)
		if err != nil {
			return nil, err
		}
	}

	// Create new object, save to db
	o := &Collection{}
	{
		initialColSel, err := _colSelectionCreateInitial(catalog)
		if err != nil {
			return nil, err
		}
		o.Name = newCollectionName
		o.Catalog = catalog
		o.CatalogID = catalog.ID
		o.ColSelections = append(o.ColSelections, initialColSel)
		err = o.save(o)
		if err != nil {
			return nil, err
		}
	}

	// set State to "Completed"
	{
		err := o.stateChange(o, "Completed", nil)
		if err != nil {
			o.stateChange(o, "Failed", err)
			return nil, err
		}
	}

	// return object
	return o, nil
}

// interface stateChangePostHandleXStater
func (o *Collection) stateChangePostHandleXState(oldState string, oldError error, newXstate *XState) error {
	err := o.save(o)
	if err != nil {
		return err
	}
	return nil
}

// collectionLoad loads from db
func collectionLoad(name string) (*Collection, error) {
	if err := _isValidDNSLabel(name); err != nil {
		return nil, err
	}

	o := &Collection{}
	// The following db.Where... will not do nested-preloading, that will be done latter
	// with the o.reload(o) call
	err := db.Where("name = ?", name).First(o).Error
	if err != nil {
		return nil, err
	}
	err = o.reload(o)
	if err != nil {
		return nil, err
	}
	return o, nil
}

// collectionsOverview returns list of maps with usefull info of all collections
//
//	colsInfo, err := collectionsOverview()
//	for _, a_colInfo := range colsInfo {
//	  fmt.Println("Collection Name: " , a_colInfo.Name)
//	  fmt.Println("Collection CatalogName: " , a_colInfo.CatalogName)
//	  fmt.Println("Collection State: ", a_colInfo.State)
//	  fmt.Println("Collection ErrorStr: ", a_colInfo.ErrorStr)
//	}
func collectionsOverview() (colsInfo []CollectionInfo, err error) {
	colList := []*Collection{}
	err = db.
		Preload(clause.Associations). // direct-1-level-deep-fields are loaded
		Find(&colList).Error
	if err != nil {
		return nil, err
	}
	for _, col := range colList {
		colsInfo = append(colsInfo, CollectionInfo{
			Name:        col.Name,
			CatalogName: col.Catalog.Name,
			State:       col.State,
			ErrorStr:    col.ErrorString,
		})
	}
	if colsInfo == nil {
		colsInfo = []CollectionInfo{}
	}
	return colsInfo, nil
}

func (c *Collection) getTimestamp() (timestamp time.Time, err error) {
	var colSelLatest *ColSelection
	{
		colSelLatest, err = c.colSelectionLatest()
		if err != nil {
			return timestamp, err
		}
	}

	return colSelLatest.getTimestamp()
}

func (c *Collection) getTimestampFormated() (timestampFormated string, err error) {
	var colSelLatest *ColSelection
	{
		colSelLatest, err = c.colSelectionLatest()
		if err != nil {
			return "", err
		}
	}

	return colSelLatest.getTimestampFormated()
}

func (c *Collection) getCollectionReplayableDirTgz() (collectionReplayableDirTgz []byte, collectionReplayableBasename string, err error) {
	// <collectionReplayableParentDir>/<collectionReplayableBasename>
	// |------------<collectionReplayableDir>-----------------------|
	//
	// <collectionReplayableParentDir>/<collectionReplayableBasename>/<collectionEditWorkdirBasename>
	// |------------<collectionEditWorkdir>---------------------------------------------------------|
	//
	// <collectionReplayableParentDir>/<collectionReplayableBasename>/<catalogDirBasename>
	// |------------<catalogDir>---------------------------------------------------------|
	//
	var collectionReplayableParentDir, collectionReplayableDir, catalogDirBasename, catalogDir, collectionEditWorkdirBasename, collectionEditWorkdir string

	// create collectionReplayableDir with its subdirs and files
	{
		// set and mkdir collectionReplayableDir
		{
			// set and mkdir collectionReplayableParentDir
			{
				collectionReplayableParentDir, err = os.MkdirTemp("", "collectionReplayableParentDir")
				if err != nil {
					return nil, "", err
				}
				defer os.RemoveAll(collectionReplayableParentDir)
			}

			// set and mkdir collectionReplayableDir and collectionReplayableBasename
			{
				collectionReplayableBasename, err = c.getCollectionReplayableDirBasename()
				if err != nil {
					return nil, "", err
				}

				collectionReplayableDir = filepath.Join(collectionReplayableParentDir, collectionReplayableBasename)
				err = os.MkdirAll(collectionReplayableDir, 0700)
				if err != nil {
					return nil, "", err
				}
			}
		}

		// set, mkdir and extract catalogDir
		{
			catalogDirBasename = c.Catalog.getCatalogDirBasenameString()
			if err != nil {
				return nil, "", err
			}

			catalogDir = filepath.Join(collectionReplayableDir, catalogDirBasename)
			err = os.MkdirAll(catalogDir, 0700)
			if err != nil {
				return nil, "", err
			}

			catalogDirTgz, err := c.Catalog.getCatalogDirTgz()
			if err != nil {
				return nil, "", err
			}
			err = extractTgz2Dir(catalogDirTgz, catalogDir)
			if err != nil {
				return nil, "", err
			}
		}

		// set, mkdir and extract collectionEditWorkdir
		var cselLatest *ColSelection
		{
			cselLatest, err = c.colSelectionLatest()
			if err != nil {
				return nil, "", err
			}

			collectionEditWorkdirBasename, err = cselLatest.getCollectionEditWorkdirBasename()
			if err != nil {
				return nil, "", err
			}

			collectionEditWorkdir = filepath.Join(collectionReplayableDir, collectionEditWorkdirBasename)
			err = os.MkdirAll(collectionEditWorkdir, 0700)
			if err != nil {
				return nil, "", err
			}

			var per *ProcessingEngineRunner
			{
				per = cselLatest.ProcessingEngineRunner
				err = per.reload(per)
				if err != nil {
					return nil, "", err
				}
			}

			collectionEditWorkdirTgz, err := per.getCollectionEditWorkdirTgz()
			if err != nil {
				return nil, "", err
			}
			err = extractTgz2Dir(collectionEditWorkdirTgz, collectionEditWorkdir)
			if err != nil {
				return nil, "", err
			}
		}

		// create README.md
		var readmeFilepath string
		{
			readmeFilepath = filepath.Join(collectionReplayableDir, "README.md")
			readmeContent := `
This directory is a **Collection replayable**, which contains *all* the necessary files to:

- analyse last CollectionEdit logs, in ` + "`" + collectionEditWorkdirBasename + `/CollectionEditFiles/TrackLogs/*

- optionally make local replays (re-execute) of the CollectionEdit, for troubleshooting and development
  ` + "```" + `
  # Set manually the required env-vars (like tokens and secrets), and then run a replay with:
  ./` + catalogDirBasename + `/bin/internal/bash -c '` + catalogDirBasename + `/bin/CollectionEdit.Replay.sh ` + collectionEditWorkdirBasename + `/' 
 
  TODO: sample printout 
  ` + "```" + `

  Replays execute the same Catalog (and Processing Engines) used by the vendingmaxine, in a *self-contained reproducible shell enviroment* that is copied end executed locally. 

  The replays are mostly usefull for troubleshooting unexpected problems, and for developing new versions of the Catalog/Processing Engines.

  These local replays are run independently, its results are not uploaded to the vendingmaxine. They will however re-execute the processing engines, including any change made by them
  But this execution happens always outside and independently of the vendingMaxine. 

  To see debug traces from bash scripts, set ` + "`" + `export DEBUGBASHXTRACE=true` + "`" + ` before execution.

`
			err = writeToFile(readmeContent, readmeFilepath)
			if err != nil {
				return nil, "", err
			}
		}
	}

	// create collectionReplayableDirTgz from collectionReplayableParentDir
	{
		collectionReplayableDirTgz, err = compressDir2Tgz(collectionReplayableParentDir)
		if err != nil {
			return nil, "", err
		}
	}

	return collectionReplayableDirTgz, collectionReplayableBasename, nil
}

func (c *Collection) getCollectionReplayableDirBasename() (collectionReplayableDirBasename string, err error) {
	colSelLatest, err := c.colSelectionLatest()
	if err != nil {
		return "", err
	}

	collectionEditWorkdirBasename, err := colSelLatest.getCollectionEditWorkdirBasename()
	if err != nil {
		return "", err
	}

	collectionReplayableDirBasename = "replayable." + collectionEditWorkdirBasename
	return collectionReplayableDirBasename, nil
}

func (c *Collection) appendAndRunColSelection(jsonInput string, jsonOutput string, requestingUser string) error {
	// Reload c from db, to assure it's in-sync
	if err := c.reload(c); err != nil {
		return err
	}
	// Verify c can be updated
	if err := c.canBeUpdated(); err != nil {
		return err
	}

	// Verify jsonInput == currentColSel.JsonOutput
	// It must always (except when its the first colSel in which case the verification is skipped)
	{
		currentColSel, err := c.colSelectionLatest()
		if err != nil {
			return err
		}
		isInitialColSel := false
		if currentColSel.JsonInput == initialColSelData["jsonInput"] &&
			currentColSel.JsonOutput == initialColSelData["jsonOutput"] &&
			currentColSel.RequestingUser == initialColSelData["requestingUser"] {
			isInitialColSel = true
		}
		if !isInitialColSel {
			// this is not first colSel, lets verify: jsonInput == currentColSel.JsonOutput
			if jsonInput != currentColSel.JsonOutput {
				return fmt.Errorf("error: jsonInput =! currentColSel.JsonOutput, aborting change-request, no changes made")
			}
		}
	}

	// Create new ColSelection, append into c.ColSelections, and save c to db
	var csel *ColSelection
	{
		var err error
		csel, err = newColSelection(c.Catalog, jsonInput, jsonOutput, requestingUser)
		if err != nil {
			return err
		}
		c.ColSelections = append(c.ColSelections, csel)
		err = c.save(c) // everytime we change c, we need to save it on db as soon as possible
		if err != nil {
			return err
		}
	}

	// Run ColSelection in paralel routine and return
	go csel.run()
	return nil
}

func (c *Collection) colSelectionLatest() (*ColSelection, error) {
	cselsLen := len(c.ColSelections)
	csel := c.ColSelections[cselsLen-1]

	err := csel.reload(csel)
	if err != nil {
		return nil, err
	}

	return csel, nil
}

func (c *Collection) canBeUpdated() error {
	switch c.State {
	case "Running":
		return fmt.Errorf("collecion %s cannot be edited/updated as its in state %s", c.Name, c.State)
	}
	return nil
}

// _recalculateStateAndError method:
//
//	from the cs := c.ColSelections[-1], recalculate c.State and c.Error
//	  cs.State/Error                 =>  c.State/Error
//	  "Pending"/nil                  =>  "Pending"/nil
//	  "Running"/nil                  =>  "Running"/nil
//	  "Completed"/nil                =>  "Completed"/nil
//	  "Failed"/error                 =>  "Failed"/error
//	Use c.StateChange(newState, newError)
func (c *Collection) _recalculateStateAndError(csel *ColSelection) {
	_ = c.reload(c)       // reload object from db
	_ = csel.reload(csel) // reload object from db

	// skip if csel != cs (cs := c.ColSelections[-1])
	cs := c.ColSelections[len(c.ColSelections)-1]
	if cs.ID != csel.ID {
		return
	}
	switch csel.State {
	case "Pending":
		c.stateChange(c, "Pending", nil)
	case "Running":
		c.stateChange(c, "Running", nil)
	case "Completed":
		c.stateChange(c, "Completed", nil)
	case "Failed":
		// IMPROVEMENT: This error here should be improved to indicate the originating colSelection
		c.stateChange(c, "Failed", csel.error())
	default:
		panic(fmt.Sprintf("Unrecognized csel.State %s", csel.State))
	}
}

func (o *Collection) gormID() uint {
	return o.ID
}

func _isValidDNSLabel(name string) error {
	errBack := fmt.Errorf("invalid name '%s' (ex: 'lowcase-nounderscores-63max')", name)
	if len(name) > 63 {
		return errBack
	}
	// not perfect, but good enough ;)
	r, _ := regexp.Compile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
	if b := r.MatchString(name); !b {
		return errBack
	}
	return nil
}
