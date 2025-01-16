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

func (c *Collection) Name() string {
	return c.dbCollectionIfc.GetName()
}
