package gormCrud

import (
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TestModel is a sample struct to test CRUD operations.
type TestModel struct {
	GormCrud[TestModel] // Embed GormCrud
	Name                string
	Description         string
}

// TestModelA represents the first model in a one-to-one relationship.
type TestModelA struct {
	GormCrud[TestModelA]
	Name         string
	TestModelB   *TestModelB   // 1TestModelA-to-1TestModelB relationship
	TestModelC   []*TestModelC // 1TestModelA-to-manyTestModelC relationship
	TestModelD   *TestModelD   // manyTestModelA-to-1TestModelD - 2 fields: TestModelD and TestModelDID (mandatory!)
	TestModelDID uint
}

// TestModelB represents the second model in a one-to-one relationship.
type TestModelB struct {
	GormCrud[TestModelB]
	TestModelAID uint `gorm:"unique"` // 1TestModelA-to-1TestModelB relationship (mandatory!)
	Description  string
}

// TestModelC represents the third model in a one-to-many relationship with TestModelA.
type TestModelC struct {
	GormCrud[TestModelC]
	TestModelAID uint // 1TestModelA-to-manyTestModelC relationship
	Description  string
}

// TestModelD represents the fourth model in a many-to-one relationship with TestModelA.
// No special fields needed inside TestModelD
type TestModelD struct {
	GormCrud[TestModelD]
	Name        string
	Description string
}

// Add User and AutoBus types for testing many-to-many relationships
type AutoBus struct {
	GormCrud[AutoBus]
	BusName string
	// manyUser-to-manyAutoBuses (internally uses a join table user_autobuses)
	Users []*User `gorm:"many2many:user_autobuses;"`
}

type User struct {
	GormCrud[User]
	Name string
	// manyUser-to-manyAutoBuses (internally uses a join table user_autobuses)
	AutoBuses []*AutoBus `gorm:"many2many:user_autobuses;"`
}

// setupTest initializes a new in-memory database, migrates the TestModel, and resets package states.
func setupTest(t *testing.T, sqliteDsn string) *gorm.DB {
	// Close existing connection (if any)
	var err error
	if Db != nil {
		err = close()
	}
	assert.NoError(t, err)

	// Reset package states
	Db = nil
	migratedTypes = make(map[reflect.Type]bool)
	dbInitialized = false

	// sqliteDsn defaults to an in-memory database, unique for each test, to make disposable and fast dbs
	// However to test concurrent operations, dont use in-memory databases and prefer instead a file-based database - to avoid thread-safety issues ;)
	if sqliteDsn == "" {
		sqliteDsn = "file:" + t.Name() + "?mode=memory&cache=shared"
	}

	// Initialize DB with the provided filename
	err = InitializeDB(sqliteDsn)
	assert.NoError(t, err)

	// Migrate the TestModel
	err = MigrateModels(&TestModel{}, &TestModelA{}, &TestModelB{}, &TestModelC{}, &TestModelD{}, &User{}, &AutoBus{})
	assert.NoError(t, err)

	return Db
}

// TestSaveAndLoad tests saving and loading a record.
func TestSaveAndLoad(t *testing.T) {
	setupTest(t, "")

	// Create and save a test instance
	instance := &TestModel{Name: "Test", Description: "Test Description"}
	err := instance.Save(instance)
	assert.NoError(t, err)
	assert.NotZero(t, instance.ID, "ID should be auto-incremented")

	// Load the saved instance
	results, err := instance.LoadWhere("name = ?", "Test")
	assert.NoError(t, err)
	assert.Len(t, results, 1, "One record should be loaded")
	assert.Equal(t, "Test Description", results[0].Description)
}

// TestDelete tests deleting a record.
func TestDelete(t *testing.T) {
	setupTest(t, "")

	// Create and save a test instance
	instance := &TestModel{Name: "ToDelete", Description: "Delete me"}
	err := instance.Save(instance)
	assert.NoError(t, err)

	// Delete the instance
	err = instance.Delete(instance)
	assert.NoError(t, err)

	// Verify deletion
	results, err := instance.LoadWhere("id = ?", instance.ID)
	assert.NoError(t, err)
	assert.Empty(t, results, "No records should remain after deletion")
}

// TestLoadWhereWithConditions tests loading records with specific conditions.
func TestLoadWhereWithConditions(t *testing.T) {
	setupTest(t, "")

	// Create multiple test instances
	instance1 := &TestModel{Name: "Alice", Description: "First"}
	err := instance1.Save(instance1)
	assert.NoError(t, err)

	instance2 := &TestModel{Name: "Bob", Description: "Second"}
	err = instance2.Save(instance2)
	assert.NoError(t, err)

	// Load all records
	results, err := instance1.LoadWhere("1 = 1") // Load all
	assert.NoError(t, err)
	assert.Len(t, results, 2, "All records should be loaded")

	// Load records with a condition
	results, err = instance1.LoadWhere("name = ?", "Bob")
	assert.NoError(t, err)
	assert.Len(t, results, 1, "Only one record should match the condition")
	assert.Equal(t, "Second", results[0].Description)
}

// TestConcurrentOperations tests concurrent saves to ensure thread safety.
func TestConcurrentOperations(t *testing.T) {
	// sqlite in-memory database is not thread-safe, so we specify a file-based database
	sqliteDsn := "/tmp/" + t.Name() + ".sqlite"
	// Remove the SQLite database file if it already exists
	_ = os.Remove(sqliteDsn)
	setupTest(t, sqliteDsn)

	var wg sync.WaitGroup
	saveFunc := func() {
		defer wg.Done()
		instance := &TestModel{Name: "Concurrent", Description: "Test"}
		err := instance.Save(instance)
		assert.NoError(t, err)
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go saveFunc()
	}
	wg.Wait()

	// Verify all records are saved
	var count int64
	err := Db.Model(&TestModel{}).Where("name = ?", "Concurrent").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(10), count, "All concurrent saves should succeed")
}

func TestModelGormFields(t *testing.T) {
	setupTest(t, "")

	instance := &TestModel{Name: "GormFieldsCheck", Description: "Check gorm.Model fields"}
	err := instance.Save(instance)
	assert.NoError(t, err)
	assert.NotZero(t, instance.ID, "Inherited ID should be auto-incremented")
	assert.False(t, instance.CreatedAt.IsZero(), "CreatedAt should be automatically set")
	assert.False(t, instance.UpdatedAt.IsZero(), "UpdatedAt should be automatically set")

	results, err := instance.LoadWhere("id = ?", instance.ID)
	assert.NoError(t, err)
	assert.Len(t, results, 1, "Should find the saved record")
	assert.Equal(t, instance.ID, results[0].ID)
	assert.False(t, results[0].CreatedAt.IsZero(), "Loaded CreatedAt should not be zero")
	assert.False(t, results[0].UpdatedAt.IsZero(), "Loaded UpdatedAt should not be zero")
}

// TestCRUDModelA tests CRUD operations on TestModelA.
func TestCRUDModelA(t *testing.T) {
	setupTest(t, "")

	// Create and save a TestModelA instance first
	modelA := &TestModelA{Name: "ModelA"}
	err := modelA.Save(modelA)
	assert.NoError(t, err)

	// Create and save a TestModelB instance with reference to TestModelA
	modelB := &TestModelB{Description: "Related ModelB", TestModelAID: modelA.ID}
	err = modelB.Save(modelB)
	assert.NoError(t, err)

	// Reload TestModelA to verify association
	results, err := modelA.LoadWhere("name = ?", "ModelA")
	assert.NoError(t, err)
	assert.Len(t, results, 1, "One TestModelA record should be loaded")
	assert.NotNil(t, results[0].TestModelB)
	assert.Equal(t, "Related ModelB", results[0].TestModelB.Description)
}

// TestReadModelBFromModelA tests reading TestModelB from TestModelA.
func TestReadModelBFromModelA(t *testing.T) {
	setupTest(t, "")

	// Create and save a TestModelA instance first
	modelA := &TestModelA{Name: "ModelA2"}
	err := modelA.Save(modelA)
	assert.NoError(t, err)

	// Create and save a TestModelB instance with reference to TestModelA
	modelB := &TestModelB{Description: "Another ModelB", TestModelAID: modelA.ID}
	err = modelB.Save(modelB)
	assert.NoError(t, err)

	// Retrieve TestModelA and access TestModelB using clause.Associations
	var retrievedModelA TestModelA
	err = Db.Preload(clause.Associations).First(&retrievedModelA, modelA.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, retrievedModelA.TestModelB)
	assert.Equal(t, "Another ModelB", retrievedModelA.TestModelB.Description)
}

func TestReload(t *testing.T) {
	setupTest(t, "")

	// Create and save a test instance
	instance := &TestModel{Name: "Original", Description: "Initial Description"}
	err := instance.Save(instance)
	assert.NoError(t, err)

	// Modify the instance directly in the DB to simulate external change
	err = Db.Model(&TestModel{}).Where("id = ?", instance.ID).Update("description", "Updated Description").Error
	assert.NoError(t, err)

	// Reload the instance
	err = instance.Reload(instance)
	assert.NoError(t, err)

	// Assert the field has been updated
	assert.Equal(t, "Updated Description", instance.Description, "Reload should update the description field")
}

// TestCRUDModelC tests CRUD operations for TestModelC in a one-to-many relationship with TestModelA.
func TestCRUDModelCWithModelA(t *testing.T) {
	setupTest(t, "")

	// Create and save a TestModelA instance
	modelA := &TestModelA{Name: "ModelA_ForC"}
	err := modelA.Save(modelA)
	assert.NoError(t, err)

	// Create and save multiple TestModelC instances related to TestModelA
	modelC1 := &TestModelC{TestModelAID: modelA.ID, Description: "Detail 1"}
	modelC2 := &TestModelC{TestModelAID: modelA.ID, Description: "Detail 2"}

	err = modelC1.Save(modelC1)
	assert.NoError(t, err)
	err = modelC2.Save(modelC2)
	assert.NoError(t, err)

	// Read: Load related TestModelC instances
	results, err := modelA.LoadWhere("id = ?", modelA.ID)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Len(t, results[0].TestModelC, 2)

	// Update: Modify a TestModelC instance
	modelC1.Description = "Updated Detail 1"
	err = modelC1.Save(modelC1)
	assert.NoError(t, err)

	// Verify update
	updated, err := modelC1.LoadWhere("id = ?", modelC1.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Detail 1", updated[0].Description)

	// Delete: Remove a TestModelC instance
	err = modelC2.Delete(modelC2)
	assert.NoError(t, err)

	// Verify deletion
	deleted, err := modelC2.LoadWhere("id = ?", modelC2.ID)
	assert.NoError(t, err)
	assert.Empty(t, deleted, "TestModelC instance should be deleted")
}

// TestCRUDModelAWithModelC tests CRUD operations creating TestModelA with associated TestModelC instances.
func TestCRUDModelAWithModelC(t *testing.T) {
	setupTest(t, "")

	// Create and save a TestModelA instance
	modelA := &TestModelA{
		Name: "ModelA_WithC",
		TestModelC: []*TestModelC{
			{Description: "C1 Description"},
			{Description: "C2 Description"},
		},
	}
	err := modelA.Save(modelA)
	assert.NoError(t, err)

	// read readModelA and then c1 and c2 from it
	res, err := (&TestModelA{}).LoadWhere("name = ?", "ModelA_WithC")
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	readModelA := res[0]
	assert.Len(t, readModelA.TestModelC, 2)
	c1 := readModelA.TestModelC[0]
	c2 := readModelA.TestModelC[1]
	assert.Equal(t, "C1 Description", c1.Description)
	assert.Equal(t, "C2 Description", c2.Description)

	// Update c2 description
	c2.Description = "Updated C2 Description"

	err = c1.Save(c1)
	assert.NoError(t, err)
	err = c2.Save(c2)
	assert.NoError(t, err)

	// Read: Retrieve TestModelA with associated TestModelC instances
	var retrievedA TestModelA
	results, err := modelA.LoadWhere("id = ?", modelA.ID)
	retrievedA = *results[0]
	assert.NoError(t, err)
	assert.Len(t, retrievedA.TestModelC, 2)

	// Update: Update TestModelA's Name
	modelA.Name = "Updated ModelA_WithC"
	err = modelA.Save(modelA)
	assert.NoError(t, err)

	// Verify update
	updatedA, err := modelA.LoadWhere("id = ?", modelA.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated ModelA_WithC", updatedA[0].Name)

	// Delete: Remove one TestModelC instance
	err = c1.Delete(c1)
	assert.NoError(t, err)

	// Verify deletion
	remainingCs, err := c1.LoadWhere("test_model_a_id = ?", modelA.ID)
	assert.NoError(t, err)
	assert.Len(t, remainingCs, 1)
	assert.Equal(t, "Updated C2 Description", remainingCs[0].Description)
}

// TestAppendTestModelC tests appending a new TestModelC to TestModelA and verifies persistence.
func TestAppendTestModelC(t *testing.T) {
	setupTest(t, "")

	// Create and save a TestModelA instance
	modelA := &TestModelA{Name: "ModelA_AppendC"}
	err := modelA.Save(modelA)
	assert.NoError(t, err)

	// Append a new TestModelC instance
	newC := &TestModelC{Description: "Appended Detail"}
	modelA.TestModelC = append(modelA.TestModelC, newC)

	// Save TestModelA with the new TestModelC
	err = modelA.Save(modelA)
	assert.NoError(t, err)

	// Reload TestModelA from the database
	reloadedA, err := (&TestModelA{}).LoadWhere("id = ?", modelA.ID)
	assert.NoError(t, err)
	assert.Len(t, reloadedA, 1, "Should find the TestModelA record")

	// Assert the new TestModelC is present
	assert.Len(t, reloadedA[0].TestModelC, 1, "TestModelA should have one TestModelC")
	assert.Equal(t, "Appended Detail", reloadedA[0].TestModelC[0].Description)
}

// TestCRUDModelD tests CRUD operations for TestModelD
func TestCRUDModelD(t *testing.T) {
	setupTest(t, "")

	// Create and save TestModelD instance
	d := &TestModelD{Name: "ModelD1", Description: "Description for ModelD1"}
	err := d.Save(d)
	assert.NoError(t, err)

	// Create and save TestModelA with reference to TestModelD
	a := &TestModelA{
		Name:       "ModelA1",
		TestModelD: d,
	}
	err = a.Save(a)
	assert.NoError(t, err)

	// Load TestModelA and verify association with TestModelD
	results, err := a.LoadWhere("name = ?", "ModelA1")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.NotNil(t, results[0].TestModelD)
	assert.Equal(t, "Description for ModelD1", results[0].TestModelD.Description)

	// Update TestModelD
	d.Description = "Updated Description for ModelD1"
	err = d.Save(d)
	assert.NoError(t, err)

	// Verify update
	updatedD, err := d.LoadWhere("id = ?", d.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Description for ModelD1", updatedD[0].Description)

	// Delete TestModelA and verify
	err = a.Delete(a)
	assert.NoError(t, err)
	deletedA, err := a.LoadWhere("id = ?", a.ID)
	assert.NoError(t, err)
	assert.Empty(t, deletedA, "TestModelA should be deleted")
}

// TestCRUDUserAutoBus tests CRUD operations for many-to-many relation between User and AutoBus.
func TestCRUDUserAutoBus(t *testing.T) {
	setupTest(t, "")

	// Create and save AutoBus instances
	bus1 := &AutoBus{BusName: "Express"}
	bus2 := &AutoBus{BusName: "Local"}
	err := bus1.Save(bus1)
	assert.NoError(t, err)
	err = bus2.Save(bus2)
	assert.NoError(t, err)

	// Create and save User with associated AutoBuses
	user := &User{
		Name:      "Charlie",
		AutoBuses: []*AutoBus{bus1, bus2},
	}
	err = user.Save(user)
	assert.NoError(t, err)

	// Load User and verify associations
	loadedUsers, err := user.LoadWhere("name = ?", "Charlie")
	assert.NoError(t, err)
	assert.Len(t, loadedUsers, 1, "One User should be loaded")
	loadedUser := loadedUsers[0]
	assert.Len(t, loadedUser.AutoBuses, 2, "User should have two associated AutoBuses")
	assert.Equal(t, "Express", loadedUser.AutoBuses[0].BusName)
	assert.Equal(t, "Local", loadedUser.AutoBuses[1].BusName)

	// Update association: Append another AutoBus
	loadedUser.AutoBuses = append(loadedUser.AutoBuses, &AutoBus{BusName: "Shuttle"})
	err = loadedUser.Save(loadedUser)
	assert.NoError(t, err)

	// Verify update
	updatedUsers, err := (&User{}).LoadWhere("name = ?", "Charlie")
	assert.NoError(t, err)
	assert.Len(t, updatedUsers, 1, "One User should be loaded after update")
	assert.Len(t, updatedUsers[0].AutoBuses, 3, "User should have three associated AutoBus after update")
	assert.Equal(t, "Express", updatedUsers[0].AutoBuses[0].BusName)

	// Delete User and verify associations
	err = user.Delete(user)
	assert.NoError(t, err)

	// Verify deletion
	deletedUsers, err := user.LoadWhere("name = ?", "Charlie")
	assert.NoError(t, err)
	assert.Empty(t, deletedUsers, "User should be deleted")
}
