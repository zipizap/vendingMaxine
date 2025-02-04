package dbCollectionPkg

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
)

// DbCollection follows gorm conventions, takes care of db-struct and db-methods

type DbCollectionIfc interface {
	GetID() uint
	GetName() (string, error)
	SetName(string) error
}

type DbCollection struct {
	gormCrud.GormCrud[DbCollection]
	Name string
}

func NewDbCollection(name string) (*DbCollection, error) {
	dbC := &DbCollection{Name: name}
	err := dbC.Save(dbC)
	if err != nil {
		return nil, err
	}
	return dbC, nil
}

func LoadDbCollection(dbColId uint) (*DbCollection, error) {
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
	// Name might change, so we reload it from the database
	{
		err := d.Reload(d)
		if err != nil {
			return "", err
		}
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
