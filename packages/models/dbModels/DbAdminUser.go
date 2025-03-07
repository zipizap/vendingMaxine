package dbModels

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
)

// DbAdminUserIfc defines the interface for DbAdminUser operations
type DbAdminUserIfc interface {
	GetID() uint
	GetOid() (string, error)
	GetDbAccessPolicies() ([]*DbAccessPolicy, error)
	GetDbAccessPoliciesIDs() ([]uint, error)
}

// DbAdminUser represents an admin user for access policies
type DbAdminUser struct {
	gormCrud.GormCrud[DbAdminUser]
	Oid              string            `gorm:"uniqueIndex"`                            // User identifier with unique constraint
	DbAccessPolicies []*DbAccessPolicy `gorm:"many2many:admin_users__db_admin_users;"` // Many-to-many relationship
}

// DbAdminUserNew creates a new DbAdminUser
// unique_oid is a unique oid in this table
func DbAdminUserNew(unique_oid string) (*DbAdminUser, error) {
	dbAAU := &DbAdminUser{
		Oid: unique_oid,
	}
	err := dbAAU.Save(dbAAU)
	if err != nil {
		return nil, err
	}
	return dbAAU, nil
}

// DbAdminUserLoad loads a DbAdminUser by ID
// It returns an error if the gorm_ID is not found
func DbAdminUserLoad(gorm_ID uint) (*DbAdminUser, error) {
	dbAAU := &DbAdminUser{}
	results, err := dbAAU.LoadWhere("id = ?", gorm_ID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbAdminUser with id %d not found", gorm_ID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbAAU = results[0]
	return dbAAU, nil
}

// DbAdminUserLoadByOid loads a DbAdminUser by oid, returns an error if not found
func DbAdminUserLoadByOid(oid string) (*DbAdminUser, error) {
	dbAAU := &DbAdminUser{}
	results, err := dbAAU.LoadWhere("oid = ?", oid)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbAdminUser with oid %s not found", oid)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbAAU = results[0]
	return dbAAU, nil
}

// DbAdminUserLoadOrCreateNew loads a DbAdminUser by oid or creates a new one
func DbAdminUserLoadOrCreateNew(oid string) (*DbAdminUser, error) {
	dbAAU := &DbAdminUser{}
	results, err := dbAAU.LoadWhere("oid = ?", oid)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return DbAdminUserNew(oid)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbAAU = results[0]
	return dbAAU, nil
}

// GetID returns the ID of the DbAdminUser
func (d *DbAdminUser) GetID() uint {
	return d.ID
}

// GetOid returns the Oid of the DbAdminUser
func (d *DbAdminUser) GetOid() (string, error) {
	return d.Oid, nil
}

// GetDbAccessPoliciesIDs returns the IDs of all associated DbAccessPolicies
func (d *DbAdminUser) GetDbAccessPoliciesIDs() ([]uint, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}

	var ids []uint
	// Query the join table directly to get just the policy IDs
	err := gormCrud.Db.Table("admin_users__db_access_admin_users").
		Where("db_access_admin_user_id = ?", d.ID).
		Pluck("db_access_policy_id", &ids).Error

	if err != nil {
		return nil, err
	}

	return ids, nil
}

// GetDbAccessPolicies returns all associated DbAccessPolicies
func (d *DbAdminUser) GetDbAccessPolicies() ([]*DbAccessPolicy, error) {
	// Get policy IDs first
	policyIDs, err := d.GetDbAccessPoliciesIDs()
	if err != nil {
		return nil, err
	}

	if len(policyIDs) == 0 {
		return []*DbAccessPolicy{}, nil
	}

	// Load all policies with the retrieved IDs
	var policies []*DbAccessPolicy
	err = gormCrud.Db.Where("id IN ?", policyIDs).Find(&policies).Error
	if err != nil {
		return nil, err
	}

	return policies, nil
}
