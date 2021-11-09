package command

import (
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
			t.Errorf("expected stdout to be 0 got %v", stdout)
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
}
