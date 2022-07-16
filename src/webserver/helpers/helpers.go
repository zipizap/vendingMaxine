package helpers

import (
	"os"
	"vendingMaxine/src/collection"
	"vendingMaxine/src/webserver/globals"
)

func Get_selectionPrevious_and_prodSchema_from_collection(theCollection_name string) (cantDo bool, consumerSelectionPreviousJson_string string, productsSchemaJson_string string, err error) {
	// This function will:
	//   a) Check if theCollection_name exists (and return error if not)
	//	 b) Check if theCollection.NewRsf_canBeCreated
	//   c) Read consumerSelectionPreviousJson_string from theCollection.LastRsf
	//   d) Read productsSchemaJson_string from file PRODUCT_SCHEMA_JSON_FILEPATH
	//
	// 		NOTE: if last_rsf does not exist (ex: new collection) then err != nil
	// 		and new collectino will never work
	// 		todo: a newly-created-collection might never work as it might not have a last_rsf for initial bootstraping

	//   a) Check if theCollection_name exists (and return error if not)
	theCollection, err := collection.GetCollection(theCollection_name)
	if err != nil {
		return false, "", "", err
	}
	//	 b) Check if theCollection.NewRsf_canBeCreated
	yes, err := theCollection.NewRsf_canBeCreated()
	if err != nil {
		return false, "", "", err
	}
	if !yes {
		cantDo = true
		return cantDo, "", "", nil
	}

	//   c) Read consumerSelectionPreviousJson_string from theCollection.LastRsf
	last_rsf, err := theCollection.LastRsf()
	// NOTE: if lasT_rsf does not exist (ex: new collection) then err != nil
	// and new collectino will never work
	// todo: a newly-created-collection will never work as it does not have a last_rsf for bootstraping
	if err != nil {
		return false, "", "", err
	}
	consumerSelectionPreviousJson_string, err = last_rsf.Get_OverallLatestUpdateData_Decoded("consumer-selection.next.json")
	if err != nil {
		return false, "", "", err
	}

	//   d) Read productsSchemaJson_string from file PRODUCT_SCHEMA_JSON_FILEPATH
	productsSchemaJson_bytes, err := os.ReadFile(globals.PRODUCT_SCHEMA_JSON_FILEPATH)
	if err != nil {
		return false, "", "", err
	}
	productsSchemaJson_string = string(productsSchemaJson_bytes)
	return false, consumerSelectionPreviousJson_string, productsSchemaJson_string, nil
}
