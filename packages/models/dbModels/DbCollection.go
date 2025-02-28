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
	AppendDbColRevision() error
	GetCreationDate() (time.Time, error)
	GetModDate() (time.Time, error)
}

type DbCollection struct {
	gormCrud.GormCrud[DbCollection]
	Name           string
	DbAccessPolicy *DbAccessPolicy  // 1DbAccessPolicy-to-1DbCollection
	DbColRevisions []*DbColRevision // 1DbCollection-to-manyDbColRevisions, loaded on demand by GetDbColRevisions()
}

func DbCollectionNew(
	name string,
	adminUsers []string, adminGroups []string,
	readerUsers []string, readerGroups []string,
) (*DbCollection, error) {
	dbC := &DbCollection{Name: name}
	err := dbC.Save(dbC)
	if err != nil {
		return nil, err
	}
	dbC.DbAccessPolicy, err = DbAccessPolicyNew(dbC.ID, adminUsers, adminGroups, readerUsers, readerGroups)
	if err != nil {
		return nil, err
	}
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

func (d *DbCollection) AppendDbColRevision() error {
	_, err := DbColRevisionNew(d.ID)
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
