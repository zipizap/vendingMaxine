/*
Package gormCrud provides a generic CRUD layer for SQLite databases using GORM.

# Usage:

 1. Embed `GormCrud[T]` in your model structs to inherit CRUD operations.
 2. Call `InitializeDB(sqliteFilename string)` to set up the database connection
 3. Call `MigrateModels(&MyModel1{}, &MyModel2{}, ...)`.
 4. Use the CRUD methods on your models: myModel1.Save(), myModel2.LoadWhere(), myModel2.Delete()

The GormCrud[T] will:

  - Automatically create a table for the model, if it doesn't exist.
  - Provide methods for CRUD operations: Save(), LoadWhere(), Reload(), Delete()
  - Embed gorm.Model fields: ID, CreatedAt, UpdatedAt, DeletedAt

Example:

	type TypeA struct {
		gormCrud.GormCrud[TypeA]
		Name string
	}

	func main() {
		 // Initialize the database with a specific filename
		if err := gormCrud.InitializeDB("mydb.db"); err != nil {
			panic(err)
		}

		// Migrate models
		if err := gormCrud.MigrateModels(&TypeA{}); err != nil {
			panic(err)
		}

		instance := &TypeA{Name: "test"}
		err := instance.Save(instance) // Create/Update
		results, err := instance.LoadWhere("name = ?", "test") // Read
		err = instance.Reload(instance)
		err = instance.Delete(instance)
	}

Methods:

  - InitializeDB(sqliteFilename string) error
    Initializes the database connection. Must be called once before any CRUD operations.

  - MigrateModels(models ...interface{}) error
    Performs AutoMigrate for the provided models. Must be called at program startup for each model.

  - .Save(instance *T) error
    Creates or updates a record in the database.

  - .LoadWhere(query string, args ...interface{}) ([]*T, error)
    Retrieves records matching the given conditions.

  - .Reload(instance *T) error
    Reloads the instance from the database using its ID, updating all fields.

  - .Delete(instance *T) error
    Deletes the record from the database.

Database Setup:
  - Uses SQLite saving to disk-file.
  - Call `InitializeDB("mydb.db")` to set the database file.
  - Call `MigrateModels(...yourModels)` to perform AutoMigration (via GORM AutoMigrate)
  - Thread-safe initialization and migration
  - For testing, uses isolated in-memory databases

Dependencies:
  - github.com/stretchr/testify/assert (for tests)
  - gorm.io/gorm
  - gorm.io/driver/sqlite

Install:

	go get -u gorm.io/gorm
	go get -u gorm.io/driver/sqlite
*/
package gormCrud
