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
	GetDescription() (string, error)
	SetDescription(string) error
	GetDbRevStates() ([]*DbRevState, error)
	GetCreationDate() (time.Time, error)
	GetModDate() (time.Time, error)
	GetDbRevStateLatest() (dbRevState *DbRevState, err error)
	GetDbRevStateLatestName() (string, error)
	GetRevStateLatestUserWhoTriggered() (string, error)
	IsEditable() (bool, error)
	AppendDbRevState(revStateName string, userWhoTriggered string, logs []byte) error
}

// DbColRevision represents a collection revision in the database
type DbColRevision struct {
	gormCrud.GormCrud[DbColRevision]
	DbCollectionID uint // 1DbCollection-to-manyDbColRevisions - Foreign key to DbCollection
	Description    string
	DbRevStates    []*DbRevState // 1DbColRevision-to-manyDbRevStates
}

// DbColRevisionNew creates a new DbColRevision
func DbColRevisionNew(dbCollectionID uint, colRevDescription string, prevColRev_revStateName string, initialRevStateName string, userWhoTriggered string) (dbColRev *DbColRevision, err error) {
	// Validate if transition prevColRev_revStateName -> initialRevStateName is allowed
	isValid, err := isValidRevStateTransition(prevColRev_revStateName, initialRevStateName)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("invalid state transition from <previous-ColRev>-RevState '%s' to <new-ColRev>-RevState '%s'", prevColRev_revStateName, initialRevStateName)
	}

	dbColRev = &DbColRevision{}
	// Set dbColRev.DbCollectionID, needed for DbRevStateNew()
	dbColRev.DbCollectionID = dbCollectionID
	// Set description
	dbColRev.Description = colRevDescription
	// Save the DbColRevision to sets its ID required for DbRevStateNew()
	err = dbColRev.Save(dbColRev)
	if err != nil {
		return nil, err
	}

	// Set dbColRev.DbRevStates (only after validating the state transition!)
	initialDbRevState, err := DbRevStateNew(dbColRev.ID, initialRevStateName, userWhoTriggered)
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

// GetDescription returns the description of the DbColRevision
func (d *DbColRevision) GetDescription() (string, error) {
	if err := d.Reload(d); err != nil {
		return "", err
	}
	return d.Description, nil
}

// SetDescription sets the description of the DbColRevision
func (d *DbColRevision) SetDescription(description string) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.Description = description
	return d.Save(d)
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

// GetRevStateLatestUserWhoTriggered returns the user who triggered the latest RevState, or empty string if none exists
func (d *DbColRevision) GetRevStateLatestUserWhoTriggered() (string, error) {
	latestDbRevState, err := d.GetDbRevStateLatest()
	if err != nil {
		return "", err
	}
	if latestDbRevState == nil {
		return "", nil
	}
	return latestDbRevState.GetUserWhoTriggered()
}

// AppendDbRevState adds a new RevState to the collection revision.
func (d *DbColRevision) AppendDbRevState(newRevStateName string, userWhoTriggered string, logs []byte) error {
	// Validate if transition from the currRevState --> newRevState is allowed
	// currRevState is the latest RevState in this ColRev
	currRevStateName, err := d.GetDbRevStateLatestName()
	if err != nil {
		return err
	}
	isValid, err := isValidRevStateTransition(currRevStateName, newRevStateName)
	if err != nil {
		return err
	}
	if !isValid {
		return fmt.Errorf("invalid state transition inside this ColRev, from current-RevState '%s' to new-RevState '%s'", currRevStateName, newRevStateName)
	}

	// Create the new RevState (only after validating the state transition!)
	_, err = DbRevStateNew(d.ID, newRevStateName, userWhoTriggered)
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
	currRevStateName, err := d.GetDbRevStateLatestName()
	if err != nil {
		return false, err
	}
	return isValidRevStateTransition(currRevStateName, "CollectionEditOngoing")
}

/*
RevStates flow diagram:

    A ColRev-N starts from the ColRev-N-1 "Ready", and then follows transitions which finally ends-up in either "Ready" or "ErrorZZZZ"
    A ColRev-N+1 can only start from a "Ready"-ColRev-N but cannot start from a "ErrorZZZZ"-ColRev-N

            _________ColRev-N-1_____.___________ ColRev-N _________________________________________________.
                                    .                                                                      .

        CollectionEdit FLOW

                        Ready ___   .                                                               Ready  .
                                 \__.__                                                               A    .
                                    .  V                                                              |    .
                                    . CollectionEditOngoing   --->  CollectionEditCancelled  >------->+    .
                                    .       v                                                         |    .
                                    . CollectionEditCompleted                                         |    .
                                    .       |                                                         |    .
                                    .       v                                                         |    .
                                    . ProvisioningOngoing     --->  ErrorProvisioningFailed           |    .
                                    .       v                                                         A    .
                                    . ProvisioningCompleted   >-------------------------------------->+    .


        NewCollectionCreated FLOW

            NewCollectionCreated >--.-------------------------------------------------------------> Ready  .




*/
// Define valid transitions based on the flows diagram
// Include in this map keys all the existing states, even if they dont have any transition ("MyStateWithNoTransitions" = {})
var validStateTransitionsFlows = []map[string][]string{
	{
		// CollectionEdit flow
		"Ready":                   {"CollectionEditOngoing"},
		"CollectionEditOngoing":   {"CollectionEditCancelled", "CollectionEditCompleted"},
		"CollectionEditCancelled": {"Ready"},
		"CollectionEditCompleted": {"ProvisioningOngoing"},
		"ProvisioningOngoing":     {"ErrorProvisioningFailed", "ProvisioningCompleted"},
		"ProvisioningCompleted":   {"Ready"},
	},
	{
		// NewCollectionCreated flow
		"NewCollectionCreated": {"Ready"},
	},
}

func isValidRevStateTransition(currStateName string, newStateName string) (isValid bool, err error) {
	for _, flowMap := range validStateTransitionsFlows {
		if validNextStates, exists := flowMap[currStateName]; exists {
			for _, validNextState := range validNextStates {
				if validNextState == newStateName {
					return true, nil
				}
			}
		}
	}
	return false, nil
}
