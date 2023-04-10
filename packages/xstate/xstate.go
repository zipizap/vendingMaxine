package xstate

import "fmt"

/*
======= USE IT LIKE THIS ========

import (
  "vendingMaxine/packages/xstate"
)

// 1. embed into struct
type Mys struct {
  xstate.XState  `gorm:"embedded"`
}

// 2. constructor: set initial state, register callback
func NewMys() (*Mys, err) {
	mys := &Mys{
		XState: xstate.XState{
			State: "Initializing",
		}
	}
	mys.RegisterObserverCallback(
		func(oldState string, oldError error, xstate *XState) error {
			mys.SaveToDb()
		}
	)
	...
	return mys
}

// 3. Change state when you need (it will implicitly run all observerCallback functions)
err := mys.StateChange("Running", nil)
...

*/

type XStateObserverCallback func(oldState string, oldError error, xstate *XState) error

type XState struct {
	State             string                   // "Pending" > "Running" > "Completed" or "Failed"
	ErrorString       string                   // set non-empty when State=="Failed"
	observerCallbacks []XStateObserverCallback // called after each state change
}

func (x *XState) Error() error {
	if x.ErrorString == "" {
		return nil
	} else {
		return fmt.Errorf(x.ErrorString)
	}
}

func (x *XState) RegisterObserverCallback(xso XStateObserverCallback) {
	x.observerCallbacks = append(x.observerCallbacks, xso)
}

func (x *XState) StateChange(nextState string, nextError error) error {
	oldState := x.State
	oldError := x.Error()
	x.State = nextState
	if nextError != nil {
		x.ErrorString = nextError.Error()
	} else {
		x.ErrorString = ""
	}
	for _, a_callback := range x.observerCallbacks {
		err := a_callback(oldState, oldError, x)
		if err != nil {
			return err
		}
	}
	return nil
}
