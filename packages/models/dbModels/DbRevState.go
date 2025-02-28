package dbModels

import (
	"fmt"
	"time"
	"vendingMaxine/packages/gormCrud"
)

// DbRevStateIfc defines the interface for DbRevState operations
type DbRevStateIfc interface {
	GetID() uint
	GetDbColRevisionID() (uint, error)
	GetRevStateName() (string, error)
	GetCreatedAt() (time.Time, error)
	IsEditable() (bool, error)
	GetUserWhoTriggered() (string, error)
	GetLogs() ([]byte, error)
}

// DbRevState represents a revision state in the database
type DbRevState struct {
	gormCrud.GormCrud[DbRevState]
	DbColRevisionID  uint
	RevStateName     string // One of the defined state names
	UserWhoTriggered string
	Logs             []byte // For storing logs in ProvisioningFailed and ProvisioningCompleted states
}

// DbRevStateNew creates a new DbRevState
func DbRevStateNew(dbColRevisionID uint, revStateName string, userWhoTriggered string, logs []byte) (*DbRevState, error) {
	// Validate RevStateName
	if !isValidRevStateName(revStateName) {
		return nil, fmt.Errorf("invalid RevStateName: %s", revStateName)
	}

	// Validate state-transition with DbColRevision.IsValidRevStateTransition()
	dbColRev, err := DbColRevisionLoad(dbColRevisionID)
	if err != nil {
		return nil, err
	}
	isValid, err := dbColRev.IsValidRevStateTransition(revStateName)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, fmt.Errorf("invalid state transition in ColRev %d to new state %s", dbColRevisionID, revStateName)
	}

	dbRevState := &DbRevState{
		DbColRevisionID:  dbColRevisionID,
		RevStateName:     revStateName,
		UserWhoTriggered: userWhoTriggered,
	}

	// Only set logs for states that need them
	if revStateName == "ProvisioningFailed" || revStateName == "ProvisioningCompleted" {
		dbRevState.Logs = logs
	}

	err = dbRevState.Save(dbRevState)
	if err != nil {
		return nil, err
	}
	return dbRevState, nil
}

// DbRevStateLoad loads a DbRevState by ID
func DbRevStateLoad(dbRevStateID uint) (*DbRevState, error) {
	dbRevState := &DbRevState{}
	results, err := dbRevState.LoadWhere("id = ?", dbRevStateID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbRevState with id %d not found", dbRevStateID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}

	dbRevState = results[0]
	return dbRevState, nil
}

// GetID returns the ID of the DbRevState
func (d *DbRevState) GetID() uint {
	return d.ID
}

// GetDbColRevisionID returns the ID of the associated DbColRevision
func (d *DbRevState) GetDbColRevisionID() (uint, error) {
	if err := d.Reload(d); err != nil {
		return 0, err
	}
	return d.DbColRevisionID, nil
}

// GetRevStateName returns the name of the RevState
func (d *DbRevState) GetRevStateName() (string, error) {
	if err := d.Reload(d); err != nil {
		return "", err
	}
	return d.RevStateName, nil
}

// GetCreatedAt returns the creation date of the RevState
func (d *DbRevState) GetCreatedAt() (time.Time, error) {
	if err := d.Reload(d); err != nil {
		return time.Time{}, err
	}
	return d.CreatedAt, nil
}

// IsEditable returns whether the collection is editable in this state
func (d *DbRevState) IsEditable() (bool, error) {
	if err := d.Reload(d); err != nil {
		return false, err
	}

	// Only CollectionEditOngoing state is editable
	return d.RevStateName == "CollectionEditOngoing", nil
}

// GetUserWhoTriggered returns the user who triggered this state
func (d *DbRevState) GetUserWhoTriggered() (string, error) {
	if err := d.Reload(d); err != nil {
		return "", err
	}
	return d.UserWhoTriggered, nil
}

// GetLogs returns the logs for ProvisioningFailed and ProvisioningCompleted states
func (d *DbRevState) GetLogs() ([]byte, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}

	// Logs are only available for ProvisioningFailed and ProvisioningCompleted states
	if d.RevStateName != "ProvisioningFailed" && d.RevStateName != "ProvisioningCompleted" {
		return nil, fmt.Errorf("logs are not available for state %s", d.RevStateName)
	}

	return d.Logs, nil
}

// isValidRevStateName checks if the provided state name is valid
func isValidRevStateName(name string) bool {
	validNames := map[string]bool{
		"CollectionEditOngoing":   true,
		"CollectionEditCancelled": true,
		"CollectionEditCompleted": true,
		"ProvisioningOngoing":     true,
		"ProvisioningFailed":      true,
		"ProvisioningCompleted":   true,
	}
	return validNames[name]
}
