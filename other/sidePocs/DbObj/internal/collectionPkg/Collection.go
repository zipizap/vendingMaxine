package collectionPkg

import (
	"DbObj/internal/collectionPkg/dbCollectionPkg"
)

type Collection struct {
	dbCollectionIfc dbCollectionPkg.DbCollectionIfc
}

func CreateCollection(name string) (*Collection, error) {
	dbCollection, err := dbCollectionPkg.CreateDbCollection(name)
	if err != nil {
		return nil, err
	}
	return &Collection{dbCollectionIfc: dbCollection}, nil
}

func LoadCollectionByName(name string) (*Collection, error) {
	dbCollection, err := dbCollectionPkg.LoadDbCollectionByName(name)
	if err != nil {
		return nil, err
	}
	return &Collection{dbCollectionIfc: dbCollection}, nil
}

func (c *Collection) GetName() (name string, err error) {
	name, err = c.dbCollectionIfc.GetName()
	return name, err
}

func (c *Collection) SetName(name string) (err error) {
	err = c.dbCollectionIfc.SetName(name)
	return err
}
