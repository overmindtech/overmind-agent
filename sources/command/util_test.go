package command

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/overmindtech/discovery"
)

func TestRun(t *testing.T) {
	t.Run("with working comand", func(t *testing.T) {
		params := CommandParams{
			Command:      "hostname",
			ExpectedExit: 0,
			Timeout:      1 * time.Second,
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItem(t, item)

		var name interface{}
		var exitCode interface{}
		var stdout interface{}
		var hostname string

		name, err = item.Attributes.Get("name")

		if err != nil {
			t.Error(err)
		}

		if name != "hostname" {
			t.Errorf("expected name to be \"hostname\" got %v", name)
		}

		exitCode, err = item.Attributes.Get("exitCode")

		if err != nil {
			t.Error(err)
		}

		if exitCode != float64(0) {
			t.Errorf("expected exitCode to be 0 got %v", exitCode)
		}

		stdout, err = item.Attributes.Get("stdout")

		if err != nil {
			t.Error(err)
		}

		hostname, err = os.Hostname()

		if err != nil {
			t.Error(err)
		}

		if fmt.Sprint(stdout) != hostname {
			t.Errorf("expected stdout to be %v got %v", hostname, stdout)
		}
	})

	t.Run("with timeout", func(t *testing.T) {
		t.Run("returning before the timeout", func(t *testing.T) {
			params := CommandParams{
				Command: "sleep",
				Args: []string{
					"1",
				},
				Timeout: 2 * time.Second,
			}

			item, err := params.Run()

			if err != nil {
				t.Error(err)
			}

			discovery.TestValidateItem(t, item)
		})

		t.Run("timing out", func(t *testing.T) {
			params := CommandParams{
				Command: "sleep",
				Args: []string{
					"1",
				},
				Timeout: 500 * time.Millisecond,
			}

			_, err := params.Run()

			if err == nil {
				t.Error("No error returned, command should have timed out")
			}
		})
	})

	t.Run("with non-zero exit codes", func(t *testing.T) {
		t.Run("an unexpected non-zero exit should fail", func(t *testing.T) {
			params := CommandParams{
				Command:      "cat",
				Args:         []string{"somethingNotReal"},
				ExpectedExit: 0,
			}

			_, err := params.Run()

			if err == nil {
				t.Error("Expected command to fail but it didn't")
			}
		})

		t.Run("an expected non-zero exit should pass", func(t *testing.T) {
			params := CommandParams{
				Command:      "cat",
				Args:         []string{"somethingNotReal"},
				ExpectedExit: 1,
			}

			_, err := params.Run()

			if err != nil {
				t.Fatal(err)
			}
		})
	})

	t.Run("stdout should work", func(t *testing.T) {
		params := CommandParams{
			Command: "echo",
			Args:    []string{"qwerty"},
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stdout, _ := item.Attributes.Get("stdout"); stdout != "qwerty" {
			t.Errorf("expected stdout to be qwerty, got %v", stdout)
		}
	})

	t.Run("stderr should work", func(t *testing.T) {
		params := CommandParams{
			Command: "perl",
			Args: []string{
				"-e",
				"print STDERR qwerty",
			},
		}

		item, err := params.Run()

		if err != nil {
			t.Fatal(err)
		}

		if stderr, _ := item.Attributes.Get("stderr"); stderr != "qwerty" {
			t.Errorf("expected stderr to be qwerty, got \"%v\"", stderr)
		}
	})
}

const jsonString = `{
	"timeout": "500ms",
	"command": "cat",
	"args": [
		"hosts"
	],
	"expected_exit": 0,
	"dir": "/etc",
	"env": {
		"TEST": "foo"
	}
}`

var jsonObject = CommandParams{
	Command:      "cat",
	Args:         []string{"hosts"},
	ExpectedExit: 0,
	Timeout:      500 * time.Millisecond,
	Dir:          "/etc",
	Env: map[string]string{
		"TEST": "foo",
	},
}

func TestMarshalJSON(t *testing.T) {
	var b []byte
	var err error
	var resultString string

	b, err = json.MarshalIndent(jsonObject, "", "	")

	if err != nil {
		t.Fatal(err)
	}

	resultString = string(b)

	if resultString != jsonString {
		t.Errorf("JSON did not match. Expected:\n%v\nGot:\n%v", jsonString, resultString)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	var cp CommandParams

	err := json.Unmarshal([]byte(jsonString), &cp)

	if err != nil {
		t.Fatal(err)
	}

	if cp.Command != jsonObject.Command {
		t.Errorf("Command did not match, got %v, expected %v", cp.Command, jsonObject.Command)
	}

	if cp.Dir != jsonObject.Dir {
		t.Errorf("Dir did not match, got %v, expected %v", cp.Dir, jsonObject.Dir)
	}

	if cp.ExpectedExit != jsonObject.ExpectedExit {
		t.Errorf("ExpectedExit did not match, got %v, expected %v", cp.ExpectedExit, jsonObject.ExpectedExit)
	}

	if cp.Args[0] != jsonObject.Args[0] {
		t.Errorf("Args[0] did not match, got %v, expected %v", cp.Args[0], jsonObject.Args[0])
	}

	if cp.Timeout.String() != jsonObject.Timeout.String() {
		t.Errorf("Timeout.String() did not match, got %v, expected %v", cp.Timeout.String(), jsonObject.Timeout.String())
	}

	if cp.Env["TEST"] != jsonObject.Env["TEST"] {
		t.Errorf("Env[\"TEST\"] did not match, got %v, expected %v", cp.Env["TEST"], jsonObject.Env["TEST"])
	}
}
