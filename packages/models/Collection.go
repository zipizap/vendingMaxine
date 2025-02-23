package models

import (
	"fmt"
	"vendingMaxine/packages/models/dbModels"
)

type Collection struct {
	dbCollectionIfc dbModels.DbCollectionIfc // unexported field, only used by Collection package and not other packages
}

// Constructor creates dbCollectionIfc and public-methods use dbCollectionIfc to access r/w data

func NewCollection(name string) (*Collection, error) {
	// create DbCollection into dbCollectionIfc and return Collection
	c := &Collection{}
	var err error
	c.dbCollectionIfc, err = dbModels.NewDbCollection(name)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Ex: col, err := LoadCollection("ColID-1234")
func LoadCollection(colId string) (*Collection, error) {
	// load DbCollection from dbCollectionIfc and return Collection
	dbColId, err := colID_2_dbColID(colId)
	if err != nil {
		return nil, err
	}
	c := &Collection{}
	c.dbCollectionIfc, err = dbModels.LoadDbCollection(dbColId)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Ex: "ColID-1234"
func (c *Collection) GetID() (colIdString string, err error) {
	colIdUint := c.dbCollectionIfc.GetID()
	if colIdUint == 0 {
		err = fmt.Errorf("colIdUint is 0, ?maybe collection does not exist in db?")
		return colIdString, err
	}
	colIdString = dbColID_2_colID(colIdUint)
	return colIdString, err
}

func (c *Collection) GetName() (string, error) {
	return c.dbCollectionIfc.GetName()
}

func (c *Collection) Rename(newName string) error {
	return c.dbCollectionIfc.SetName(newName)
}
