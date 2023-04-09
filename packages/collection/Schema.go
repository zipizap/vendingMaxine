package collection

import (
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Schema struct {
	gorm.Model
	dbMethods
	VersionName string `gorm:"unique,uniqueIndex,not null"`
	Json        string `gorm:"not null"`
}

// SchemaNew method should create a new object o and
//
//   - set the new object fields from its corresponding arguments
//   - verify if o.Json is a valid json and if not then return error
//   - verify the sql unique constraints, or return an error
//     Should check all possible errors
//     If inside this method, there is any error at any step, then return the error
//     If this function is executed without errors, then:
//   - call o.Save(o) and return the create object o
func SchemaNew(versionName string, jsonStr string) (*Schema, error) {
	var o Schema
	o.VersionName = versionName
	o.Json = jsonStr

	// Verify if o.Json is a valid json
	if err := json.Unmarshal([]byte(o.Json), &map[string]interface{}{}); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	// Verify the sql unique constraints
	if err := db.Where("version_name = ?", o.VersionName).First(&Schema{}).Error; err == nil {
		return nil, errors.New("version name already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("error checking unique constraint: %w", err)
	}

	err := o.save(&o)
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func SchemaLoadLatest() (*Schema, error) {
	o := &Schema{}
	err := db.Last(o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		initialSchema, err1 := schemaCreateInitial()
		if err1 != nil {
			return nil, err1
		}
		return initialSchema, nil
	} else if err != nil {
		return nil, err
	}
	return o, nil
}

func schemaCreateInitial() (*Schema, error) {
	initialVersionName := "initial empty schema"
	initialJson := ""
	initialSchema, err := SchemaNew(initialVersionName, initialJson)
	if err != nil {
		return nil, err
	}
	return initialSchema, nil
}

func (o *Schema) gormID() uint {
	return o.ID
}
