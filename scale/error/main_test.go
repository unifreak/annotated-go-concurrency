package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime/debug"
	"testing"
)

type MyError struct {
	Inner      error
	Message    string
	StackTrace string
	Misc       map[string]interface{}
}

// wrapError is a helper for our low-level and intermediate module to wrap error
// into a well-formed error. Wrapped error will be considered handled correctly
// and can be directly display to user.
func wrapError(err error, messagef string, msgArgs ...interface{}) MyError {
	return MyError{
		Inner:      err,
		Message:    fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc:       make(map[string]interface{}),
	}
}

func (err MyError) Error() string {
	return err.Message
}

// "lowlevel" module

type LowLevelErr struct {
	error
}

func isGloballyExec(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, LowLevelErr{(wrapError(err, err.Error()))} // <1>
	}
	return info.Mode().Perm()&0100 == 0100, nil
}

// "intermediate" module

type IntermediateErr struct {
	error
}

func runJob(id string) error {
	const jobBinPath = "/bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {	// PROBLEMATIC! not handling error correctly
		return err
	} else if isExecutable == false {
		return wrapError(nil, "job binary is not executable")
	}

	return exec.Command(jobBinPath, "--id="+id).Run()
}

func runJobFixed(id string) error {
	const jobBinPath = "/bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {
		// Wrap the lowlevel error in our own module's error type, here we can
		// also obfuscate the low-level details.
		return IntermediateErr{wrapError(
			err,
			"cannot run job %q: requisite binaries not available",
			id,
		)}
	} else if isExecutable == false {
		return wrapError(
			nil,
			"cannot run job %q: job binary is not executable",
			id,
		)
	}

	return exec.Command(jobBinPath, "--id="+id).Run()
}
func handleError(key int, err error, message string) {
	// Log error
	log.SetPrefix(fmt.Sprintf("[logID: %v]: ", key))
	log.Printf("%#v", err)
	// Display error
	fmt.Printf("[%v] %v", key, message)
}

func TestNotWrappingErr(_ *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	err := runJob("1")
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug.\n"
		if _, ok := err.(IntermediateErr); ok { // is expected, well-formed error
			msg = err.Error() + "\n"			// we can display its msg to user
		}
		handleError(1, err, msg)
	}
}

func TestWrappingErr(_ *testing.T) {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	err := runJobFixed("1")	// fixed wrapping
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug.\n"
		if _, ok := err.(IntermediateErr); ok { // is expected, well-formed error
			msg = err.Error() + "\n"			// we can display its msg to user
		}
		handleError(1, err, msg)
	}
}