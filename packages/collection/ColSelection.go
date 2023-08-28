package collection

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

var initialColSelData = map[string]string{
	"jsonInput":      "{}",
	"jsonOutput":     "{}",
	"requestingUser": "init",
}

// State transitions:
//   - State string:   "Pending" > "Running" > "Completed" or "Failed"
//   - Error() error:  set when State=="Failed"
type ColSelection struct {
	gorm.Model
	CollectionID uint // relationship 1Collection-to-manyColSelections
	Collection   *Collection
	// NOTE: ColSelection.Collection already contains a .Catalog, which will be changed over time on renewals
	// ColSelection.Catalog is to keep a track of what was the corresponding Catalog when ColSelection was created and executed
	Catalog                *Catalog // relationship manyColSelections-to-1Catalog
	CatalogID              uint     // relationship manyColSelections-to-1Catalog
	JsonInput              string
	JsonOutput             string
	RequestingUser         string
	ProcessingEngineRunner *ProcessingEngineRunner
	dbMethods
	XState `gorm:"embedded"`
}

func newColSelection(catalog *Catalog, jsonInput string, jsonOutput string, requestingUser string) (*ColSelection, error) {
	// improvement: validate catalog

	o := &ColSelection{
		JsonInput:      jsonInput,
		JsonOutput:     jsonOutput,
		RequestingUser: requestingUser,
		Catalog:        catalog,
		CatalogID:      catalog.ID,
	}

	err := o.stateChange(o, "Pending", nil)
	if err != nil {
		o.stateChange(o, "Failed", err)
		return nil, err
	}

	return o, nil
}

func (o *ColSelection) stateChangePostHandleXState(oldState string, oldError error, newXstate *XState) error {
	err := o.save(o)
	if err != nil {
		return err
	}
	if o.Collection != nil {
		o.Collection._recalculateStateAndError(o)
	}
	return nil
}

func (csel *ColSelection) run() error {
	// reload csel from db
	{
		err := csel.reload(csel)
		if err != nil {
			return err
		}
	}

	// create new per and save it into csel.ProcessingEngineRunner
	var per *ProcessingEngineRunner
	{
		var err error
		per, err = newProcessingEngineRunner()
		csel.ProcessingEngineRunner = per
		err2 := csel.save(csel)
		if err != nil {
			csel.stateChange(csel, "Failed", err)
			return err
		}
		if err2 != nil {
			csel.stateChange(csel, "Failed", err2)
			return err2
		}
	}

	// call per.run() and then csel.reload(csel)
	{
		err := per.run()
		err2 := csel.reload(csel) // reload object from db
		if err != nil {
			csel.stateChange(csel, "Failed", err)
			return err
		}
		if err2 != nil {
			return err2
		}
	}
	return nil
}

// _recalculateStateAndError method: from the per := csel.ProcessingEngineRunner, recalculate csel.State and csel.Error
//
//	  per.State/Error                 =>  csel.State/Error
//	  "Pending"/nil                  =>  "Pending"/nil
//	  "Running"/nil                  =>  "Running"/nil
//	  "Completed"/nil                =>  "Completed"/nil
//	  "Failed"/error                 =>  "Failed"/error
//	Use csel.StateChange(newState, newError)
func (csel *ColSelection) _recalculateStateAndError(per *ProcessingEngineRunner) {
	_ = csel.reload(csel) // reload object from db
	_ = per.reload(per)   // reload object from db

	// skip if per != csel.ProcessingEngineRunner
	if per.ID != csel.ProcessingEngineRunner.ID {
		return
	}
	switch per.State {
	case "Pending":
		csel.stateChange(csel, "Pending", nil)
	case "Running":
		csel.stateChange(csel, "Running", nil)
	case "Completed":
		csel.stateChange(csel, "Completed", nil)
	case "Failed":
		// IMPROVEMENT: This error here should be improved to indicate the originating per
		csel.stateChange(csel, "Failed", per.error())
	default:
		panic(fmt.Sprintf("Unrecognized per.State %s", per.State))
	}
}

func _colSelectionCreateInitial(catalog *Catalog) (*ColSelection, error) {
	initialColSel, err := newColSelection(catalog, initialColSelData["jsonInput"], initialColSelData["jsonOutput"], initialColSelData["requestingUser"])
	if err != nil {
		return nil, err
	}
	return initialColSel, nil
}

func (o *ColSelection) gormID() uint {
	return o.ID
}

func (csel *ColSelection) getTimestamp() (timestamp time.Time, err error) {
	timestamp = csel.CreatedAt
	return timestamp, err
}

func (csel *ColSelection) getTimestampFormated() (timestampFormated string, err error) {
	timestamp, err := csel.getTimestamp()
	if err != nil {
		return "", err
	}
	timestampFormated = timestamp.Format("20060102-150405")
	return timestampFormated, err
}

func (csel *ColSelection) getCollectionEditWorkdirBasename() (collectionEditWorkdirBasename string, err error) {
	timestampStr, err := csel.getTimestampFormated()
	if err != nil {
		return "", err
	}
	collectionEditWorkdirBasename = csel.Collection.Name + "." + timestampStr
	return collectionEditWorkdirBasename, nil
}
