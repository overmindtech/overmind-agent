package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/overmind-agent/sources/util"
)

func TestGet(t *testing.T) {
	if _, err := exec.LookPath("cat"); err == nil {
		t.Run("cat a file", func(t *testing.T) {
			t.Parallel()

			if _, err := os.Stat("/etc/hosts"); err == nil {
				s := CommandSource{}

				item, err := s.Get(context.Background(), util.LocalContext, "cat /etc/hosts")

				if err != nil {
					t.Fatal(err)
				}

				discovery.TestValidateItem(t, item)
			} else {
				t.Skip("/etc/hosts does not exist on this platform, skipping")
			}
		})

		t.Run("command that fails", func(t *testing.T) {
			t.Parallel()

			s := CommandSource{}

			_, err := s.Get(context.Background(), util.LocalContext, "cat something doesn't exist")

			if err == nil {
				t.Fatal("expected error but got <nil>")
			}
		})
	} else {
		t.Skip("Skiping test as cat binary not present")
	}
}

func TestGetPowershell(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Run("Running built in powershell cmdlets", func(t *testing.T) {
			s := CommandSource{}

			item, err := s.Get(context.Background(), util.LocalContext, "Write-Host qwerty")

			if err != nil {
				t.Fatal(err)
			}

			var stdout interface{}

			stdout, err = item.Attributes.Get("stdout")

			if err != nil {
				t.Fatal(err)
			}

			if stdout != "qwerty\n" {
				t.Errorf("Expected stdout to be 'qwerty', got %v", stdout)
			}
		})

		t.Run("Running multiline powershell scripts", func(t *testing.T) {
			s := CommandSource{}

			script := `
			Get-CimInstance -ClassName Win32_LocalTime
			Get-CimInstance -ClassName Win32_ComputerSystem
			`

			item, err := s.Get(context.Background(), util.LocalContext, script)

			if err != nil {
				t.Fatal(err)
			}

			var stdout interface{}

			stdout, err = item.Attributes.Get("stdout")

			if err != nil {
				t.Fatal(err)
			}

			computerSystemRegex := regexp.MustCompile("TotalPhysicalMemory")
			localTimeRegex := regexp.MustCompile("WeekInMonth")
			stdoutString := fmt.Sprint(stdout)

			if !(computerSystemRegex.MatchString(stdoutString) && localTimeRegex.MatchString(stdoutString)) {
				t.Errorf("Unexpected result:\r\n%v", stdout)
			}
		})
	} else {
		t.Skip("Powershell tests only run on windows")
	}
}

func TestSearch(t *testing.T) {
	s := CommandSource{}

	t.Run("with a valid command", func(t *testing.T) {
		t.Parallel()

		if _, err := os.Stat("/etc/hosts"); err == nil {
			query := `{
				"command": "cat hosts",
				"expected_exit": 0,
				"dir": "/etc",
				"env": {
					"FOO": "BAR"
				}
			}`

			items, err := s.Search(context.Background(), util.LocalContext, query)

			if err != nil {
				t.Fatal(err)
			}

			discovery.TestValidateItems(t, items)
		} else {
			t.Skip("skipping because /etc/hosts does not exist")
		}
	})

	t.Run("with invalid json", func(t *testing.T) {
		t.Parallel()

		query := `{
			"command": "cat",
			"args": ["hosts"]
			"expected_exit": 0
			"dir": "/etc"
			"env": {
				"FOO": "BAR"
			}
		}`

		_, err := s.Search(context.Background(), util.LocalContext, query)

		if err == nil {
			t.Fatal("expected error but got <nil>")
		}

		expectedError := regexp.MustCompile("could not unmarshal JSON query")

		if !expectedError.MatchString(err.Error()) {
			t.Fatalf("expected error to match %v, got %v", expectedError, err)
		}
	})

	t.Run("with invalid data types", func(t *testing.T) {
		t.Parallel()

		query := `{
			"command": "cat",
			"args": ["hosts"],
			"expected_exit": 0,
			"dir": "/etc",
			"env": {
				"FOO": 12
			}
		}`

		_, err := s.Search(context.Background(), util.LocalContext, query)

		if err == nil {
			t.Fatal("expected error but got <nil>")
		}

		expectedError := regexp.MustCompile("could not unmarshal JSON query")

		if !expectedError.MatchString(err.Error()) {
			t.Fatalf("expected error to match %v, got %v", expectedError, err)
		}
	})

	t.Run("with a timeout", func(t *testing.T) {
		t.Parallel()

		if _, err := exec.LookPath("sleep"); err == nil {
			query := `{
				"command": "sleep 60",
				"expected_exit": 0
			}`

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			_, err := s.Search(ctx, util.LocalContext, query)

			if err == nil {
				t.Fatal("expected error but got <nil>")
			}

			expectedError := regexp.MustCompile("command execution timed out")

			if !expectedError.MatchString(err.Error()) {
				t.Fatalf("expected error to match %v, got %v", expectedError, err)
			}
		} else {
			t.Skip("skipping because sleep binary cannot be found")
		}
	})
}

func TestFind(t *testing.T) {
	t.Parallel()

	s := CommandSource{}

	_, err := s.Find(context.Background(), util.LocalContext)

	if err == nil {
		t.Fatal("expected error but got <nil>")
	}

	expectedError := regexp.MustCompile("the command source only supports the Get and Search methods")

	if !expectedError.MatchString(err.Error()) {
		t.Errorf("expected error to match %v, got %v", expectedError, err)
	}

}

func TestHidden(t *testing.T) {
	t.Parallel()

	s := CommandSource{}

	if !s.Hidden() {
		t.Error("Source should be hidden")
	}
}
