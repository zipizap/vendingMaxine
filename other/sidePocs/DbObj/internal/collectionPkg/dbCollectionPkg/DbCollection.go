package dbCollectionPkg

import (
	"DbObj/internal/dbPkg"
	"fmt"

	"gorm.io/gorm"
)

func init() {
	dbPkg.AddModelForAutoMigration(&DbCollection{})
}

type DbCollectionIfc interface {
	GetName() string
	// CollectionRevisions() []DbCollectionRevision
}

type DbCollection struct {
	Name string
	// DbCollectionRevisions []DbCollectionRevision
}

func CreateDbCollection(name string) (*DbCollection, error) {
	// Check if there is already an existing record with the same name in the db
	existingCollection, err := loadDbCollectionByName(name)
	if err != nil {
		return nil, err
	}
	if existingCollection != nil {
		return nil, fmt.Errorf("a collection with the name '%s' already exists", name)
	}

	// Save the new record into the db
	newCollection := &DbCollection{Name: name}
	// err = saveDbCollection(newCollection)
	// if err != nil {
	// 	return nil, err
	// }

	return newCollection, nil
}

// Dummy functions to represent database operations
func loadDbCollectionByName(name string) (*DbCollection, error) {

	var dbCol DbCollection
	result := dbPkg.Db.Where("name = ?", name).First(&dbCol)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}

	return &dbCol, nil
}

// func saveDbCollection(dbCol *DbCollection) error {
// 	// TODO: Implement actual database save logic
// 	return nil
// }

func (c *DbCollection) reload() {
	// TODO: reload record from db
	// return
}

// ------------ Interface methods ------------
func (c *DbCollection) GetName() string {
	c.reload()
	return c.Name
}

/*
func (c *DbCollection) CollectionRevisions() []DbCollectionRevision {
    return c.DbCollectionRevisions
}
*/
