package command

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

const DefaultTimeout = 10 * time.Second

type CommandParams struct {
	// Command specifies the command to run
	Command string `json:"command"`

	// Args specifies to pass to the command
	Args []string `json:"args"`

	// ExpectedExit is the expected exit code (usually 0)
	ExpectedExit int `json:"expected_exit"`

	// Timeout before cancelling the command
	Timeout time.Duration `json:"timeout"`

	// Dir specifies the working directory of the command.
	Dir string `json:"dir"`

	// Env specifies environment variables that should be set when running the
	// command
	Env map[string]string `json:"env"`
}

func (cp *CommandParams) Run() (*sdp.Item, error) {
	if cp.Timeout == 0 {
		// Use default timeout if none was specified
		cp.Timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cp.Timeout)
	defer cancel()

	// TODO: Run inside a shell
	// TODO: Handle using powershell on windows
	command := exec.CommandContext(ctx, cp.Command, cp.Args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command.Stdout = &stdout
	command.Stderr = &stderr
	command.Dir = cp.Dir
	command.Env = envToString(cp.Env)

	err := command.Run()

	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			if e.Error() == "signal: killed" {
				return nil, &sdp.ItemRequestError{
					ErrorType:   sdp.ItemRequestError_OTHER,
					ErrorString: fmt.Sprintf("command execution timed out.\nSTDOUT: %v\nSTDERR: %v", stdout.String(), stderr.String()),
					Context:     util.LocalContext,
				}
			}

			if e.ExitCode() != cp.ExpectedExit {
				return nil, &sdp.ItemRequestError{
					ErrorType:   sdp.ItemRequestError_OTHER,
					ErrorString: fmt.Sprintf("command execution failed. Error: %v\nSTDOUT: %v\nSTDERR: %v", err, stdout.String(), stderr.String()),
					Context:     util.LocalContext,
				}
			}
		default:
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: fmt.Sprintf("command execution failed. Error: %v\nSTDOUT: %v\nSTDERR: %v", err, stdout.String(), stderr.String()),
				Context:     util.LocalContext,
			}
		}
	}

	var attributes *sdp.ItemAttributes

	attributes, err = sdp.ToAttributes(map[string]interface{}{
		"name":     cp.Command,
		"args":     cp.Args,
		"exitCode": command.ProcessState.ExitCode(),
		"stdout":   strings.TrimSuffix(stdout.String(), "\n"),
		"stderr":   strings.TrimSuffix(stderr.String(), "\n"),
	})

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("error during attribute conversion: %v", err),
			Context:     util.LocalContext,
		}
	}

	item := sdp.Item{
		Type:            "command",
		UniqueAttribute: "name",
		Context:         util.LocalContext,
		Attributes:      attributes,
		LinkedItemRequests: []*sdp.ItemRequest{
			{
				Type:    "file",
				Method:  sdp.RequestMethod_GET,
				Query:   cp.Command,
				Context: util.LocalContext,
			},
		},
	}

	return &item, nil
}

// envToString Converts a map of environment variables to an array of equals
// separated strings
func envToString(envs map[string]string) []string {
	envStrings := make([]string, 0)

	for k, v := range envs {
		envStrings = append(envStrings, fmt.Sprintf("%v=%v", k, v))
	}

	return envStrings
}
