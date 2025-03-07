package dbModels

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
)

type DbAccessPolicyIfc interface {
	GetID() uint
	GetAdminUsers() ([]string, error)
	SetAdminUsers([]string) error
	GetAdminGroups() ([]string, error)
	SetAdminGroups([]string) error
	GetReaderUsers() ([]string, error)
	SetReaderUsers([]string) error
	GetReaderGroups() ([]string, error)
	SetReaderGroups([]string) error
	IsAdmin(user string, groups []string) (bool, error)
	IsReader(user string, groups []string) (bool, error)
	GetDbCollectionID() (uint, error)
}

type DbAccessPolicyParams struct {
	AdminUsers   []string
	AdminGroups  []string
	ReaderUsers  []string
	ReaderGroups []string
}

type DbAccessPolicy struct {
	gormCrud.GormCrud[DbAccessPolicy]
	DbCollectionID uint           `gorm:"unique"`                                 // 1DbAccessPolicy-to-1DbCollection
	AdminUsers     []*DbAdminUser `gorm:"many2many:admin_users__db_admin_users;"` // []string
	AdminGroups    bool           // []*DbAdminGroup        `gorm:"many2many:admin_groups__db_admin_groups;"`   // []string
	ReaderUsers    bool           // []*DbReaderUser        `gorm:"many2many:reader_users__db_reader_users;"`   // []string
	ReaderGroups   bool           // []*DbAccessReaderGroup `gorm:"many2many:reader_groups__db_reader_groups;"` // []string
}

func DbAccessPolicyNew(dbCollectionID uint, adminUsers, adminGroups, readerUsers, readerGroups []string) (*DbAccessPolicy, error) {
	// Create dbAP and Save() to get an ID
	var dbAP *DbAccessPolicy
	{
		dbAP = &DbAccessPolicy{
			DbCollectionID: dbCollectionID,
		}
		err := dbAP.Save(dbAP)
		if err != nil {
			return nil, err
		}
	}

	// convert adminUsers []string to dbAdminUsers []*DbAdminUser
	var dbAdminUsers []*DbAdminUser
	{
		for _, adminUser := range adminUsers {
			dbAAU, err := DbAdminUserLoadOrCreateNew(adminUser)
			if err != nil {
				return nil, err
			}
			dbAdminUsers = append(dbAdminUsers, dbAAU)
		}
	}
	dbAP.AdminUsers = dbAdminUsers

	// TODO: AdminGroups, ReaderUsers, ReaderGroups

	return dbAP, nil
}

func DbAccessPolicyLoad(dbAccessPolicyID uint) (*DbAccessPolicy, error) {
	dbAP := &DbAccessPolicy{}
	results, err := dbAP.LoadWhere("id = ?", dbAccessPolicyID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbAccessPolicy with id %d not found", dbAccessPolicyID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbAP = results[0]
	return dbAP, nil
}

func (d *DbAccessPolicy) GetID() uint {
	return d.ID
}

func (d *DbAccessPolicy) GetAdminUsers() ([]string, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	dbAdminUsers := d.AdminUsers
	var adminUsers []string
	for _, dbAdminUser := range dbAdminUsers {
		oid, err := dbAdminUser.GetOid()
		if err != nil {
			return nil, err
		}
		adminUsers = append(adminUsers, oid)
	}
	return adminUsers, nil
}

func (d *DbAccessPolicy) SetAdminUsers(users []string) error {
	newAdminUsers := users
	oldAdminUsers, err := d.GetAdminUsers()
	if err != nil {
		return err
	}

	// Calculate oldAdminUsersToDelete
	oldAdminUsersToDelete := []string{}
	{
		for _, oldAdminUser := range oldAdminUsers {
			found := false
			for _, newAdminUser := range newAdminUsers {
				if oldAdminUser == newAdminUser {
					found = true
					break
				}
			}
			if !found {
				oldAdminUsersToDelete = append(oldAdminUsersToDelete, oldAdminUser)
			}
		}
	}

	// Delete oldAdminUsersToDelete
	{
		for _, oldAdminUserToDelete := range oldAdminUsersToDelete {
			oldDbAdminUserToDelete, err := DbAdminUserLoadByOid(oldAdminUserToDelete)
			if err != nil {
				return err
			}
			err = d.deleteElementFromFieldWithMany2ManyAssociation("AdminUsers", oldDbAdminUserToDelete)
			if err != nil {
				return err
			}
		}
	}

	// Create newDbAdminUsers and save to d
	var newDbAdminUsers []*DbAdminUser
	{
		for _, newAdminUser := range newAdminUsers {
			newDbAdminUser, err := DbAdminUserLoadOrCreateNew(newAdminUser)
			if err != nil {
				return err
			}
			newDbAdminUsers = append(newDbAdminUsers, newDbAdminUser)
		}

		// Save newDbAdminUsers to d
		d.AdminUsers = newDbAdminUsers
		err = d.Save(d)
		if err != nil {
			return err
		}

	}

	return nil
}

func (d *DbAccessPolicy) GetAdminGroups() ([]string, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}

	// TODO
	return []string{}, nil
}

func (d *DbAccessPolicy) SetAdminGroups(groups []string) error {
	// TODO
	return nil

	// if err := d.Reload(d); err != nil {
	// 	return err
	// }
	// d.AdminGroups = groups
	// return d.Save(d)
}

func (d *DbAccessPolicy) GetReaderUsers() ([]string, error) {
	// TODO
	return []string{}, nil

	// if err := d.Reload(d); err != nil {
	// 	return nil, err
	// }
	// return d.ReaderUsers, nil
}

func (d *DbAccessPolicy) SetReaderUsers(users []string) error {
	// TODO
	return nil

	// if err := d.Reload(d); err != nil {
	// 	return err
	// }
	// d.ReaderUsers = users
	// return d.Save(d)
}

func (d *DbAccessPolicy) GetReaderGroups() ([]string, error) {
	// TODO
	return []string{}, nil

	// if err := d.Reload(d); err != nil {
	// 	return nil, err
	// }
	// return d.ReaderGroups, nil
}

func (d *DbAccessPolicy) SetReaderGroups(groups []string) error {
	// TODO
	return nil

	// if err := d.Reload(d); err != nil {
	// 	return err
	// }
	// d.ReaderGroups = groups
	// return d.Save(d)
}

func (d *DbAccessPolicy) IsAdmin(user string, groups []string) (bool, error) {
	if err := d.Reload(d); err != nil {
		return false, err
	}

	// Make a db-query to check if user is in AdminUsers
	var count int64
	err := gormCrud.Db.Table("admin_users__db_admin_users").
		Joins("JOIN db_admin_users ON admin_users__db_admin_users.db_admin_user_id = db_admin_users.id").
		Where("admin_users__db_admin_users.db_access_policy_id = ? AND db_admin_users.oid = ?", d.ID, user).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	// TODO: Check if any of user's groups are in AdminGroups
	// Currently AdminGroups is a boolean field and not fully implemented

	return false, nil
}

func (d *DbAccessPolicy) IsReader(user string, groups []string) (bool, error) {
	// TODO
	return false, nil

	// if err := d.Reload(d); err != nil {
	// 	return false, err
	// }
	// for _, readerUser := range d.ReaderUsers {
	// 	if readerUser == user {
	// 		return true, nil
	// 	}
	// }
	// for _, readerGroup := range d.ReaderGroups {
	// 	for _, group := range groups {
	// 		if readerGroup == group {
	// 			return true, nil
	// 		}
	// 	}
	// }
	// return false, nil
}

func (d *DbAccessPolicy) GetDbCollectionID() (dbColID uint, err error) {
	if err = d.Reload(d); err != nil {
		return 0, err
	}
	dbColID = d.DbCollectionID
	return dbColID, nil
}

func (d *DbAccessPolicy) deleteElementFromFieldWithMany2ManyAssociation(fieldName string, elementToDelete interface{}) error {
	err := gormCrud.Db.Model(d).Association(fieldName).Delete(elementToDelete)
	if err != nil {
		return err
	}
	return d.Reload(d)
}
