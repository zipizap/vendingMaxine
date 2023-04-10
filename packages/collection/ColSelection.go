package collection

import (
	"fmt"
	"vendingMaxine/packages/xstate"

	"gorm.io/gorm"
)

// State transitions:
//   - State string:   "Pending" > "Running" > "Completed" or "Failed"
//   - Error() error:  set when State=="Failed"
type ColSelection struct {
	gorm.Model
	CollectionID           uint    // relationship 1Collection-to-manyColSelections
	Schema                 *Schema // relationship manyColSelections-to-1Schema
	SchemaID               uint    // relationship manyColSelections-to-1Schema
	JsonInput              string
	JsonOutput             string
	RequestingUser         string
	ProcessingEngineRunner *ProcessingEngineRunner
	dbMethods
	xstate.XState `gorm:"embedded"`
}

func newColSelection(schema *Schema, jsonInput string, jsonOutput string, requestingUser string) (*ColSelection, error) {
	// newColSelection method should create a new object o and
	//   - call o.RegisterObserverCallback(func(oldState string, oldError error, xstate *xstate.XState) error {
	//     o.Save(o); return nil
	//     }
	//   - set the new object fields from its corresponding arguments
	//   - check all possible errors
	//     If inside this method, there is any error at any step, then:
	//   - call o.StateChange("Failed", error) and return the error
	//     If method is executed without errors, then:
	//   - call o.StateChange("Pending", nil)
	//   - return the created object o

	// validate schema.ID == SchemaLatest().ID
	schemaLatest, err := schemaLoadLatest()
	if err != nil {
		return nil, err
	}
	if schema.ID != schemaLatest.ID {
		// schema is not the same as schemaLatest
		return nil, fmt.Errorf("used schema is not schemaLatest")
	}

	o := &ColSelection{
		Schema:         schemaLatest, // we need to assure the Schema field only gets an already-existing-in-db schema to avoid creating it unintentionally. Using schemaLatest achieves this effect
		JsonInput:      jsonInput,
		JsonOutput:     jsonOutput,
		RequestingUser: requestingUser,
	}

	o.RegisterObserverCallback(func(oldState string, oldError error, xstate *xstate.XState) error {
		o.save(o)
		return nil
	})
	err = o.StateChange("Pending", nil)
	if err != nil {
		o.StateChange("Failed", err)
		return nil, err
	}

	return o, nil
}

func (csel *ColSelection) run() error {
	// run method should
	//   - set o.ProcessingEngineRunner,err=NewProcessingEngineRunner()
	//   - set o.ProcessingEngineRunner.RegisterObserverCallback( to call o.recalculateStateAndError(...) )
	//   - call err = o.ProcessingEngineRunner.run()
	//   - check all possible errors
	//     If inside this method, there is any error at any step, then:
	//   - call o.StateChange("Failed", error) and return the error

	err := csel.reload(csel) // reload object from db
	if err != nil {
		return err
	}
	per, err := newProcessingEngineRunner()
	csel.ProcessingEngineRunner = per
	csel.save(csel)
	if err != nil {
		csel.StateChange("Failed", err)
		return err
	}
	// ObserverCallback to run csel.RecalculateStateAndError()
	per.RegisterObserverCallback(
		func(oldState string, oldError error, xstate *xstate.XState) error {
			csel._recalculateStateAndError(per)
			return nil
		})
	err = per.run()
	if err != nil {
		csel.StateChange("Failed", err)
		return err
	}
	return nil
}

// _recalculateStateAndError method: from the per := csel.ProcessingEngineRunner, recalculate csel.State and csel.Error
//
//	  per.State/Error                 =>  csel.State/Error
//	  "Pending"/nil                  =>  "Pending"/nil
//	  "Running"/nil                  =>  "Running"/nil
//	  "Completed"/nil                =>  "Completed"/nil
//	  "Failed"/error                 =>  "Failed"/error
//	Use csel.StateChange(newState, newError)
func (csel *ColSelection) _recalculateStateAndError(per *ProcessingEngineRunner) {
	_ = csel.reload(csel) // reload object from db

	// skip if per != csel.ProcessingEngineRunner
	if per.ID != csel.ProcessingEngineRunner.ID {
		return
	}
	switch per.State {
	case "Pending":
		csel.StateChange("Pending", nil)
	case "Running":
		csel.StateChange("Running", nil)
	case "Completed":
		csel.StateChange("Completed", nil)
	case "Failed":
		// IMPROVEMENT: This error here should be improved to indicate the originating per
		csel.StateChange("Failed", per.Error())
	default:
		panic(fmt.Sprintf("Unrecognized per.State %s", per.State))
	}
}

func _colSelectionCreateInitial() (*ColSelection, error) {
	latestSchema, err := schemaLoadLatest()
	if err != nil {
		return nil, err
	}
	initialJsonInput := "{}"
	initialJsonOutput := "{}"
	initialRequestingUser := "init"
	initialColSel, err := newColSelection(latestSchema, initialJsonInput, initialJsonOutput, initialRequestingUser)
	if err != nil {
		return nil, err
	}
	return initialColSel, nil
}

func (o *ColSelection) gormID() uint {
	return o.ID
}
