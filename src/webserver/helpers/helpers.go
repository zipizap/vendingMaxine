package helpers

import (
	"os"
	"vendingMachine/src/collection"
	"vendingMachine/src/webserver/globals"
)

// This function will:
//   - Check if theCollection_name exists (and return error if not)
//   - Read consumerSelectionPreviousJson_string from theCollection.LastRsf
//   - Read productsSchemaJson_string from file PRODUCT_SCHEMA_JSON_FILEPATH
//
// 		NOTE: if lasT_rsf does not exist (ex: new collection) then err != nil
// 		and new collectino will never work
// 		todo: a newly-created-collection will never work as it does not have a last_rsf for bootstraping
//
func Get_selectionPrevious_and_prodSchema_from_collection(theCollection_name string) (consumerSelectionPreviousJson_string string, productsSchemaJson_string string, err error) {
	// if theCollection does not exist, reply with error
	theCollection, err := collection.GetCollection(theCollection_name)
	if err != nil {
		return "", "", err
	}

	// Read consumerSelectionPreviousJson_string from theCollection.LastRsf
	last_rsf, err := theCollection.LastRsf()
	// NOTE: if lasT_rsf does not exist (ex: new collection) then err != nil
	// and new collectino will never work
	// todo: a newly-created-collection will never work as it does not have a last_rsf for bootstraping
	if err != nil {
		return "", "", err
	}
	consumerSelectionPreviousJson_string, err = last_rsf.Get_OverallLatestUpdateData_Decoded("consumer-selection.next.json")
	if err != nil {
		return "", "", err
	}

	// Read productsSchemaJson_string from file
	productsSchemaJson_bytes, err := os.ReadFile(globals.PRODUCT_SCHEMA_JSON_FILEPATH)
	if err != nil {
		return "", "", err
	}
	productsSchemaJson_string = string(productsSchemaJson_bytes)
	return consumerSelectionPreviousJson_string, productsSchemaJson_string, nil
}
