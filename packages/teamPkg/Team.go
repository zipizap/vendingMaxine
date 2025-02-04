package teamPkg

import (
	"fmt"
	"strconv"
	"strings"
	"vendingMaxine/packages/teamPkg/dbTeamPkg"
	"vendingMaxine/packages/teamPkg/userGroupPkg"
)

// Empty types, to be implemented later.

type Team struct {
	dbTeamIfc dbTeamPkg.DbTeamIfc
}

func NewTeam(name string) (*Team, error) {
	dbTeam, err := dbTeamPkg.NewDbTeam(name)
	if err != nil {
		return nil, err
	}
	return &Team{dbTeamIfc: dbTeam}, nil
}

func LoadTeam(teamID string) (*Team, error) {
	// helper: extract uint id from "TeamID-<id>"
	parts := strings.Split(teamID, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid teamID format")
	}
	idUint, err := strconv.ParseUint(parts[1], 10, 0)
	if err != nil {
		return nil, err
	}

	dbTeam, err := dbTeamPkg.LoadDbTeam(uint(idUint))
	if err != nil {
		return nil, err
	}
	return &Team{dbTeamIfc: dbTeam}, nil
}

func (t *Team) GetID() (string, error) {
	id := t.dbTeamIfc.GetID()
	if id == 0 {
		return "", fmt.Errorf("invalid team id")
	}
	return fmt.Sprintf("TeamID-%d", id), nil
}

func (t *Team) GetName() (string, error) {
	return t.dbTeamIfc.GetName()
}

func (t *Team) Rename(newName string) error {
	return t.dbTeamIfc.SetName(newName)
}

func (t *Team) GetUsers() ([]userGroupPkg.User, error) {
	return t.dbTeamIfc.GetUsers()
}

func (t *Team) SetUsers(users []userGroupPkg.User) error {
	return t.dbTeamIfc.SetUsers(users)
}

func (t *Team) GetGroups() ([]userGroupPkg.Group, error) {
	return t.dbTeamIfc.GetGroups()
}

func (t *Team) SetGroups(groups []userGroupPkg.Group) error {
	return t.dbTeamIfc.SetGroups(groups)
}

func (t *Team) IsMember(user userGroupPkg.User, group userGroupPkg.Group) (bool, error) {
	return t.dbTeamIfc.IsMember(user, group)
}
