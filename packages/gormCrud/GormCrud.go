package gormCrud

import (
	"fmt"
	"reflect"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	Db            *gorm.DB
	migratedTypes = make(map[reflect.Type]bool)
	migrateMutex  sync.Mutex
	dbInitialized bool
)

// GormCrud provides generic CRUD operations for the model type T.
type GormCrud[T any] struct {
	gorm.Model
}

// InitializeDB initializes the database connection with the provided SQLite filename.
// This should be called once at the start of the application or test.
func InitializeDB(sqliteFilename string) error {
	if dbInitialized {
		return nil
	}
	var err error
	Db, err = gorm.Open(sqlite.Open(sqliteFilename), &gorm.Config{})
	if err != nil {
		return err
	}
	dbInitialized = true
	return nil
}

// close closes the database connection.
// intented to be used only for tests
func close() error {
	sqlDB, err := Db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// MigrateModels performs AutoMigrate for the provided models.
// This should be called at program startup for each model.
func MigrateModels(models ...interface{}) error {
	if !dbInitialized {
		return fmt.Errorf("database not initialized. Call InitializeDB first")
	}

	migrateMutex.Lock()
	defer migrateMutex.Unlock()
	for _, model := range models {
		t := reflect.TypeOf(model)
		if !migratedTypes[t] {
			if err := Db.AutoMigrate(model); err != nil {
				return err
			}
			migratedTypes[t] = true
		}
	}
	return nil
}

// Save persists or updates an instance of T into the database.
func (g *GormCrud[T]) Save(instance *T) error {
	if Db == nil {
		return fmt.Errorf("database not initialized")
	}
	return Db.Save(instance).Error
}

// LoadWhere retrieves matching records of T based on the given query, from the database.
func (g *GormCrud[T]) LoadWhere(query string, args ...interface{}) ([]*T, error) {
	if Db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	var results []*T
	err := Db.Where(query, args...).Preload(clause.Associations).Find(&results).Error
	return results, err
}

// Reload reloads the given instance from the database using its ID field.
func (g *GormCrud[T]) Reload(instance *T) error {
	if Db == nil {
		return fmt.Errorf("database not initialized")
	}
	return Db.Preload(clause.Associations).First(instance).Error
}

// Delete removes the specified instance of T from the database.
func (g *GormCrud[T]) Delete(instance *T) error {
	if Db == nil {
		return fmt.Errorf("database not initialized")
	}
	return Db.Delete(instance).Error
}
