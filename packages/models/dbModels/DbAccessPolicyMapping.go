package dbModels

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
)

type DbAccessPolicyMapping struct {
	gormCrud.GormCrud[DbAccessPolicyMapping]
	DbAccessPolicyID uint // 1DbAccessPolicy-to-manyDbAccessPolicyMapping
	DbCollectionID   uint
	Role             string
	User             string
	Group            string
}

// DbAccessPolicyMappingExists checks if a DbAccessPolicyMapping with the same parameters already exists
// Returns either:
//   - error if there is an error
//   - nil if no DbAccessPolicyMapping with the same parameters exists
//   - the DbAccessPolicyMapping if it exists
func DbAccessPolicyMappingExists(dbAccessPolicyID uint, dbCollectionID uint, role, user, group string) (*DbAccessPolicyMapping, error) {
	dbAPM := &DbAccessPolicyMapping{}
	results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND db_collection_id = ? AND role = ? AND user = ? AND group = ?", dbAccessPolicyID, dbCollectionID, role, user, group)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil // No error, but no result found
	}
	if len(results) > 1 {
		return nil, fmt.Errorf("multiple DbAccessPolicyMappings found with the same parameters")
	}
	return results[0], nil
}

func DbAccessPolicyMappingNew(dbAccessPolicyID uint, dbCollectionID uint, role, user, group string) (*DbAccessPolicyMapping, error) {
	// Create dbAPM and Save()
	var dbAPM *DbAccessPolicyMapping

	// Assure that an DbAccessPolicyMapping with same parameters does not exist yet
	// If it already exists, return an error
	{
		dbAPM, err := DbAccessPolicyMappingExists(dbAccessPolicyID, dbCollectionID, role, user, group)
		if err != nil {
			return nil, err
		}
		if dbAPM != nil {
			// DbAccessPolicyMapping with same parameters already exists - return an error
			return nil, fmt.Errorf("DbAccessPolicyMapping with same parameters already exists")
		}
	}

	// Create a new DbAccessPolicyMapping and .Save() it
	dbAPM = &DbAccessPolicyMapping{
		DbAccessPolicyID: dbAccessPolicyID,
		DbCollectionID:   dbCollectionID,
		Role:             role,
		User:             user,
		Group:            group,
	}
	err := dbAPM.Save(dbAPM)
	if err != nil {
		return nil, err
	}

	return dbAPM, nil
}

func DbAccessPolicyMappingLoad(DbAccessPolicyMappingID uint) (*DbAccessPolicyMapping, error) {
	dbAPM := &DbAccessPolicyMapping{}
	results, err := dbAPM.LoadWhere("id = ?", DbAccessPolicyMappingID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbAccessPolicyMapping with id %d not found", DbAccessPolicyMappingID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbAPM = results[0]

	return dbAPM, nil
}
