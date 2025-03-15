package dbModels

import (
	"fmt"
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/sharedTypes"
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
	GetRole(user string, groups []string) (role string, err error)
	IsAdmin(user string, groups []string) (bool, error)
	IsReader(user string, groups []string) (bool, error)
	GetDbCollectionID() (uint, error)
}

type DbAccessPolicy struct {
	gormCrud.GormCrud[DbAccessPolicy]
	DbCollectionID         uint                     `gorm:"unique"` // 1DbAccessPolicy-to-1DbCollection
	DbAccessPolicyMappings []*DbAccessPolicyMapping // 1DbAccessPolicy-to-manyDbAccessPolicyMappings
}

func DbAccessPolicyNew(dbCollectionID uint, accessPolicyParams sharedTypes.AccessPolicyParams) (*DbAccessPolicy, error) {
	// Create a new DbAccessPolicy
	dbAP := &DbAccessPolicy{
		DbCollectionID: dbCollectionID,
	}

	// Save the DbAccessPolicy to get an ID
	err := dbAP.Save(dbAP)
	if err != nil {
		return nil, err
	}

	// Create mappings for admin users
	for _, user := range accessPolicyParams.AdminUsers {
		_, err := DbAccessPolicyMappingNew(dbAP.ID, dbCollectionID, "admin", user, "")
		if err != nil {
			return nil, err
		}
	}

	// Create mappings for admin groups
	for _, group := range accessPolicyParams.AdminGroups {
		_, err := DbAccessPolicyMappingNew(dbAP.ID, dbCollectionID, "admin", "", group)
		if err != nil {
			return nil, err
		}
	}

	// Create mappings for reader users
	for _, user := range accessPolicyParams.ReaderUsers {
		_, err := DbAccessPolicyMappingNew(dbAP.ID, dbCollectionID, "reader", user, "")
		if err != nil {
			return nil, err
		}
	}

	// Create mappings for reader groups
	for _, group := range accessPolicyParams.ReaderGroups {
		_, err := DbAccessPolicyMappingNew(dbAP.ID, dbCollectionID, "reader", "", group)
		if err != nil {
			return nil, err
		}
	}
	return dbAP, nil
}

// DbAccessPolicyLoad loads a DbAccessPolicy from the database by its ID
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

// GetAdminUsers returns a list of users with admin role for this policy
func (d *DbAccessPolicy) GetAdminUsers() ([]string, error) {
	// Create a temporary DbAccessPolicyMapping to query
	dbAPM := &DbAccessPolicyMapping{}

	// Query all mappings with admin role and non-empty user field for this policy
	results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user != ?", d.ID, "admin", "")
	if err != nil {
		return nil, err
	}

	// Extract user names from the results
	adminUsers := make([]string, 0, len(results))
	for _, mapping := range results {
		adminUsers = append(adminUsers, mapping.User)
	}

	return adminUsers, nil
}

func (d *DbAccessPolicy) GetAdminGroups() ([]string, error) {
	// Create a temporary DbAccessPolicyMapping to query
	dbAPM := &DbAccessPolicyMapping{}

	// Query all mappings with admin role and non-empty group field for this policy
	results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group != ?", d.ID, "admin", "")
	if err != nil {
		return nil, err
	}

	// Extract group names from the results
	adminGroups := make([]string, 0, len(results))
	for _, mapping := range results {
		adminGroups = append(adminGroups, mapping.Group)
	}

	return adminGroups, nil
}

func (d *DbAccessPolicy) GetReaderUsers() ([]string, error) {
	// Create a temporary DbAccessPolicyMapping to query
	dbAPM := &DbAccessPolicyMapping{}

	// Query all mappings with reader role and non-empty user field for this policy
	results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user != ?", d.ID, "reader", "")
	if err != nil {
		return nil, err
	}

	// Extract user names from the results
	readerUsers := make([]string, 0, len(results))
	for _, mapping := range results {
		readerUsers = append(readerUsers, mapping.User)
	}

	return readerUsers, nil
}

func (d *DbAccessPolicy) GetReaderGroups() ([]string, error) {
	// Create a temporary DbAccessPolicyMapping to query
	dbAPM := &DbAccessPolicyMapping{}

	// Query all mappings with reader role and non-empty group field for this policy
	results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group != ?", d.ID, "reader", "")
	if err != nil {
		return nil, err
	}

	// Extract group names from the results
	readerGroups := make([]string, 0, len(results))
	for _, mapping := range results {
		readerGroups = append(readerGroups, mapping.Group)
	}

	return readerGroups, nil
}

func (d *DbAccessPolicy) SetAdminUsers(newAdminUsers []string) error {
	// Get current admin users
	oldAdminUsers, err := d.GetAdminUsers()
	if err != nil {
		return err
	}

	// Find users to delete (those in old list but not in new list)
	oldAdminUsersToDelete := []string{}
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

	// Find users to add (those in new list but not in old list)
	newAdminUsersToAdd := []string{}
	for _, newAdminUser := range newAdminUsers {
		found := false
		for _, oldAdminUser := range oldAdminUsers {
			if newAdminUser == oldAdminUser {
				found = true
				break
			}
		}
		if !found {
			newAdminUsersToAdd = append(newAdminUsersToAdd, newAdminUser)
		}
	}

	// Delete users that are no longer in the list
	dbAPM := &DbAccessPolicyMapping{}
	for _, oldAdminUserToDelete := range oldAdminUsersToDelete {
		// Find the mapping for this user
		results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user = ?", d.ID, "admin", oldAdminUserToDelete)
		if err != nil {
			return err
		}

		// Delete each mapping found
		for _, mapping := range results {
			if err := mapping.Delete(mapping); err != nil {
				return err
			}
		}
	}

	// Add new users
	dbColID, err := d.GetDbCollectionID()
	if err != nil {
		return err
	}

	for _, userToAdd := range newAdminUsersToAdd {
		// Check if this user already has this role in this collection
		existingMapping, err := DbAccessPolicyMappingExists(d.ID, dbColID, "admin", userToAdd, "")
		if err != nil {
			return err
		}

		// Skip if mapping already exists
		if existingMapping != nil {
			continue
		}

		// Create new mapping
		_, err = DbAccessPolicyMappingNew(d.ID, dbColID, "admin", userToAdd, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DbAccessPolicy) SetAdminGroups(groups []string) error {
	// Get current admin groups
	oldAdminGroups, err := d.GetAdminGroups()
	if err != nil {
		return err
	}

	// Find groups to delete
	groupsToDelete := []string{}
	for _, oldGroup := range oldAdminGroups {
		found := false
		for _, newGroup := range groups {
			if oldGroup == newGroup {
				found = true
				break
			}
		}
		if !found {
			groupsToDelete = append(groupsToDelete, oldGroup)
		}
	}

	// Find groups to add
	groupsToAdd := []string{}
	for _, newGroup := range groups {
		found := false
		for _, oldGroup := range oldAdminGroups {
			if newGroup == oldGroup {
				found = true
				break
			}
		}
		if !found {
			groupsToAdd = append(groupsToAdd, newGroup)
		}
	}

	// Delete groups that are no longer in the list
	dbAPM := &DbAccessPolicyMapping{}
	for _, groupToDelete := range groupsToDelete {
		results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group = ?", d.ID, "admin", groupToDelete)
		if err != nil {
			return err
		}

		for _, mapping := range results {
			if err := mapping.Delete(mapping); err != nil {
				return err
			}
		}
	}

	// Add new groups
	dbColID, err := d.GetDbCollectionID()
	if err != nil {
		return err
	}
	for _, groupToAdd := range groupsToAdd {
		existingMapping, err := DbAccessPolicyMappingExists(d.ID, dbColID, "admin", "", groupToAdd)
		if err != nil {
			return err
		}

		if existingMapping != nil {
			continue
		}

		_, err = DbAccessPolicyMappingNew(d.ID, dbColID, "admin", "", groupToAdd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DbAccessPolicy) SetReaderUsers(users []string) error {
	// Get current reader users
	oldReaderUsers, err := d.GetReaderUsers()
	if err != nil {
		return err
	}

	// Find users to delete
	usersToDelete := []string{}
	for _, oldUser := range oldReaderUsers {
		found := false
		for _, newUser := range users {
			if oldUser == newUser {
				found = true
				break
			}
		}
		if !found {
			usersToDelete = append(usersToDelete, oldUser)
		}
	}

	// Find users to add
	usersToAdd := []string{}
	for _, newUser := range users {
		found := false
		for _, oldUser := range oldReaderUsers {
			if newUser == oldUser {
				found = true
				break
			}
		}
		if !found {
			usersToAdd = append(usersToAdd, newUser)
		}
	}

	// Delete users that are no longer in the list
	dbAPM := &DbAccessPolicyMapping{}
	for _, userToDelete := range usersToDelete {
		results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user = ?", d.ID, "reader", userToDelete)
		if err != nil {
			return err
		}

		for _, mapping := range results {
			if err := mapping.Delete(mapping); err != nil {
				return err
			}
		}
	}

	// Add new users
	dbColID, err := d.GetDbCollectionID()
	if err != nil {
		return err
	}

	for _, userToAdd := range usersToAdd {
		existingMapping, err := DbAccessPolicyMappingExists(d.ID, dbColID, "reader", userToAdd, "")
		if err != nil {
			return err
		}

		if existingMapping != nil {
			continue
		}

		_, err = DbAccessPolicyMappingNew(d.ID, dbColID, "reader", userToAdd, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DbAccessPolicy) SetReaderGroups(groups []string) error {
	// Get current reader groups
	oldReaderGroups, err := d.GetReaderGroups()
	if err != nil {
		return err
	}

	// Find groups to delete
	groupsToDelete := []string{}
	for _, oldGroup := range oldReaderGroups {
		found := false
		for _, newGroup := range groups {
			if oldGroup == newGroup {
				found = true
				break
			}
		}
		if !found {
			groupsToDelete = append(groupsToDelete, oldGroup)
		}
	}

	// Find groups to add
	groupsToAdd := []string{}
	for _, newGroup := range groups {
		found := false
		for _, oldGroup := range oldReaderGroups {
			if newGroup == oldGroup {
				found = true
				break
			}
		}
		if !found {
			groupsToAdd = append(groupsToAdd, newGroup)
		}
	}

	// Delete groups that are no longer in the list
	dbAPM := &DbAccessPolicyMapping{}
	for _, groupToDelete := range groupsToDelete {
		results, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group = ?", d.ID, "reader", groupToDelete)
		if err != nil {
			return err
		}

		for _, mapping := range results {
			if err := mapping.Delete(mapping); err != nil {
				return err
			}
		}
	}

	// Add new groups
	dbColID, err := d.GetDbCollectionID()
	if err != nil {
		return err
	}

	for _, groupToAdd := range groupsToAdd {
		existingMapping, err := DbAccessPolicyMappingExists(d.ID, dbColID, "reader", "", groupToAdd)
		if err != nil {
			return err
		}

		if existingMapping != nil {
			continue
		}

		_, err = DbAccessPolicyMappingNew(d.ID, dbColID, "reader", "", groupToAdd)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetRole returns the role of a user for this policy
// role is either "admin", "reader", or ""
func (d *DbAccessPolicy) GetRole(user string, groups []string) (role string, err error) {
	var isAdmin bool
	isAdmin, err = d.IsAdmin(user, groups)
	if err != nil {
		return "", err
	}
	if isAdmin {
		return "admin", nil
	}

	isReader, err := d.IsReader(user, groups)
	if err != nil {
		return "", err
	}
	if isReader {
		return "reader", nil
	}
	return "", nil
}

// IsAdmin checks if a user is an admin for this policy
func (d *DbAccessPolicy) IsAdmin(user string, groups []string) (bool, error) {
	// Check if user directly has admin role
	dbAPM := &DbAccessPolicyMapping{}
	{
		userResults, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user = ?", d.ID, "admin", user)
		if err != nil {
			return false, err
		}
		if len(userResults) > 0 {
			return true, nil
		}
	}

	// Get all admin groups for this policy in a single query
	adminGroupMappings, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group != ?", d.ID, "admin", "")
	if err != nil {
		return false, err
	}

	// Convert to a map for faster lookups
	adminGroupsMap := make(map[string]bool)
	for _, mapping := range adminGroupMappings {
		adminGroupsMap[mapping.Group] = true
	}

	// Check if any of user's groups are in the admin groups map
	for _, group := range groups {
		if adminGroupsMap[group] {
			return true, nil
		}
	}

	return false, nil
}

// IsReader checks if a user is a reader for this policy
func (d *DbAccessPolicy) IsReader(user string, groups []string) (bool, error) {
	// Check if user is an admin (admins can read)
	{
		isAdmin, err := d.IsAdmin(user, groups)
		if err != nil {
			return false, err
		}

		if isAdmin {
			return true, nil
		}
	}

	// Check if user directly has reader role
	dbAPM := &DbAccessPolicyMapping{}
	{
		userResults, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND user = ?", d.ID, "reader", user)
		if err != nil {
			return false, err
		}

		if len(userResults) > 0 {
			return true, nil
		}
	}

	// Get all reader groups for this policy in a single query
	readerGroupMappings, err := dbAPM.LoadWhere("db_access_policy_id = ? AND role = ? AND group != ?", d.ID, "reader", "")
	if err != nil {
		return false, err
	}

	// Convert to a map for faster lookups
	readerGroupsMap := make(map[string]bool)
	for _, mapping := range readerGroupMappings {
		readerGroupsMap[mapping.Group] = true
	}

	// Check if any of user's groups are in the reader groups map
	for _, group := range groups {
		if readerGroupsMap[group] {
			return true, nil
		}
	}

	return false, nil
}

func (d *DbAccessPolicy) GetDbCollectionID() (dbColID uint, err error) {
	dbColID = d.DbCollectionID
	return dbColID, nil
}
