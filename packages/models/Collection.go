package models

import (
	"fmt"
	"time"
	"vendingMaxine/packages/models/dbModels"
	"vendingMaxine/packages/sharedTypes"
)

type Collection struct {
	dbIfc dbModels.DbCollectionIfc // unexported field, only used by Collection package and not other packages
}

// CollectionNew creates a new Collection
// It also creates a new AccessPolicy for the Collection from accessPolicyParams. Remember to include the creator-user in the AdminUsers beforehand!
func CollectionNew(
	collectionName string,
	description string, // New parameter
	accessPolicyParams sharedTypes.AccessPolicyParams,
	userWhoTriggered string,
) (*Collection, error) {
	// create DbCollection into dbCollectionIfc and return Collection
	c := &Collection{}
	var err error
	c.dbIfc, err = dbModels.DbCollectionNew(
		collectionName,
		description, // Pass description to DbCollectionNew
		accessPolicyParams,
		userWhoTriggered)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Ex: col, err := LoadCollection("ColID-1234")
func CollectionLoad(colID string) (*Collection, error) {
	// load DbCollection from dbCollectionIfc and return Collection
	colIDuint, err := convert_IDstring_2_IDuint(colID)
	if err != nil {
		return nil, err
	}
	c := &Collection{}
	c.dbIfc, err = dbModels.DbCollectionLoad(colIDuint)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Collection_convert_ID_2_IDuint(ID string) (IDuint uint, err error) {
	return convert_IDstring_2_IDuint(ID)
}
func Collection_convert_IDuint_2_ID(IDuint uint) (ID string) {
	return convert_IDuint_2_IDstring("CollectionID", IDuint)
}

// Ex: "ColID-1234"
func (o *Collection) GetID() (ID string, err error) {
	IDuint := o.dbIfc.GetID()
	if IDuint == 0 {
		return "", fmt.Errorf("IDuint is 0, ?maybe collection does not exist in db?")
	}
	ID = Collection_convert_IDuint_2_ID(IDuint)
	return ID, nil
}

func (c *Collection) GetName() (string, error) {
	return c.dbIfc.GetName()
}

func (c *Collection) GetDescription() (string, error) {
	return c.dbIfc.GetDescription()
}

func (c *Collection) SetDescription(newDescription string) error {
	return c.dbIfc.SetDescription(newDescription)
}

func (c *Collection) GetAccessPolicy() (*AccessPolicy, error) {
	dbAP, err := c.dbIfc.GetDbAccessPolicy()
	if err != nil {
		return nil, err
	}
	dbAPIDuint := dbAP.GetID()
	apIDuint := dbAPIDuint
	apID := AccessPolicy_convert_IDuint_2_ID(apIDuint)
	var ap *AccessPolicy
	ap, err = AccessPolicyLoad(apID)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func (c *Collection) Rename(newName string) error {
	return c.dbIfc.SetName(newName)
}

func (c *Collection) GetColRevisions() (colRevs []*ColRevision, err error) {
	// get dbColRevs
	dbColRevs, err := c.dbIfc.GetDbColRevisions()
	if err != nil {
		return nil, err
	}

	if len(dbColRevs) == 0 {
		return colRevs, nil
	}

	// convert dbColRevs to colRevs
	colRevs = make([]*ColRevision, 0, len(dbColRevs))
	for _, dbColRev := range dbColRevs {
		colRev := &ColRevision{dbIfc: dbColRev}
		colRevs = append(colRevs, colRev)
	}

	return colRevs, nil
}

// GetColRevisionLatest returns the latest ColRevision
// If ColRevisionLatest does not exist, it returns nil, nil
func (c *Collection) GetColRevisionLatest() (*ColRevision, error) {
	dbColRevLatest, err := c.dbIfc.GetDbColRevisionLatest()
	if err != nil {
		return nil, err
	}

	// if there was no ColRevision(s), then return nil
	if dbColRevLatest == nil {
		return nil, nil
	}

	return &ColRevision{dbIfc: dbColRevLatest}, nil
}

// GetRevStateLatestName returns the name of the latest-RevState from the latest-ColRev
// If there are no ColRevs, then returns "", nil
func (c *Collection) GetRevStateLatestName() (string, error) {
	return c.dbIfc.GetRevStateLatestName()
}

// GetRevStateLatestUserWhoTriggered returns the user who triggered the latest RevState from the latest ColRev
// If there are no ColRevs, then returns "", nil
func (c *Collection) GetRevStateLatestUserWhoTriggered() (string, error) {
	return c.dbIfc.GetRevStateLatestUserWhoTriggered()
}

// appendColRevision appends a new ColRevision to the collection.
// It implicitly validates if it is possible to transit from whatever-current-state to the proposed initialRevStateName
// It's private method, to be used by other public methods of this package, like Do_CollectionEdit()
func (c *Collection) appendColRevision(colRevDescription string, initialRevStateName string, userWhoTriggered string) error {
	return c.dbIfc.AppendDbColRevision(colRevDescription, initialRevStateName, userWhoTriggered)
}

func (c *Collection) GetCreationDate() (time.Time, error) {
	return c.dbIfc.GetCreationDate()
}

func (c *Collection) GetModDate() (time.Time, error) {
	colRevLatest, err := c.GetColRevisionLatest()
	if err != nil {
		return time.Time{}, err
	}

	if colRevLatest == nil {
		return c.dbIfc.GetModDate()
	}

	return colRevLatest.GetModDate()
}

// IsEditable returns whether the collection in the current state can start a collectionEdit
func (c *Collection) IsEditable() (bool, error) {
	colRevLatest, err := c.GetColRevisionLatest()
	if err != nil {
		return false, err
	}

	// If there's no revision yet, default to not editable
	if colRevLatest == nil {
		return false, nil
	}

	// Get the latest revision state and check if it's editable
	return colRevLatest.IsEditable()
}

// Do_CollectionEdit starts a collectionEdit
func (c *Collection) DoCollectionEdit(colRevDescription string, userWhoTriggered string) error {
	initialStateName := "CollectionEditOngoing"
	return c.appendColRevision(colRevDescription, initialStateName, userWhoTriggered)
}
