package xstate

import "fmt"

/*
======= USE IT LIKE THIS ========

import (
  "vendingMaxine/packages/xstate"
)

// 1.1. embed into struct
type Mys struct {
  xstate.XState  `gorm:"embedded"`
}

// 1.2. constructor: set initial state
func NewMys() (*Mys, err) {
	mys := &Mys{
		XState: xstate.XState{
			State: "Initializing",
		}
	}
	...
	return mys
}

// 1.3. define Mys.StateChangePostHandle()
// - to implement interface StateChangePostHandler
// - to make any db.Save() or callback methods from other objects
func (o *Mys) StateChangePostHandle(oldState string, oldError error, newXstate *xstate.XState) error {
	o.save(o)
	p := o.someParentObj
	return p.StateChange(p, newXstate.State, newXstate.Error())
}

// 2. Change state when you need (it will implicitly call mys.StateChangePostHandle())
err := mys.StateChange(mys, "Running", nil)
...

*/

type StateChangePostHandler interface {
	StateChangePostHandle(oldState string, oldError error, xstate *XState) error
}

type XState struct {
	State       string // "Pending" > "Running" > "Completed" or "Failed"
	ErrorString string // set non-empty when State=="Failed"
}

func (x *XState) Error() error {
	if x.ErrorString == "" {
		return nil
	} else {
		return fmt.Errorf(x.ErrorString)
	}
}

func (x *XState) StateChange(i StateChangePostHandler, nextState string, nextError error) error {
	oldState := x.State
	oldError := x.Error()
	x.State = nextState
	if nextError != nil {
		x.ErrorString = nextError.Error()
	} else {
		x.ErrorString = ""
	}
	err := i.StateChangePostHandle(oldState, oldError, x)
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
