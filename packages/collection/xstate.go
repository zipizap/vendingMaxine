package collection

import "fmt"

/*
======= USE IT LIKE THIS ========

import (
  "vendingMaxine/packages/xstate"
)

// 1.1. embed into struct
type Mys struct {
  XState  `gorm:"embedded"`
}

// 1.2. constructor: set initial state
func NewMys() (*Mys, err) {
	mys := &Mys{
		XState: XState{
			State: "Initializing",
		}
	}
	...
	return mys
}

// 1.3. define Mys.stateChangePostHandleXState()
// - to implement interface stateChangePostHandleXStater
// - to make any db.Save() or callback methods from other objects
func (o *Mys) stateChangePostHandleXState(oldState string, oldError error, newXstate *XState) error {
	o.save(o)
	p := o.someParentObj
	return p.StateChange(p, newXstate.State, newXstate.Error())
}

// 2. Change state when you need (it will implicitly call mys.stateChangePostHandleXState())
err := mys.StateChange(mys, "Running", nil)
...

*/

type stateChangePostHandleXStater interface {
	stateChangePostHandleXState(oldState string, oldError error, xstate *XState) error
}

type XState struct {
	State       string // "Pending" > "Running" > "Completed" or "Failed"
	ErrorString string // set non-empty when State=="Failed"
}

func (x *XState) error() error {
	if x.ErrorString == "" {
		return nil
	} else {
		return fmt.Errorf(x.ErrorString)
	}
}

func (x *XState) stateChange(i stateChangePostHandleXStater, nextState string, nextError error) error {
	oldState := x.State
	oldError := x.error()
	x.State = nextState
	if nextError != nil {
		x.ErrorString = nextError.Error()
	} else {
		x.ErrorString = ""
	}
	err := i.stateChangePostHandleXState(oldState, oldError, x)
	if err != nil {
		return err
	}
	return nil
}

// type XStateObserverCallback func(oldState string, oldError error, xstate *XState) error
//
// func (x *XState) RegisterObserverCallback(xso XStateObserverCallback) {
// 	x.observerCallbacks = append(x.observerCallbacks, xso)
// }
//
// func (x *XState) StateChange(nextState string, nextError error) error {
// 	oldState := x.State
// 	oldError := x.Error()
// 	x.State = nextState
// 	if nextError != nil {
// 		x.ErrorString = nextError.Error()
// 	} else {
// 		x.ErrorString = ""
// 	}
// 	for _, a_callback := range x.observerCallbacks {
// 		err := a_callback(oldState, oldError, x)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }
