package models

import (
	"fmt"
	"time"
	"vendingMaxine/packages/models/dbModels"
)

// RevState represents a revision state
type RevState struct {
	dbIfc dbModels.DbRevStateIfc
}

// RevStateNew creates a new RevState
func RevStateNew(colRevisionID string, revStateName string, userWhoTriggered string, logs []byte) (*RevState, error) {
	colRevisionIDuint, err := ColRevision_convert_ID_2_IDuint(colRevisionID)
	if err != nil {
		return nil, err
	}

	r := &RevState{}
	r.dbIfc, err = dbModels.DbRevStateNew(colRevisionIDuint, revStateName, userWhoTriggered, logs)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// RevStateLoad loads a RevState by ID
func RevStateLoad(revStateID string) (*RevState, error) {
	revStateIDuint, err := convert_IDstring_2_IDuint(revStateID)
	if err != nil {
		return nil, err
	}

	r := &RevState{}
	r.dbIfc, err = dbModels.DbRevStateLoad(revStateIDuint)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// RevState_convert_ID_2_IDuint converts a RevState ID string to the underlying uint ID
func RevState_convert_ID_2_IDuint(ID string) (IDuint uint, err error) {
	return convert_IDstring_2_IDuint(ID)
}

// RevState_convert_IDuint_2_ID converts a uint ID to a RevState ID string
func RevState_convert_IDuint_2_ID(IDuint uint) (ID string) {
	return convert_IDuint_2_IDstring("RevStateID", IDuint)
}

// GetID returns the ID of the RevState as a string
func (r *RevState) GetID() (ID string, err error) {
	IDuint := r.dbIfc.GetID()
	if IDuint == 0 {
		return "", fmt.Errorf("IDuint is 0, ?maybe revision state does not exist in db?")
	}
	ID = RevState_convert_IDuint_2_ID(IDuint)
	return ID, nil
}

// GetColRevisionID returns the ID of the associated ColRevision
func (r *RevState) GetColRevisionID() (string, error) {
	dbColRevisionID, err := r.dbIfc.GetDbColRevisionID()
	if err != nil {
		return "", err
	}
	return ColRevision_convert_IDuint_2_ID(dbColRevisionID), nil
}

// RevStateName returns the name of the RevState
func (r *RevState) RevStateName() (string, error) {
	return r.dbIfc.GetRevStateName()
}

// CreationDate returns the creation date of the RevState
func (r *RevState) CreationDate() (time.Time, error) {
	return r.dbIfc.GetCreatedAt()
}

// IsEditable returns whether the collection is editable in this state
func (r *RevState) IsEditable() (bool, error) {
	return r.dbIfc.IsEditable()
}

// UserWhoTriggered returns the user who triggered this state
func (r *RevState) UserWhoTriggered() (string, error) {
	return r.dbIfc.GetUserWhoTriggered()
}

// GetLogs returns the logs for ProvisioningFailed and ProvisioningCompleted states
func (r *RevState) GetLogs() ([]byte, error) {
	return r.dbIfc.GetLogs()
}

// GetColRevision returns the ColRevision this state belongs to
func (r *RevState) GetColRevision() (*ColRevision, error) {
	colRevisionID, err := r.GetColRevisionID()
	if err != nil {
		return nil, err
	}
	return ColRevisionLoad(colRevisionID)
}
