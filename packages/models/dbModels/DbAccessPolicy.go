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

type DbAccessPolicy struct {
	gormCrud.GormCrud[DbAccessPolicy]
	DbCollectionID uint `gorm:"unique"` // 1DbAccessPolicy-to-1DbCollection
	AdminUsers     []string
	AdminGroups    []string
	ReaderUsers    []string
	ReaderGroups   []string
}

func DbAccessPolicyNew(dbCollectionID uint, adminUsers, adminGroups, readerUsers, readerGroups []string) (*DbAccessPolicy, error) {
	dbAP := &DbAccessPolicy{
		DbCollectionID: dbCollectionID,
		AdminUsers:     adminUsers,
		AdminGroups:    adminGroups,
		ReaderUsers:    readerUsers,
		ReaderGroups:   readerGroups,
	}
	err := dbAP.Save(dbAP)
	if err != nil {
		return nil, err
	}
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
	return d.AdminUsers, nil
}

func (d *DbAccessPolicy) SetAdminUsers(users []string) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.AdminUsers = users
	return d.Save(d)
}

func (d *DbAccessPolicy) GetAdminGroups() ([]string, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	return d.AdminGroups, nil
}

func (d *DbAccessPolicy) SetAdminGroups(groups []string) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.AdminGroups = groups
	return d.Save(d)
}

func (d *DbAccessPolicy) GetReaderUsers() ([]string, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	return d.ReaderUsers, nil
}

func (d *DbAccessPolicy) SetReaderUsers(users []string) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.ReaderUsers = users
	return d.Save(d)
}

func (d *DbAccessPolicy) GetReaderGroups() ([]string, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	return d.ReaderGroups, nil
}

func (d *DbAccessPolicy) SetReaderGroups(groups []string) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.ReaderGroups = groups
	return d.Save(d)
}

func (d *DbAccessPolicy) IsAdmin(user string, groups []string) (bool, error) {
	if err := d.Reload(d); err != nil {
		return false, err
	}
	for _, adminUser := range d.AdminUsers {
		if adminUser == user {
			return true, nil
		}
	}
	for _, adminGroup := range d.AdminGroups {
		for _, group := range groups {
			if adminGroup == group {
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *DbAccessPolicy) IsReader(user string, groups []string) (bool, error) {
	if err := d.Reload(d); err != nil {
		return false, err
	}
	for _, readerUser := range d.ReaderUsers {
		if readerUser == user {
			return true, nil
		}
	}
	for _, readerGroup := range d.ReaderGroups {
		for _, group := range groups {
			if readerGroup == group {
				return true, nil
			}
		}
	}
	return false, nil
}

func (d *DbAccessPolicy) GetDbCollectionID() (dbColID uint, err error) {
	if err = d.Reload(d); err != nil {
		return 0, err
	}
	dbColID = d.DbCollectionID
	return dbColID, nil
}
