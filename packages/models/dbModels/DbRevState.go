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
	GetCreationDate() (time.Time, error)
	GetUserWhoTriggered() (string, error)
}

// DbRevState represents a revision state in the database
type DbRevState struct {
	gormCrud.GormCrud[DbRevState]
	DbColRevisionID  uint
	RevStateName     string // One of the defined state names
	UserWhoTriggered string
}

// DbRevStateNew creates a new DbRevState
// NOTE: this constructor does now know about previous state and does not validate state transitions - that should be done beforehand by the caller
func DbRevStateNew(dbColRevisionID uint, revStateName string, userWhoTriggered string) (*DbRevState, error) {

	dbRevState := &DbRevState{
		DbColRevisionID:  dbColRevisionID,
		RevStateName:     revStateName,
		UserWhoTriggered: userWhoTriggered,
	}

	err := dbRevState.Save(dbRevState)
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
	return d.DbColRevisionID, nil
}

// GetRevStateName returns the name of the RevState
func (d *DbRevState) GetRevStateName() (string, error) {
	return d.RevStateName, nil
}

// GetCreatedAt returns the creation date of the RevState
func (d *DbRevState) GetCreationDate() (time.Time, error) {
	return d.CreatedAt, nil
}

// GetUserWhoTriggered returns the user who triggered this state
func (d *DbRevState) GetUserWhoTriggered() (string, error) {
	return d.UserWhoTriggered, nil
}
