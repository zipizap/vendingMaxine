package dbModels

import (
	"fmt"
	"time"
	"vendingMaxine/packages/gormCrud"
)

// DbColRevisionIfc defines the interface for DbColRevision operations
type DbColRevisionIfc interface {
	GetID() uint
	GetDbCollectionID() (uint, error)
	GetDbRevStates() ([]*DbRevState, error)
	GetCreationDate() (time.Time, error)
	GetModDate() (time.Time, error)
	GetDbRevStateLatest() (dbRevState *DbRevState, err error)
	GetDbRevStateLatestName() (string, error)
	IsEditable() (bool, error)
	AppendDbRevState(revStateName string, userWhoTriggered string, logs []byte) error
	IsValidRevStateTransition(newStateName string) (isValid bool, err error)
}

// DbColRevision represents a collection revision in the database
type DbColRevision struct {
	gormCrud.GormCrud[DbColRevision]
	DbCollectionID uint          // 1DbCollection-to-manyDbColRevisions - Foreign key to DbCollection
	DbRevStates    []*DbRevState // 1DbColRevision-to-manyDbRevStates
}

// DbColRevisionNew creates a new DbColRevision
func DbColRevisionNew(dbCollectionID uint, initialRevStateName string, userWhoTriggered string) (dbColRev *DbColRevision, err error) {
	dbColRev = &DbColRevision{}

	// Set dbColRev.DbCollectionID
	dbColRev.DbCollectionID = dbCollectionID
	// Save the DbColRevision to sets its ID required for DbRevStateNew()
	err = dbColRev.Save(dbColRev)
	if err != nil {
		return nil, err
	}

	// Set dbColRev.DbRevStates
	initialDbRevState, err := DbRevStateNew(dbColRev.ID, initialRevStateName, userWhoTriggered, nil)
	if err != nil {
		return nil, err
	}
	dbColRev.DbRevStates = append(dbColRev.DbRevStates, initialDbRevState)
	err = dbColRev.Save(dbColRev)
	if err != nil {
		return nil, err
	}
	return dbColRev, nil
}

// DbColRevisionLoad loads a DbColRevision by ID, (withuot preloading DbRevStates)
func DbColRevisionLoad(dbColRevID uint) (*DbColRevision, error) {
	dbColRev := &DbColRevision{}
	results, err := dbColRev.LoadWhere("id = ?", dbColRevID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbColRevision with id %d not found", dbColRevID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}

	dbColRev = results[0]
	return dbColRev, nil
}

// GetID returns the ID of the DbColRevision
func (d *DbColRevision) GetID() uint {
	return d.ID
}

// GetDbCollectionID returns the ID of the associated DbCollection
func (d *DbColRevision) GetDbCollectionID() (uint, error) {
	if d.DbCollectionID == 0 {
		return 0, fmt.Errorf("DbCollectionID is 0, unexpected")
	}
	return d.DbCollectionID, nil
}

// GetDbRevStates returns all associated DbRevStates
func (d *DbColRevision) GetDbRevStates() ([]*DbRevState, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}

	// Reload the associated DbRevStates
	dbRevStates := []*DbRevState{}
	dbRevState := &DbRevState{}
	dbRevStatesResults, err := dbRevState.LoadWhere("db_col_revision_id = ?", d.ID)
	if err != nil {
		return nil, err
	}

	// use DbRevStateLoad() to assure loading of DbRevState with all its fields
	for i, dbRevState := range dbRevStatesResults {
		dbRevStatesResults[i], err = DbRevStateLoad(dbRevState.ID)
		if err != nil {
			return nil, err
		}
	}

	dbRevStates = append(dbRevStates, dbRevStatesResults...)
	d.DbRevStates = dbRevStates

	return d.DbRevStates, nil
}

// GetDbRevStateLatest returns the latest DbRevState or nil if none exists
func (d *DbColRevision) GetDbRevStateLatest() (dbRevState *DbRevState, err error) {
	dbRevStates, err := d.GetDbRevStates()
	if err != nil {
		return nil, err
	}
	if len(dbRevStates) == 0 {
		return nil, nil
	}
	dbRevState = dbRevStates[len(dbRevStates)-1]
	return dbRevState, nil
}

// GetDbRevStateLatestName returns the name of the latest RevState, or "" if none exists
func (d *DbColRevision) GetDbRevStateLatestName() (string, error) {
	latestDbRevState, err := d.GetDbRevStateLatest()
	if err != nil {
		return "", err
	}
	if latestDbRevState == nil {
		return "", nil
	}
	return latestDbRevState.GetRevStateName()
}

// AppendDbRevState adds a new RevState to the collection revision.
// Its a simple wrapper around DbRevStateNew(), enforcing the DbColRevisionID.
func (d *DbColRevision) AppendDbRevState(revStateName string, userWhoTriggered string, logs []byte) error {
	_, err := DbRevStateNew(d.ID, revStateName, userWhoTriggered, logs)
	return err
}

// GetCreationDate returns the creation date of the first RevState
func (d *DbColRevision) GetCreationDate() (time.Time, error) {
	return d.CreatedAt, nil
}

// GetModDate returns the update date of the most-recent RevState
func (d *DbColRevision) GetModDate() (time.Time, error) {
	latestState, err := d.GetDbRevStateLatest()
	if err != nil {
		return time.Time{}, err
	}
	// If no RevState exists, return the DbColRevision's UpdatedAt
	if latestState == nil {
		return d.UpdatedAt, nil
	}

	return latestState.UpdatedAt, nil
}

// IsEditable returns whether the collection revision in the current state can start a collectionEdit
func (d *DbColRevision) IsEditable() (bool, error) {
	return d.IsValidRevStateTransition("CollectionEditOngoing")
}

/*
```
RevStates flow diagram:

	A ColRev-N starts from the ColRev-N-1 "Ready", and then follows transitions which finally ends-up in either "Ready" or "ErrorZZZZ"
	A ColRev-N+1 can only start from a "Ready"-ColRev-N but cannot start from a "ErrorZZZZ"-ColRev-N

	_ColRev-N-1_ ___________ ColRev-N ______________________________________________

	Ready ___                                                                  Ready
	         \__                                                                 A
	            v                                                                |
	             CollectionEditOngoing   --->  CollectionEditCancelled  >------->+
	                   v                                                         |
	             CollectionEditCompleted                                         |
	                   |                                                         |
	                   v                                                         |
	             ProvisioningOngoing     --->  ErrorProvisioningFailed           |
	                   v                                                         A
	             ProvisioningCompleted   >-------------------------------------->+

```
*/
// Define valid transitions based on the flow diagram
// Include in this map keys all the existing states, even if they dont have any transition ("MyStateWithNoTransitions" = {})
// as the map will also be used to detect valid state names
var validStateTransitions = map[string][]string{
	"Ready":                   {"CollectionEditOngoing"},
	"CollectionEditOngoing":   {"CollectionEditCancelled", "CollectionEditCompleted"},
	"CollectionEditCancelled": {"Ready"},
	"CollectionEditCompleted": {"ProvisioningOngoing"},
	"ProvisioningOngoing":     {"ErrorProvisioningFailed", "ProvisioningCompleted"},
	"ProvisioningCompleted":   {"Ready"},
}

func (d *DbColRevision) IsValidRevStateTransition(newStateName string) (isValid bool, err error) {
	var currStateName string
	currStateName, err = d.GetDbRevStateLatestName()
	if err != nil {
		return false, err
	}

	if currStateName == "" {
		currStateName = "Ready"
	}

	if validNextStates, exists := validStateTransitions[currStateName]; exists {
		for _, validNextState := range validNextStates {
			if validNextState == newStateName {
				return true, nil
			}
		}
	}
	return false, nil
}
