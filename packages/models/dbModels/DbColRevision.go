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
func DbColRevisionNew(dbCollectionID uint) (*DbColRevision, error) {
	dbColRev := &DbColRevision{
		DbCollectionID: dbCollectionID,
	}
	err := dbColRev.Save(dbColRev)
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
	// For now its better to load the associated DbRevStates on demand,
	// by using .GetDbRevStates()
	/*
		// Load associated DbRevStates
		dbRevStates := []*DbRevState{}
		dbRevState := &DbRevState{}
		dbRevStatesResults, err := dbRevState.LoadWhere("db_col_revision_id = ?", dbColRev.ID)
		if err != nil {
			return nil, err
		}
		dbRevStates = append(dbRevStates, dbRevStatesResults...)
		dbColRev.DbRevStates = dbRevStates
	*/

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

// GetCreationDate returns the creation date of the first RevState
func (d *DbColRevision) GetCreationDate() (time.Time, error) {
	return d.CreatedAt, nil
}

// GetModDate returns the update date of the most-recent RevState
func (d *DbColRevision) GetModDate() (time.Time, error) {
	latestState, err := d.GetDbRevStateLatest()
	if err != nil {
		return d.UpdatedAt, nil // Fall back to revision's update time if no states exist
	}
	return latestState.UpdatedAt, nil
}

// GetDbRevStateLatest returns the latest DbRevState
func (d *DbColRevision) GetDbRevStateLatest() (dbRevState *DbRevState, err error) {
	dbRevStates, err := d.GetDbRevStates()
	if err != nil {
		return nil, err
	}
	if len(dbRevStates) == 0 {
		return nil, fmt.Errorf("no DbRevStates found for DbColRevision %d", d.ID)
	}
	dbRevState = dbRevStates[len(dbRevStates)-1]
	return dbRevState, nil
}

// GetDbRevStateLatestName returns the name of the latest RevState
func (d *DbColRevision) GetDbRevStateLatestName() (string, error) {
	latestDbRevState, err := d.GetDbRevStateLatest()
	if err != nil {
		return "", err
	}
	return latestDbRevState.GetRevStateName()
}

// IsEditable returns whether the collection revision is editable
func (d *DbColRevision) IsEditable() (bool, error) {
	revStates, err := d.GetDbRevStates()
	if err != nil {
		return false, err
	}
	if len(revStates) == 0 {
		return false, fmt.Errorf("no RevStates found for ColRevision %d", d.ID)
	}
	latestRevState := revStates[len(revStates)-1]
	return latestRevState.IsEditable()
}

// AppendDbRevState adds a new RevState to the collection revision.
// Its a simple wrapper around DbRevStateNew() which sets the DbColRevisionID.
func (d *DbColRevision) AppendDbRevState(revStateName string, userWhoTriggered string, logs []byte) error {
	_, err := DbRevStateNew(d.ID, revStateName, userWhoTriggered, logs)
	return err
}

// IsValidRevStateTransition checks if a state transition is allowed
func (d *DbColRevision) IsValidRevStateTransition(newStateName string) (isValid bool, err error) {
	var currStateName string
	currStateName, err = d.GetDbRevStateLatestName()
	if err != nil {
		return false, err
	}

	// Define valid transitions based on the flow diagram
	validTransitions := map[string][]string{
		"CollectionEditOngoing":   {"CollectionEditCancelled", "CollectionEditCompleted"},
		"CollectionEditCompleted": {"ProvisioningOngoing"},
		"ProvisioningOngoing":     {"ProvisioningFailed", "ProvisioningCompleted"},
	}

	if validNextStates, exists := validTransitions[currStateName]; exists {
		for _, validState := range validNextStates {
			if validState == newStateName {
				return true, nil
			}
		}
	}
	return false, nil
}
