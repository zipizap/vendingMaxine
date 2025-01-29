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
	GetName() (string, error)
	SetName(name string) error
	// CollectionRevisions() []DbCollectionRevision
}

type DbCollection struct {
	gorm.Model
	Name string
	// DbCollectionRevisions []DbCollectionRevision
}

// Create a new DbCollection in the database
func CreateDbCollection(name string) (*DbCollection, error) {
	// Check if there is already an existing record with the same name in the db
	existingDbCollection, err := LoadDbCollectionByName(name)
	if err != nil {
		return nil, err
	}
	if existingDbCollection != nil {
		return nil, fmt.Errorf("a collection with the name '%s' already exists", name)
	}

	// Save the new record into the db
	newDbCollection := &DbCollection{Name: name}
	err = newDbCollection.save()
	if err != nil {
		return nil, err
	}

	return newDbCollection, nil
}

// Load a DbCollection record from the database by name
func LoadDbCollectionByName(name string) (*DbCollection, error) {

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

/*
	TODO
		- finish this func ()
		- do the same for reload() and ñloadWhere()
*/

// Save a DbCollection record into the database.
// If the record exists, it will be updated. If it doesn't exist, it will be created.
func save(theModelObjToSave interface{}, aDummyModelObj interface{}) error {

	var count int64
	// Check if the record exists based on the ID
	gormModel := theModelObjToSave.(gorm.Model)
	id := gormModel.ID
	result := dbPkg.Db.Model(aDummyModelObj).Where("id = ?", id).Count(&count)
	if result.Error != nil {
		return fmt.Errorf("database error: %v", result.Error)
	}

	if count == 0 {
		// Create new record
		if err := dbPkg.Db.Create(theModelObjToSave).Error; err != nil {
			return fmt.Errorf("failed to create record: %v", err)
		}
	} else {
		// Update existing record
		if err := dbPkg.Db.Save(theModelObjToSave).Error; err != nil {
			return fmt.Errorf("failed to update record: %v", err)
		}
	}
	return nil
}

func (dbCol *DbCollection) save() error {
	return save(dbCol, &DbCollection{})
}

func (dbCol *DbCollection) reload() error {
	result := dbPkg.Db.First(dbCol, dbCol.ID)
	if result.Error != nil {
		return fmt.Errorf("database error: %v", result.Error)
	}
	return nil
}

// ------------ Interface methods ------------
func (dbCol *DbCollection) GetName() (name string, err error) {
	if err = dbCol.reload(); err != nil {
		name = ""
		return name, err
	}
	name = dbCol.Name
	return name, nil
}

func (dbCol *DbCollection) SetName(name string) error {
	dbCol.Name = name
	if err := dbCol.save(); err != nil {
		return err
	}
	return nil
}

/*
func (c *DbCollection) CollectionRevisions() []DbCollectionRevision {
    return c.DbCollectionRevisions
}
*/
