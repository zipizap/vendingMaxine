package collection

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"gorm.io/gorm"
)

var peDirpath string

func initProcessingEngineRunner(processingEnginesDirpath string) {
	peDirpath = processingEnginesDirpath
}

// State transitions:
//   - State string:   "Pending" > "Running" > "Completed" or "Failed"
//   - Error() error:  set when State=="Failed"
type ProcessingEngineRunner struct {
	gorm.Model
	ColSelectionID    uint
	ColSelection      *ColSelection
	ProcessingEngines []*ProcessingEngine
	dbMethods
	XState `gorm:"embedded"`
}

func newProcessingEngineRunner() (*ProcessingEngineRunner, error) {
	// newProcessingEngineRunner creates a new object o and
	//   - call o.RegisterObserverCallback(func(oldState string, oldError error, xstate *XState) error {
	//     o.Save(o); return nil
	//     }
	//   - set the new object fields from its corresponding arguments
	//     Should check all possible errors.
	//     If inside this method, there is any error at any step, then:
	//   - call o.StateChange("Failed", error) and return the error
	//     If method is executed without errors, then:
	//   - call o.StateChange("Pending", nil)
	//   - return the created object o

	o := &ProcessingEngineRunner{}
	err := o.stateChange(o, "Pending", nil)
	if err != nil {
		o.stateChange(o, "Failed", err)
		return nil, err
	}
	return o, nil
}
func (o *ProcessingEngineRunner) stateChangePostHandleXState(oldState string, oldError error, newXstate *XState) error {
	err := o.save(o)
	if err != nil {
		return err
	}
	if o.ColSelection != nil {
		o.ColSelection._recalculateStateAndError(o)
	}
	return nil
}

func (per *ProcessingEngineRunner) run() error {
	// run lists all linux executable files found in peDirpath in lexicographical descending order,
	//
	//	and for each binPath found:
	//	  - create a corresponding ProcessingEngine instance pe,err=NewProcessingEngine(binPath, []string{})
	//	  - append pe to o.ProcessingEngines slice
	//	  - call pe.RegisterObserverCallback( to call o.recalculateStateAndError(...) )
	//	  - call err = pe.run()
	//	Should check all possible errors.
	//	If inside this method, there is any error at any step, then:
	//	+ call o.StateChange("Failed", error) and return the error

	err := per.reload(per) // reload object from db
	if err != nil {
		return err
	}
	// This method should list all linux executable files found in peDirpath in lexicographical descending order.
	binpathsDir := peDirpath
	var binPaths []string
	{
		files, err := ioutil.ReadDir(binpathsDir)
		if err != nil {
			per.stateChange(per, "Failed", err)
			return err
		}

		for _, file := range files {
			if file.Mode().IsRegular() && (file.Mode()&0111 != 0) {
				binPaths = append(binPaths, filepath.Join(binpathsDir, file.Name()))
			}
		}

		sort.Slice(binPaths, func(i, j int) bool {
			return binPaths[i] < binPaths[j]
		})
	}

	// For each binPath found:
	for _, binPath := range binPaths {
		pe, err := newProcessingEngine(binPath, []string{})
		if err != nil {
			per.stateChange(per, "Failed", err)
			return err
		}
		per.ProcessingEngines = append(per.ProcessingEngines, pe)
		err = per.save(per)
		if err != nil {
			return err
		}
		err = pe.run()
		err2 := per.reload(per) // reload object from db
		if err2 != nil {
			return err2
		}
		if err != nil {
			per.stateChange(per, "Failed", err)
			return err
		}
	} // end for
	return nil
}

// _recalculateStateAndError recalculates per.State and per.Error from pe.State and pe.Error
//
//	From the pe := per.ProcessingEngine[-1], recalculate per.State and per.Error
//	  pe.State/Error                 =>  per.State/Error
//	  "Pending"/nil                  =>  "Pending"/nil
//	  "Running"/nil                  =>  "Running"/nil
//	  "Completed"/nil                =>  "Completed"/nil
//	  "Failed"/error                 =>  "Failed"/error
//	Use per.StateChange(newState, newError)
func (per *ProcessingEngineRunner) _recalculateStateAndError(pe *ProcessingEngine) {
	_ = per.reload(per) // reload object from db
	_ = pe.reload(pe)   // reload object from db

	// skip if pe != per.ProcessingEngine[-1]
	if pe.ID != per.ProcessingEngines[len(per.ProcessingEngines)-1].ID {
		return
	}
	switch pe.State {
	case "Pending":
		per.stateChange(per, "Pending", nil)
	case "Running":
		per.stateChange(per, "Running", nil)
	case "Failed":
		// IMPROVEMENT: This error here should be improved to indicate the originating per
		per.stateChange(per, "Failed", pe.error())
	case "Completed":
		per.stateChange(per, "Completed", nil)
	}
}

func (o *ProcessingEngineRunner) gormID() uint {
	return o.ID
}
