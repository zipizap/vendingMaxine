package dbModels

import (
	"fmt"
	"time"
	"vendingMaxine/packages/gormCrud"
)

// DbCollection follows gorm conventions, takes care of db-struct and db-methods

type DbCollectionIfc interface {
	GetID() uint
	GetName() (string, error)
	SetName(string) error
	GetDbAccessPolicy() (dbAP *DbAccessPolicy, err error)
	GetDbColRevisions() (dbColRevs []*DbColRevision, err error)
	GetDbColRevisionLatest() (dbColRevLatest *DbColRevision, err error)
	AppendDbColRevision(initialRevStateName string, userWhoTriggered string) error
	GetCreationDate() (time.Time, error)
	GetModDate() (time.Time, error)
	GetRevStateLatestName() (string, error)
	GetRevStateLatestUserWhoTriggered() (string, error)
}

type DbCollection struct {
	gormCrud.GormCrud[DbCollection]
	Name           string
	DbAccessPolicy *DbAccessPolicy  //this might be unnecessary... `gorm:"foreignKey:DbCollectionID;references:ID"` // 1DbAccessPolicy-to-1DbCollection
	DbColRevisions []*DbColRevision //this might be unnecessary... `gorm:"foreignKey:DbCollectionID;references:ID"` // 1DbCollection-to-manyDbColRevisions, loaded on demand by GetDbColRevisions()
}

func DbCollectionNew(
	name string,
	adminUsers []string, adminGroups []string,
	readerUsers []string, readerGroups []string,
	userWhoTriggered string,
) (*DbCollection, error) {
	dbC := &DbCollection{}
	// Col.Name
	dbC.Name = name
	// Saving sets the dbC.ID, needed for DbColRevisionNew()
	err := dbC.Save(dbC)
	if err != nil {
		return nil, err
	}

	// A new DbCollection always starts with a new DbColRevision in state "Ready"
	// Col.DbColRevisions[]
	prevColRev_revStateName := "NewCollectionCreated"
	initialRevStateName := "Ready"
	dbColRev, err := DbColRevisionNew(dbC.ID, prevColRev_revStateName, initialRevStateName, userWhoTriggered)
	if err != nil {
		// Delete the DbCollection if creating the revision failed
		if err := dbC.Delete(dbC); err != nil {
			// Log the error but continue with the original error
			fmt.Printf("Failed to clean up DbCollection after previous error: %v\n", err)
		}
		return nil, err
	}
	dbC.DbColRevisions = append(dbC.DbColRevisions, dbColRev)

	// Col.DbAccessPolicy
	dbC.DbAccessPolicy, err = DbAccessPolicyNew(dbC.ID, adminUsers, adminGroups, readerUsers, readerGroups)
	if err != nil {
		return nil, err
	}

	// final save
	err = dbC.Save(dbC)
	if err != nil {
		return nil, err
	}
	return dbC, nil
}

func DbCollectionLoad(dbColId uint) (*DbCollection, error) {
	dbC := &DbCollection{}
	results, err := dbC.LoadWhere("id = ?", dbColId)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbCollection with id %d not found", dbColId)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}

	dbC = results[0]
	return dbC, nil
}

func (d *DbCollection) GetID() uint {
	return d.ID
}

func (d *DbCollection) GetName() (string, error) {
	// Name might change, so we always reload it from the database
	if err := d.Reload(d); err != nil {
		return "", err
	}
	return d.Name, nil
}

func (d *DbCollection) SetName(newName string) error {
	// Reload before Saving: to assure we have the latest data from db
	{
		err := d.Reload(d)
		if err != nil {
			return err
		}
	}
	d.Name = newName
	return d.Save(d)
}

func (d *DbCollection) GetDbAccessPolicy() (dbAP *DbAccessPolicy, err error) {
	dbAP, err = DbAccessPolicyLoad(d.DbAccessPolicy.ID)
	return dbAP, err
}

func (d *DbCollection) GetDbColRevisions() (dbColRevs []*DbColRevision, err error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}

	// Reload the associated DbColRevisions
	dbColRevs = []*DbColRevision{}
	dbColRev := &DbColRevision{}
	dbColRevsResults, err := dbColRev.LoadWhere("db_collection_id = ?", d.ID)
	if err != nil {
		return nil, err
	}

	// use DbColRevisionLoad() to assure loading of DbColRevision with all its fields
	for i, a_dbColRev := range dbColRevsResults {
		dbColRevsResults[i], err = DbColRevisionLoad(a_dbColRev.ID)
		if err != nil {
			return nil, err
		}
	}

	if len(dbColRevsResults) > 0 {
		dbColRevs = append(dbColRevs, dbColRevsResults...)
	}
	d.DbColRevisions = dbColRevs

	return d.DbColRevisions, nil
}

// GetDbColRevisionLatest returns the latest DbColRevision
// If there are no DbColRevisions, it returns nil
func (d *DbCollection) GetDbColRevisionLatest() (dbColRevLatest *DbColRevision, err error) {
	// Get all the dbColRevs
	dbColRevs, err := d.GetDbColRevisions()
	if err != nil {
		return nil, err
	}

	// If there are no dbColRevs()), return nil
	if len(dbColRevs) == 0 {
		return nil, nil
	}

	// Return the latest-dbColRev (last in array)
	return dbColRevs[len(dbColRevs)-1], nil
}

func (d *DbCollection) AppendDbColRevision(initialRevStateName string, userWhoTriggered string) error {
	prevColRev_revStateName, err := d.GetRevStateLatestName()
	if err != nil {
		return err
	}
	_, err = DbColRevisionNew(d.ID, prevColRev_revStateName, initialRevStateName, userWhoTriggered)
	return err
}

func (d *DbCollection) GetCreationDate() (time.Time, error) {
	if err := d.Reload(d); err != nil {
		return time.Time{}, err
	}
	return d.CreatedAt, nil
}

func (d *DbCollection) GetModDate() (time.Time, error) {
	if err := d.Reload(d); err != nil {
		return time.Time{}, err
	}
	return d.UpdatedAt, nil
}

// GetRevStateLatestName returns the name of the latest-RevState from the latest-ColRev of this collection.
// If there is no ColRev(s), it returns an empty string.
func (d *DbCollection) GetRevStateLatestName() (string, error) {
	// Get all ColRevs
	dbColRevs, err := d.GetDbColRevisions()
	if err != nil {
		return "", err
	}

	// If there are no ColRevs, return empty string
	if len(dbColRevs) == 0 {
		return "", nil
	}

	// Get the latest ColRev (last in array)
	latestDbColRev := dbColRevs[len(dbColRevs)-1]

	// Return the name of its current state
	return latestDbColRev.GetDbRevStateLatestName()
}

// GetRevStateLatestUserWhoTriggered returns the user who triggered the latest RevState from the latest ColRev.
// If there is no ColRev(s), it returns an empty string.
func (d *DbCollection) GetRevStateLatestUserWhoTriggered() (string, error) {
	// Get all ColRevs
	dbColRevs, err := d.GetDbColRevisions()
	if err != nil {
		return "", err
	}

	// If there are no ColRevs, return empty string
	if len(dbColRevs) == 0 {
		return "", nil
	}

	// Get the latest ColRev (last in array)
	latestDbColRev := dbColRevs[len(dbColRevs)-1]

	// Return the user who triggered its current state
	return latestDbColRev.GetRevStateLatestUserWhoTriggered()
}
