package models

import (
	"fmt"
	"time"
	"vendingMaxine/packages/models/dbModels"
)

// ColRevision represents a collection revision
type ColRevision struct {
	dbIfc dbModels.DbColRevisionIfc // unexported field, only used by ColRevision package and not other packages
}

// ColRevisionNew creates a new ColRevision
func ColRevisionNew(collectionID string) (*ColRevision, error) {
	collectionIDuint, err := Collection_convert_ID_2_IDuint(collectionID)
	if err != nil {
		return nil, err
	}

	c := &ColRevision{}
	c.dbIfc, err = dbModels.DbColRevisionNew(collectionIDuint)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ColRevisionLoad loads a ColRevision by ID
func ColRevisionLoad(colRevID string) (*ColRevision, error) {
	colRevIDuint, err := convert_IDstring_2_IDuint(colRevID)
	if err != nil {
		return nil, err
	}

	c := &ColRevision{}
	c.dbIfc, err = dbModels.DbColRevisionLoad(colRevIDuint)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ColRevision_convert_ID_2_IDuint converts a ColRevision ID string to the underlying uint ID
func ColRevision_convert_ID_2_IDuint(ID string) (IDuint uint, err error) {
	return convert_IDstring_2_IDuint(ID)
}

// ColRevision_convert_IDuint_2_ID converts a uint ID to a ColRevision ID string
func ColRevision_convert_IDuint_2_ID(IDuint uint) (ID string) {
	return convert_IDuint_2_IDstring("ColRevID", IDuint)
}

// GetID returns the ID of the ColRevision as a string
func (c *ColRevision) GetID() (ID string, err error) {
	IDuint := c.dbIfc.GetID()
	if IDuint == 0 {
		return "", fmt.Errorf("IDuint is 0, ?maybe collection revision does not exist in db?")
	}
	ID = ColRevision_convert_IDuint_2_ID(IDuint)
	return ID, nil
}

// GetCollectionID returns the ID of the associated Collection
func (c *ColRevision) GetCollectionID() (string, error) {
	dbCollectionID, err := c.dbIfc.GetDbCollectionID()
	if err != nil {
		return "", err
	}
	return Collection_convert_IDuint_2_ID(dbCollectionID), nil
}

// GetRevStates returns all associated GetRevStates
func (c *ColRevision) GetRevStates() ([]*RevState, error) {
	dbRevStates, err := c.dbIfc.GetDbRevStates()
	if err != nil {
		return nil, err
	}

	revStates := make([]*RevState, 0, len(dbRevStates))
	for _, dbRevState := range dbRevStates {
		revState := &RevState{dbIfc: dbRevState}
		revStates = append(revStates, revState)
	}

	return revStates, nil
}

// GetRevStateLatest returns the latest RevState
func (c *ColRevision) GetRevStateLatest() (*RevState, error) {
	revStates, err := c.GetRevStates()
	if err != nil {
		return nil, err
	}

	if len(revStates) == 0 {
		return nil, fmt.Errorf("no RevStates found for ColRevision")
	}

	return revStates[len(revStates)-1], nil
}

// GetCreationDate returns the creation date of the first RevState
func (c *ColRevision) GetCreationDate() (time.Time, error) {
	return c.dbIfc.GetCreationDate()
}

// GetModDate returns the creation date of the latest RevState
func (c *ColRevision) GetModDate() (time.Time, error) {
	return c.dbIfc.GetModDate()
}

// IsEditable returns whether the collection revision is editable
func (c *ColRevision) IsEditable() (bool, error) {
	return c.dbIfc.IsEditable()
}

// GetRevStateLatestName returns the name of the latest RevState
func (c *ColRevision) GetRevStateLatestName() (string, error) {
	return c.dbIfc.GetDbRevStateLatestName()
}

// AppendRevState adds a new RevState to the collection revision
func (c *ColRevision) AppendRevState(revStateName string, userWhoTriggered string, logs []byte) error {
	return c.dbIfc.AppendDbRevState(revStateName, userWhoTriggered, logs)
}

// GetCollection returns the Collection this revision belongs to
func (c *ColRevision) GetCollection() (*Collection, error) {
	collectionID, err := c.GetCollectionID()
	if err != nil {
		return nil, err
	}
	return CollectionLoad(collectionID)
}
