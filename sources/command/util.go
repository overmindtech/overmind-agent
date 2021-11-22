package command

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

const DefaultTimeout = 10 * time.Second

type CommandParams struct {
	// Command specifies the command to run, including all arguments
	Command string `json:"command"`

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
}

type UserInfo struct {
	// Username The user to run the command as
	Username string `json:"username"`
	// Password (optional) The password required for that user. On linux this is the
	// password that will be provided to sudo, in wondows this is the
	// password of the user themselves
	Password string `json:"password,omitempty"`
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

// Run Runs the command and returns an item and error
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

	// Wrap in a shell
	switch runtime.GOOS {
	case "windows":
		commandString, args, err = cp.PowerShellWrap()
	default:
		commandString, args, err = cp.ShellWrap()
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
	var stdinPipe io.WriteCloser

	stdinPipe, err = command.StdinPipe()

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	command.Stderr = &stderr
	command.Stdout = &stdout
	command.Dir = cp.Dir
	command.Env = envToString(mergeEnv(cp.Env))

	// Start the command
	if err := command.Start(); err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	// Start reading/writing
	go stdinPipe.Write(cp.STDIN)

	// Wait for the command to finish
	err = command.Wait()

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
func (cp *CommandParams) ShellWrap() (string, []string, error) {
	all := make([]string, 0)

	// Get the full path to the shell, bash or sh
	if bashPath, err := exec.LookPath("bash"); err == nil {
		all = append(all, bashPath)
	} else if shPath, err := exec.LookPath("sh"); err == nil {
		all = append(all, shPath)
	} else {
		return "", []string{}, errors.New("could not find bash or sh on the PATH")
	}

	all = append(all, []string{"-c", cp.Command}...)

	return all[0], all[1:], nil
}

// PowerShellWrap Wraps a given command and args in the required arguments so
// that the command runs inside powershell
func (cp *CommandParams) PowerShellWrap() (string, []string, error) {
	powershellArgs := []string{
		"powershell.exe",
		"-NoLogo",          // Hides the copyright banner at startup
		"-NoProfile",       // Does not load the Windows PowerShell profile
		"-NonInteractive",  // Does not present an interactive prompt to the user
		"-ExecutionPolicy", // Allow running of unsigned code
		"bypass",
		cp.Command + "; exit $LASTEXITCODE",
	}

	return powershellArgs[0], powershellArgs[1:], nil
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
