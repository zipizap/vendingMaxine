package models

import (
	"fmt"
	"vendingMaxine/packages/models/dbModels"
	"vendingMaxine/packages/sharedTypes"
)

type AccessPolicy struct {
	dbIfc dbModels.DbAccessPolicyIfc
}

func AccessPolicyNew(collectionID string, accessPolicyParams sharedTypes.AccessPolicyParams) (*AccessPolicy, error) {
	collectionIDuint, err := Collection_convert_ID_2_IDuint(collectionID)
	if err != nil {
		return nil, err
	}

	ap := &AccessPolicy{}
	ap.dbIfc, err = dbModels.DbAccessPolicyNew(
		collectionIDuint,
		accessPolicyParams)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func AccessPolicyLoad(accessPolicyID string) (*AccessPolicy, error) {
	accessPolicyIDuint, err := convert_IDstring_2_IDuint(accessPolicyID)
	if err != nil {
		return nil, err
	}
	ap := &AccessPolicy{}
	dbAccessPolicyID := accessPolicyIDuint
	ap.dbIfc, err = dbModels.DbAccessPolicyLoad(dbAccessPolicyID)
	if err != nil {
		return nil, err
	}
	return ap, nil
}

func AccessPolicy_convert_ID_2_IDuint(ID string) (IDuint uint, err error) {
	return convert_IDstring_2_IDuint(ID)
}
func AccessPolicy_convert_IDuint_2_ID(IDuint uint) (ID string) {
	return convert_IDuint_2_IDstring("AccessPolicyID", IDuint)
}

// Ex: "AccessPolicyID-1234"
func (o *AccessPolicy) GetID() (ID string, err error) {
	IDuint := o.dbIfc.GetID()
	if IDuint == 0 {
		return "", fmt.Errorf("IDuint is 0, ?maybe accessPolicy does not exist in db?")
	}
	ID = AccessPolicy_convert_IDuint_2_ID(IDuint)
	return ID, nil
}

func (ap *AccessPolicy) GetAdmins() (adminUsers []string, adminGroups []string, err error) {
	adminUsers, err = ap.dbIfc.GetAdminUsers()
	if err != nil {
		return nil, nil, err
	}
	adminGroups, err = ap.dbIfc.GetAdminGroups()
	if err != nil {
		return nil, nil, err
	}
	return adminUsers, adminGroups, nil
}

func (ap *AccessPolicy) SetAdmins(adminUsers []string, adminGroups []string) (err error) {
	err = ap.dbIfc.SetAdminUsers(adminUsers)
	if err != nil {
		return err
	}
	err = ap.dbIfc.SetAdminGroups(adminGroups)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AccessPolicy) GetReaders() (readerUsers []string, readerGroups []string, err error) {
	readerUsers, err = ap.dbIfc.GetReaderUsers()
	if err != nil {
		return nil, nil, err
	}
	readerGroups, err = ap.dbIfc.GetReaderGroups()
	if err != nil {
		return nil, nil, err
	}
	return readerUsers, readerGroups, nil
}

func (ap *AccessPolicy) SetReaders(readerUsers []string, readerGroups []string) (err error) {
	err = ap.dbIfc.SetReaderUsers(readerUsers)
	if err != nil {
		return err
	}
	err = ap.dbIfc.SetReaderGroups(readerGroups)
	if err != nil {
		return err
	}
	return nil
}

func (ap *AccessPolicy) IsAdmin(user string, groups []string) (bool, error) {
	return ap.dbIfc.IsAdmin(user, groups)
}

func (ap *AccessPolicy) IsReader(user string, groups []string) (bool, error) {
	return ap.dbIfc.IsReader(user, groups)
}

func (ap *AccessPolicy) GetCollection() (*Collection, error) {
	collectionIDuint, err := ap.dbIfc.GetDbCollectionID()
	if err != nil {
		return nil, err
	}
	collectionID := Collection_convert_IDuint_2_ID(collectionIDuint)
	return CollectionLoad(collectionID)
}
