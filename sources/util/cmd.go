package util

import (
	"fmt"
	"os"
	"os/exec"
)

// Run runs a command and returns only the bytemap, panicing on error
//
// Accepts an array of strings which will be joined by spaces
func Run(args []string) []byte {
	var out []byte
	var err error

	command := exec.Command(args[0], args[1:]...)
	command.Env = os.Environ()
	command.Env = append(command.Env, "LANG=en_US.UTF-8")
	out, err = command.Output()

	// if there is an error with our execution
	// handle it here
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		fmt.Fprintf(os.Stderr, "%v", string(out))
		return []byte{}
	}

	return out
}
