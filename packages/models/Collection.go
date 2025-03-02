package models

import (
	"fmt"
	"time"
	"vendingMaxine/packages/models/dbModels"
)

type Collection struct {
	dbIfc dbModels.DbCollectionIfc // unexported field, only used by Collection package and not other packages
}

// Constructor creates dbCollectionIfc and public-methods use dbCollectionIfc to access r/w data

func CollectionNew(
	collectionName string,
	adminUsers []string, adminGroups []string,
	readerUsers []string, readerGroups []string,
	userWhoTriggered string,
) (*Collection, error) {
	// create DbCollection into dbCollectionIfc and return Collection
	c := &Collection{}
	var err error
	c.dbIfc, err = dbModels.DbCollectionNew(collectionName, adminUsers, adminGroups, readerUsers, readerGroups, userWhoTriggered)
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

func (c *Collection) GetAccessPolicy() (*AccessPolicy, error) {
	dbAP, err := c.dbIfc.GetDbAccessPolicy()
	if err != nil {
		return nil, err
	}
	dbAPIDuint := dbAP.GetID()
	apIDuint := dbAPIDuint
	apID := Collection_convert_IDuint_2_ID(apIDuint)
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
	dbColRevs, err := c.dbIfc.GetDbColRevisions()
	if err != nil {
		return nil, err
	}

	if len(dbColRevs) == 0 {
		return colRevs, nil
	}

	colRevs = make([]*ColRevision, 0, len(dbColRevs))
	for _, dbColRev := range dbColRevs {
		colRev := &ColRevision{dbIfc: dbColRev}
		colRevs = append(colRevs, colRev)
	}

	return colRevs, nil
}

// if ColRevisionLatest does not exist (happens when collection.ColRevisions is empty array)
// then it returns nil, nil
func (c *Collection) GetColRevisionLatest() (*ColRevision, error) {
	colRevs, err := c.GetColRevisions()
	if err != nil {
		return nil, err
	}

	if len(colRevs) == 0 {
		return nil, nil
	}

	return colRevs[len(colRevs)-1], nil
}

// If ColRevisionLatest does not exist, then returns "", nil
func (c *Collection) GetRevStateLatestName() (string, error) {
	colRevLatest, err := c.GetColRevisionLatest()
	if err != nil {
		return "", err
	}
	if colRevLatest == nil {
		return "", nil
	}
	return colRevLatest.GetRevStateLatestName()
}

func (c *Collection) AppendColRevision(initialRevStateName string, userWhoTriggered string) error {
	return c.dbIfc.AppendDbColRevision(initialRevStateName, userWhoTriggered)
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
