package dbTeamPkg

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/teamPkg/userGroupPkg"
)

type DbTeamIfc interface {
	GetID() uint
	GetName() (string, error)
	SetName(string) error
	GetUsers() ([]userGroupPkg.User, error)
	SetUsers([]userGroupPkg.User) error
	GetGroups() ([]userGroupPkg.Group, error)
	SetGroups([]userGroupPkg.Group) error
	IsMember(userGroupPkg.User, userGroupPkg.Group) (bool, error)
}

type DbTeam struct {
	gormCrud.GormCrud[DbTeam]
	Name   string
	Users  []userGroupPkg.User
	Groups []userGroupPkg.Group
}

func NewDbTeam(name string) (*DbTeam, error) {
	dbT := &DbTeam{Name: name}
	if err := dbT.Save(dbT); err != nil {
		return nil, err
	}
	return dbT, nil
}

func LoadDbTeam(teamID uint) (*DbTeam, error) {
	dbT := &DbTeam{}
	results, err := dbT.LoadWhere("id = ?", teamID)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("DbTeam with id %d not found", teamID)
	} else if len(results) > 1 {
		return nil, fmt.Errorf("expected 1 result, got %d", len(results))
	}
	dbT = results[0]
	return dbT, nil
}

func (d *DbTeam) GetID() uint {
	return d.ID
}

func (d *DbTeam) GetName() (string, error) {
	// Name might change, so we always reload it from the database
	if err := d.Reload(d); err != nil {
		return "", err
	}
	return d.Name, nil
}

func (d *DbTeam) SetName(newName string) error {
	// Name might change, so we always reload it from the database
	if err := d.Reload(d); err != nil {
		return err
	}
	d.Name = newName
	return d.Save(d)
}

func (d *DbTeam) GetUsers() ([]userGroupPkg.User, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	return d.Users, nil
}

func (d *DbTeam) SetUsers(users []userGroupPkg.User) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.Users = users
	return d.Save(d)
}

func (d *DbTeam) GetGroups() ([]userGroupPkg.Group, error) {
	if err := d.Reload(d); err != nil {
		return nil, err
	}
	return d.Groups, nil
}

func (d *DbTeam) SetGroups(groups []userGroupPkg.Group) error {
	if err := d.Reload(d); err != nil {
		return err
	}
	d.Groups = groups
	return d.Save(d)
}

func (d *DbTeam) IsMember(user userGroupPkg.User, group userGroupPkg.Group) (bool, error) {
	// Reload latest data
	if err := d.Reload(d); err != nil {
		return false, err
	}

	userFound := false
	groupFound := false
	for _, u := range d.Users {
		// Assuming userGroupPkg.User supports equality (may update with proper uniqueID comparison)
		if u == user {
			userFound = true
			break
		}
	}

	for _, g := range d.Groups {
		if g == group {
			groupFound = true
			break
		}
	}

	userOrGroupFound := userFound || groupFound
	return userOrGroupFound, nil
}
