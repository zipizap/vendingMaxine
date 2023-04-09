package collection

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
	"vd-alpha/packages/xstate"

	"gorm.io/gorm"
)

type ProcessingEngine struct {
	gorm.Model
	ProcessingEngineRunnerID uint
	BinPath                  string    // relative path to processing-engine binary
	BinLastModTime           time.Time // bin-file LastModTime
	RunStartTime             time.Time
	RunEndTime               time.Time
	RunArg1                  string
	RunArg2                  string
	RunArg3                  string
	RunArg4                  string
	RunArg5                  string
	RunStdout                string
	RunStderr                string
	RunExitcode              int
	dbMethods
	xstate.XState `gorm:"embedded"`
	// State string // "Pending" > "Running" > "Completed" or "Failed"
	// Error error                   // set when State=="Failed"
}

// newProcessingEngine method should create a new object o and
//   - call o.RegisterObserverCallback(func(oldState string, oldError error, xstate *XState) error {
//     o.Save(o)
//     º
//   - set the new object fields from its corresponding arguments
//     Should check all possible errors.
//     If inside this method, there is any error at any step, then:
//   - call o.StateChange("Failed", error) and return the error
//     If method is executed without errors, then:
//   - call o.StateChange("Pending", nil)
//   - return the created object o
func newProcessingEngine(binPath string, runArgs []string) (*ProcessingEngine, error) {
	//   - set the new object fields from its corresponding arguments
	o := &ProcessingEngine{
		BinPath: binPath,
	}
	// set o.RunArg1..5 - this is ugly but quick and works
	{
		if len(runArgs) == 1 {
			o.RunArg1 = runArgs[0]
		} else if len(runArgs) == 2 {
			o.RunArg1 = runArgs[0]
			o.RunArg2 = runArgs[1]
		} else if len(runArgs) == 3 {
			o.RunArg1 = runArgs[0]
			o.RunArg2 = runArgs[1]
			o.RunArg3 = runArgs[2]
		} else if len(runArgs) == 4 {
			o.RunArg1 = runArgs[0]
			o.RunArg2 = runArgs[1]
			o.RunArg3 = runArgs[2]
			o.RunArg4 = runArgs[3]
		} else if len(runArgs) == 5 {
			o.RunArg1 = runArgs[0]
			o.RunArg2 = runArgs[1]
			o.RunArg3 = runArgs[2]
			o.RunArg4 = runArgs[3]
			o.RunArg5 = runArgs[4]
		} else if len(runArgs) > 5 {
			return nil, fmt.Errorf("runArgs limited to at most 5 args")
		}
	}
	//   - call o.RegisterObserverCallback(func(oldState string, oldError error, xstate *XState) {
	//     o.Save(o)
	//   }
	o.RegisterObserverCallback(func(oldState string, oldError error, xstate *xstate.XState) error {
		o.save(o)
		return nil
	})

	o.StateChange("Pending", nil)
	return o, nil
}

// run method should:
//   - calculate and set o.BinLastModTime
//   - run the BinPath binary and fill all the o.Runxxx fields
//     If RunExitcode != 0 then call
//     o.StateChange("Failed", fmt.Errorf("ProcessingEngine %s gave exit-code %d", o.BinPath, o.RunExitcode))
//     and return that error
//     Should check all possible errors
//     If inside this method, there is any error at any step, then:
//   - call o.StateChange("Failed", error) and return the error
func (pe *ProcessingEngine) run() error {
	err := pe.reload(pe) // reload object from db
	if err != nil {
		return err
	}

	//   - calculate and set o.BinLastModTime
	fileInfo, err := os.Stat(pe.BinPath)
	if err != nil {
		pe.StateChange("Failed", err)
		return err
	}
	pe.BinLastModTime = fileInfo.ModTime()

	//   - run the BinPath binary and fill all the o.Runxxx fields
	// If RunExitcode != 0 then call
	//   o.StateChange("Failed", fmt.Errorf("ProcessingEngine %s gave exit-code %d", o.BinPath, o.RunExitcode))
	//   and return that error
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.Command(pe.BinPath, pe.RunArg1, pe.RunArg2, pe.RunArg3, pe.RunArg4, pe.RunArg5)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	pe.RunStartTime = time.Now()
	err = cmd.Run()
	pe.RunEndTime = time.Now()
	pe.RunExitcode = cmd.ProcessState.ExitCode()
	pe.RunStderr = stderrBuf.String()
	pe.RunStdout = stdoutBuf.String()
	if pe.RunExitcode != 0 {
		err := fmt.Errorf("ProcessingEngine %s gave exit-code %d", pe.BinPath, pe.RunExitcode)
		pe.StateChange("Failed", err)
		return err
	} else if err != nil {
		pe.StateChange("Failed", err)
		return err
	}

	pe.StateChange("Completed", nil)
	return nil
}

func (o *ProcessingEngine) gormID() uint {
	return o.ID
}
func (o *ProcessingEngine) runArgs() []string {
	args := []string{o.RunArg1, o.RunArg2, o.RunArg3, o.RunArg4, o.RunArg5}
	var runArgs []string
	for _, a_arg := range args {
		if a_arg != "" {
			runArgs = append(runArgs, a_arg)
		}
	}
	return runArgs
}
