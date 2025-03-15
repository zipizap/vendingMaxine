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

// CollectionsList returns the list of collections for the user and groups
// It returns two lists: readerCols and adminCols
// readerCols are the collections that the user can read
// adminCols are the collections that the user can edit
// If there are no collections, it returns empty lists
func CollectionsList(user string, groups []string) (readerCols []*Collection, adminCols []*Collection, err error) {
	// Get the list of collections
	readerDbCols, adminDbCols, err := dbModels.DbCollectionsList(user, groups)
	if err != nil {
		return nil, nil, err
	}

	// Convert to readerCols and adminCols
	readerCols = make([]*Collection, 0, len(readerDbCols))
	adminCols = make([]*Collection, 0, len(adminDbCols))
	for _, dbCol := range readerDbCols {
		var col *Collection
		{
			dbColIDuint := dbCol.GetID()
			colID := Collection_convert_IDuint_2_ID(dbColIDuint)
			col, err = CollectionLoad(colID)
			if err != nil {
				return nil, nil, err
			}
		}
		readerCols = append(readerCols, col)
	}

	for _, dbCol := range adminDbCols {
		var col *Collection
		{
			dbColIDuint := dbCol.GetID()
			colID := Collection_convert_IDuint_2_ID(dbColIDuint)
			col, err = CollectionLoad(colID)
			if err != nil {
				return nil, nil, err
			}
		}
		adminCols = append(adminCols, col)
	}

	return readerCols, adminCols, nil
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

func (c *Collection) GetRole(user string, groups []string) (role string, err error) {
	return c.dbIfc.GetRole(user, groups)
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

// GetModDate returns the last modified date of the collection
// Returns the mod date of the latest ColRev if it exists, otherwise returns the mod date of the collection
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
func (c *Collection) DoCollectionEditOngoing(colRevDescription string, userWhoTriggered string) error {
	// Create a new colRev with initialStateName "CollectionEditOngoing"
	initialStateName := "CollectionEditOngoing"
	return c.appendColRevision(colRevDescription, initialStateName, userWhoTriggered)
}

func (c *Collection) DoCollectionEditCancelled(userWhoTriggered string) error {
	// Advance existing colRev to new state "CollectionEditCancelled"
	{
		newRevStateName := "CollectionEditCancelled"
		err := c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
		if err != nil {
			return err
		}
	}

	// Do things for CollectionEditCancelled
	// ...none for now...

	// Advance existing colRev from CollectionEditCancelled to Ready
	{
		err := c.DoCollectionReady(userWhoTriggered)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) DoCollectionReady(userWhoTriggered string) error {
	// Use existing ColRev, and advance it to new state "Ready"
	newRevStateName := "Ready"
	return c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
}

// appendRevState appends a new RevState to the ColRevLatest
// Its the way to progress a existing ColRev to a new RevState
// It implicitly validates if it is possible to transit from whatever-current-state to the proposed newRevStateName
func (c *Collection) appendRevStateToColRevLatest(newRevStateName string, userWhoTriggered string) error {
	colRevLatest, err := c.GetColRevisionLatest()
	if err != nil {
		return err
	}
	if colRevLatest == nil {
		return fmt.Errorf("no ColRevision exists yet")
	}
	return colRevLatest.AppendRevState(newRevStateName, userWhoTriggered)
}

func (c *Collection) DoCollectionEditCompleted(userWhoTriggered string) error {
	// Advance existing colRev to new state "CollectionEditCompleted"
	{
		newRevStateName := "CollectionEditCompleted"
		err := c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
		if err != nil {
			return err
		}
	}

	// Process any required operations for CollectionEditCompleted
	// ...none for now...

	// Advance existing colRev from CollectionEditCompleted to ProvisioningOngoing
	{
		err := c.DoCollectionProvisioningOngoing(userWhoTriggered)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) DoCollectionProvisioningOngoing(userWhoTriggered string) error {
	// Advance existing colRev to new state "ProvisioningOngoing"
	{
		newRevStateName := "ProvisioningOngoing"
		err := c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
		if err != nil {
			return err
		}
	}

	// Process any required operations for ProvisioningOngoing
	// ...none for now...

	// Advance existing colRev from ProvisioningOngoing to ProvisioningCompleted
	{
		err := c.DoCollectionProvisioningCompleted(userWhoTriggered)
		if err != nil {
			return err
		}
	}

	return nil

}

func (c *Collection) DoCollectionProvisioningCompleted(userWhoTriggered string) error {
	// Advance existing colRev to new state "ProvisioningCompleted"
	{
		newRevStateName := "ProvisioningCompleted"
		err := c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
		if err != nil {
			return err
		}
	}

	// Process any required operations for ProvisioningCompleted
	// ...none for now...

	// Advance existing colRev from ProvisioningCompleted to Ready
	{
		err := c.DoCollectionReady(userWhoTriggered)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Collection) DoCollectionErrorProvisioningFailed(userWhoTriggered string) error {
	// Advance existing colRev to new state "ErrorProvisioningFailed"
	{
		newRevStateName := "ErrorProvisioningFailed"
		err := c.appendRevStateToColRevLatest(newRevStateName, userWhoTriggered)
		if err != nil {
			return err
		}
	}

	// Process any required operations for ErrorProvisioningFailed
	// ...none for now...

	return nil
}
