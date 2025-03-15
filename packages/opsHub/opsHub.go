package opshub

import (
	"fmt"
	"time"
	"vendingMaxine/packages/models"
	"vendingMaxine/packages/sharedTypes"
)

type CurrentClient struct {
	User   string
	Groups []string
}

type CollectionsListReq struct {
	Client CurrentClient
}

type CollectionsListResp struct {
	Collections []struct {
		CollectionID                   string
		CollectionName                 string
		CollectionDesc                 string
		RevStateLatestName             string
		RevStateLatestUserWhoTriggered string
		CollectionModDate              time.Time
		IsEditableForCurrentClient     bool
	}
}

type CollectionNewReq struct {
	Client             CurrentClient
	CollectionName     string
	CollectionDesc     string
	AccessPolicyParams sharedTypes.AccessPolicyParams
}

type CollectionNewResp struct {
	CollectionID string
}

type CollectionEditOngoingReq struct {
	Client            CurrentClient
	CollectionID      string
	ColRevDescription string
}

type CollectionEditOngoingResp struct {
}

type CollectionEditCancelledReq struct {
	Client       CurrentClient
	CollectionID string
}

type CollectionEditCancelledResp struct {
}

type CollectionReadyReq struct {
	Client       CurrentClient
	CollectionID string
}

type CollectionReadyResp struct {
}

type CollectionEditCompletedReq struct {
	Client       CurrentClient
	CollectionID string
}

type CollectionEditCompletedResp struct {
}

type CollectionProvisioningOngoingReq struct {
	Client       CurrentClient
	CollectionID string
}

type CollectionProvisioningOngoingResp struct {
}

// CollectionsList returns the list of collections for the user and groups
// It returns two lists: readerCols and adminCols
// readerCols are the collections that the user can read
// adminCols are the collections that the user can edit
// If there are no collections, it returns empty lists
func CollectionsList(req *CollectionsListReq) (resp *CollectionsListResp, err error) {
	// Initialize response
	resp = &CollectionsListResp{
		Collections: []struct {
			CollectionID                   string
			CollectionName                 string
			CollectionDesc                 string
			RevStateLatestName             string
			RevStateLatestUserWhoTriggered string
			CollectionModDate              time.Time
			IsEditableForCurrentClient     bool
		}{},
	}

	// Get collections using the models package
	readerCols, adminCols, err := models.CollectionsList(req.Client.User, req.Client.Groups)
	if err != nil {
		return nil, err
	}

	// Helper function to create a collection item for the response
	createCollectionItem := func(col *models.Collection, isEditable bool) (struct {
		CollectionID                   string
		CollectionName                 string
		CollectionDesc                 string
		RevStateLatestName             string
		RevStateLatestUserWhoTriggered string
		CollectionModDate              time.Time
		IsEditableForCurrentClient     bool
	}, error) {
		var item struct {
			CollectionID                   string
			CollectionName                 string
			CollectionDesc                 string
			RevStateLatestName             string
			RevStateLatestUserWhoTriggered string
			CollectionModDate              time.Time
			IsEditableForCurrentClient     bool
		}

		// Get collection ID
		id, err := col.GetID()
		if err != nil {
			return item, err
		}
		item.CollectionID = id

		// Get collection name
		name, err := col.GetName()
		if err != nil {
			return item, err
		}
		item.CollectionName = name

		// Get collection description
		desc, err := col.GetDescription()
		if err != nil {
			return item, err
		}
		item.CollectionDesc = desc

		// Get latest revision state name
		revStateName, err := col.GetRevStateLatestName()
		if err != nil {
			return item, err
		}
		item.RevStateLatestName = revStateName

		// Get user who triggered the latest revision state
		userWhoTriggered, err := col.GetRevStateLatestUserWhoTriggered()
		if err != nil {
			return item, err
		}
		item.RevStateLatestUserWhoTriggered = userWhoTriggered

		// Get modification date
		modDate, err := col.GetModDate()
		if err != nil {
			return item, err
		}
		item.CollectionModDate = modDate

		// Set editability based on parameter and collection's own state
		actuallyEditable := false
		if isEditable {
			actuallyEditable, err = col.IsEditable()
			if err != nil {
				return item, err
			}
		}
		item.IsEditableForCurrentClient = actuallyEditable

		return item, nil
	}

	// Process reader collections (read-only)
	for _, col := range readerCols {
		collectionItem, err := createCollectionItem(col, false)
		if err != nil {
			return nil, err
		}
		resp.Collections = append(resp.Collections, collectionItem)
	}

	// Process admin collections (editable)
	for _, col := range adminCols {
		collectionItem, err := createCollectionItem(col, true)
		if err != nil {
			return nil, err
		}
		resp.Collections = append(resp.Collections, collectionItem)
	}

	return resp, nil
}

// CollectionNew creates a new collection, and returns the collection ID
func CollectionNew(req *CollectionNewReq) (resp *CollectionNewResp, err error) {
	// Initialize response
	resp = &CollectionNewResp{}

	stringInSlice := func(a string, slice []string) bool {
		for _, b := range slice {
			if b == a {
				return true
			}
		}
		return false
	}

	// Assure that req.Client.User is always included in req.AccessPolicyParams.AdminUsers, but not duplicating a possible existing entry
	// This is to ensure that the user who creates the collection is always an admin
	if !stringInSlice(req.Client.User, req.AccessPolicyParams.AdminUsers) {
		req.AccessPolicyParams.AdminUsers = append(req.AccessPolicyParams.AdminUsers, req.Client.User)
	}

	// Create a new collection using the models package
	col, err := models.CollectionNew(req.CollectionName, req.CollectionDesc, req.AccessPolicyParams, req.Client.User)
	if err != nil {
		return nil, err
	}

	// Get collection ID
	id, err := col.GetID()
	if err != nil {
		return nil, err
	}
	resp.CollectionID = id

	return resp, nil
}

// validateUserIsAdmin checks if the user has admin rights for the collection
// Returns error if the user is not an admin
func validateUserIsAdmin(col *models.Collection, user string, groups []string) error {
	role, err := col.GetRole(user, groups)
	if err != nil {
		return err
	}
	if role != "admin" {
		colID, _ := col.GetID()
		return fmt.Errorf("in collection '%s' the user as role '%s', but this operation requires role 'admin' ", colID, role)
	}
	return nil
}

// DoCollectionEditOngoing marks the start of a CollectionEdit operation
func DoCollectionEditOngoing(req *CollectionEditOngoingReq) (resp *CollectionEditOngoingResp, err error) {
	// Initialize response
	resp = &CollectionEditOngoingResp{}

	// Get the collection using the models package
	col, err := models.CollectionLoad(req.CollectionID)
	if err != nil {
		return nil, err
	}

	// Validate user is admin
	err = validateUserIsAdmin(col, req.Client.User, req.Client.Groups)
	if err != nil {
		return nil, err
	}

	// Set the revision state description using the models package
	err = col.DoCollectionEditOngoing(req.ColRevDescription, req.Client.User)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func DoCollectionEditCancelled(req *CollectionEditCancelledReq) (resp *CollectionEditCancelledResp, err error) {
	// Initialize response
	resp = &CollectionEditCancelledResp{}

	// Get the collection using the models package
	col, err := models.CollectionLoad(req.CollectionID)
	if err != nil {
		return nil, err
	}

	// Validate that Client can edit the collection
	err = validateUserIsAdmin(col, req.Client.User, req.Client.Groups)
	if err != nil {
		return nil, err
	}

	// Set the revision state description using the models package
	err = col.DoCollectionEditCancelled(req.Client.User)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func DoCollectionEditCompleted(req *CollectionEditCompletedReq) (resp *CollectionEditCompletedResp, err error) {
	// Initialize response
	resp = &CollectionEditCompletedResp{}

	// Get the collection using the models package
	col, err := models.CollectionLoad(req.CollectionID)
	if err != nil {
		return nil, err
	}

	// Validate that Client can edit the collection
	err = validateUserIsAdmin(col, req.Client.User, req.Client.Groups)
	if err != nil {
		return nil, err
	}

	// Mark the collection edit as completed using the models package
	err = col.DoCollectionEditCompleted(req.Client.User)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
