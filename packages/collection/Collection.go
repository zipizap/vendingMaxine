package collection

import (
	"errors"
	"fmt"
	"regexp"
	"vendingMaxine/packages/xstate"

	"gorm.io/gorm"
)

// State transitions:
//   - State string:                  (never "Pending") "Running" > "Completed" or "Failed"
//   - Error() error:                 set when State=="Failed"
type Collection struct {
	gorm.Model
	Name          string          `gorm:"unique,uniqueIndex,not null"`
	ColSelections []*ColSelection // relationship 1Collection-to-manyColSelections
	dbMethods
	xstate.XState `gorm:"embedded"`
}

func collectionNew(name string) (*Collection, error) {
	// CollectionNew method should create a new object `o` and
	//
	//   - call
	//
	//     o.RegisterObserverCallback(func(oldState string, oldError error, xstate *xstate.XState) error {
	//     o.Save(o); return nil
	//     }
	//
	//   - set the new object fields from its corresponding arguments
	//
	//   - verify if the field .Name is compliant with DNS label standard as defined in RFC 1123 (like pod label-names,
	//     see https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names),
	//     and if not compliant then return an error
	//
	//   - verify the sql unique constraints, or return an error
	//
	//     If inside this method, there is any error at any step, then:
	//
	//   - dont call o.StateChange()
	//
	//   - dont call o.Save()
	//
	//   - just return the error
	//     If method is executed without errors, then:
	//
	//   - call o.StateChange("Completed", nil)
	//
	//   - return the created object o

	if !_isValidDNSLabel(name) {
		return nil, errors.New("invalid DNS label")
	}
	initialColSel, err := _colSelectionCreateInitial()
	if err != nil {
		return nil, err
	}
	o := &Collection{}
	o.Name = name
	o.ColSelections = append(o.ColSelections, initialColSel)
	o.save(o)

	o.RegisterObserverCallback(func(oldState string, oldError error, xstate *xstate.XState) error {
		o.save(o)
		return nil
	})
	err = o.StateChange("Completed", nil)
	if err != nil {
		o.StateChange("Failed", err)
		return nil, err
	}
	return o, nil
}

// collectionLoad loads from db
func collectionLoad(name string) (*Collection, error) {
	o := &Collection{}
	err := db.Where("name = ?", name).First(o).Error
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (c *Collection) appendAndRunColSelection(schema *Schema, jsonInput string, jsonOutput string, requestingUser string) error {
	err := c.reload(c) // reload object from db
	if err != nil {
		return err
	}
	if err = c._canBeUpdated(); err != nil {
		return err
	}

	var csel *ColSelection
	csel, err = newColSelection(schema, jsonInput, jsonOutput, requestingUser)
	if err != nil {
		return err
	}
	c.ColSelections = append(c.ColSelections, csel)
	err = c.save(c)
	if err != nil {
		return err
	}
	// ObserverCallback to run c.RecalculateStateAndError()
	csel.RegisterObserverCallback(
		func(oldState string, oldError error, xstate *xstate.XState) error {
			c._recalculateStateAndError(csel)
			return nil
		})
	err = csel.run()
	if err != nil {
		return err
	}
	return nil
}

func (c *Collection) colSelectionLatest() (*ColSelection, error) {
	cselsLen := len(c.ColSelections)
	csel := c.ColSelections[cselsLen-1]
	return csel, nil
}

func (c *Collection) _canBeUpdated() error {
	if c.State != "Completed" {
		return fmt.Errorf("collecion %s cannot be edited/updated as its in state %s", c.Name, c.State)
	}
	return nil
}

// _recalculateStateAndError method:
//
//	from the cs := c.ColSelections[-1], recalculate c.State and c.Error
//	  cs.State/Error                 =>  c.State/Error
//	  "Pending"/nil                  =>  "Pending"/nil
//	  "Running"/nil                  =>  "Running"/nil
//	  "Completed"/nil                =>  "Completed"/nil
//	  "Failed"/error                 =>  "Failed"/error
//	Use c.StateChange(newState, newError)
func (c *Collection) _recalculateStateAndError(csel *ColSelection) {
	_ = c.reload(c) // reload object from db
	// skip if csel != cs (cs := c.ColSelections[-1])
	cs := c.ColSelections[len(c.ColSelections)-1]
	if cs.ID != csel.ID {
		return
	}
	switch csel.State {
	case "Pending":
		c.StateChange("Pending", nil)
	case "Running":
		c.StateChange("Running", nil)
	case "Completed":
		c.StateChange("Completed", nil)
	case "Failed":
		// IMPROVEMENT: This error here should be improved to indicate the originating colSelection
		c.StateChange("Failed", csel.Error())
	default:
		panic(fmt.Sprintf("Unrecognized csel.State %s", csel.State))
	}
}

func (o *Collection) gormID() uint {
	return o.ID
}

func _isValidDNSLabel(s string) bool {
	if len(s) > 63 {
		return false
	}
	// not perfect, but good enough ;)
	r, _ := regexp.Compile("^[a-z]([-a-z0-9]*[a-z0-9])?$")
	return r.MatchString(s)
}
