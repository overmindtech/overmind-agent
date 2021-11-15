package command

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"

	"github.com/alessio/shellescape"
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

	// STDIN specifies the binary data that should be piped to the command as
	// STDIN. This can be used for example to simulate user intaction for
	// programs that read from STDIN. This will be encoded using base64 to a
	// string in JSON
	STDIN []byte `json:"stdin"`

	// RunAs specifies the user that the command should be run as
	RunAs string `json:"run_as"`
}

// MarshalJSON Converts the object to JSON
func (cp CommandParams) MarshalJSON() ([]byte, error) {
	// Modify the Timeout parameter so that it can be stored in a more readable
	// format (i.e. a string)
	type Alias CommandParams
	return json.Marshal(&struct {
		Timeout string `json:"timeout"`
		STDIN   string `json:"stdin"`
		*Alias
	}{
		Timeout: cp.Timeout.String(),
		STDIN:   base64.StdEncoding.EncodeToString(cp.STDIN),
		Alias:   (*Alias)(&cp),
	})
}

// UnmarshalJSON Converts the object from JSON
func (cp *CommandParams) UnmarshalJSON(data []byte) error {
	var timeoutDuration time.Duration
	var stdin []byte
	var err error

	type Alias CommandParams
	aux := &struct {
		Timeout string `json:"timeout"`
		STDIN   string `json:"stdin"`
		*Alias
	}{
		Alias: (*Alias)(cp),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse the `timeout` parameter using time.ParseDuration meaning that the
	// duration should be in a human readable format e.g. "300ms", "-1.5h" or
	// "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	timeoutDuration, err = time.ParseDuration(aux.Timeout)
	if err != nil {
		return err
	}

	cp.Timeout = timeoutDuration

	// Decode the STDIN from base64 to bytes
	stdin, err = base64.StdEncoding.DecodeString(aux.STDIN)

	if err != nil {
		return err
	}

	cp.STDIN = stdin
	return nil
}

func (cp *CommandParams) Run() (*sdp.Item, error) {
	if cp.Timeout == 0 {
		// Use default timeout if none was specified
		cp.Timeout = DefaultTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), cp.Timeout)
	defer cancel()

	var commandString string
	var args []string
	var err error

	if runtime.GOOS == "windows" {
		commandString, args, err = PowerShellWrap(cp.Command, cp.Args)
	} else {
		commandString, args, err = ShellWrap(cp.Command, cp.Args)
	}

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	command := exec.CommandContext(ctx, commandString, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	command.Stdout = &stdout
	command.Stderr = &stderr
	command.Dir = cp.Dir
	command.Env = envToString(mergeEnv(cp.Env))

	err = command.Run()

	if err != nil {
		// This will return an error if the context has ended
		if ctx.Err() == context.DeadlineExceeded {
			return nil, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: fmt.Sprintf("command execution timed out.\nSTDOUT: %v\nSTDERR: %v", stdout.String(), stderr.String()),
				Context:     util.LocalContext,
			}
		}

		switch e := err.(type) {
		case *exec.ExitError:
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
		"stdout":   strings.TrimSuffix(stdout.String(), platformNewline()),
		"stderr":   strings.TrimSuffix(stderr.String(), platformNewline()),
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

// ShellWrap Wraps a given command and args in the required arguments so that
// the command runs inside a shell, the default being bash, but falling back to
// sh if bash is not available
func ShellWrap(command string, args []string) (string, []string, error) {
	var shell string

	// Get the full path to the shell, bash or sh
	if bashPath, err := exec.LookPath("bash"); err == nil {
		shell = bashPath
	} else if shPath, err := exec.LookPath("sh"); err == nil {
		shell = shPath
	} else {
		return "", []string{}, errors.New("could not find bash or sh on the PATH")
	}

	// Join the command and args into a single array
	joined := []string{command}
	joined = append(joined, args...)

	// Escape this so that it can be passed as a single thing
	escaped := shellescape.QuoteCommand(joined)

	// Create the arguments which are basically /bin/bash -c {your stuff here}
	args = []string{"-c", escaped}

	return shell, args, nil
}

// PowerShellWrap Wraps a given command and args in the required arguments so
// that the command runs inside powershell
func PowerShellWrap(command string, args []string) (string, []string, error) {
	// Encode the original command as base64
	powershellArray := []string{command}
	powershellArray = append(powershellArray, args...)

	// Append deliberate exit to ensure that the powershell.exe process exits
	// with a code other than 1
	powershellArray = append(powershellArray, "; exit $LASTEXITCODE")
	powershellCommand := strings.Join(powershellArray, " ")

	powershellArgs := []string{
		"-NoLogo",          // Hides the copyright banner at startup
		"-NoProfile",       // Does not load the Windows PowerShell profile
		"-NonInteractive",  // Does not present an interactive prompt to the user
		"-ExecutionPolicy", // Allow running of unsigned code
		"bypass",
		powershellCommand,
	}

	// TODO: The stuff that I'm doing hwere means that the actual powershell
	// process will exit with 1 if the command fails, regardless of the exit
	// code of the command itself. I need to find some way to pass this though.
	// Read this:
	// https://stackoverflow.com/questions/50200325/returning-an-exit-code-from-a-powershell-script

	return "powershell.exe", powershellArgs, nil
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

// platformNewline Returns the newline string for this platform (\n or \r\n)
func platformNewline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}
}

// mergeEnv Merges the current environment variables with the supplied ones, the
// supplied ones take precedence
func mergeEnv(vars map[string]string) map[string]string {
	merged := make(map[string]string)

	// Get Current environment vars
	for _, env := range os.Environ() {
		split := strings.Split(env, "=")

		if len(split) == 2 {
			merged[split[0]] = split[1]
		}
	}

	for k, v := range vars {
		merged[k] = v
	}

	return merged
}
