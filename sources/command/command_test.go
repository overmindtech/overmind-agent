package command

import (
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/overmindtech/discovery"
	"github.com/overmindtech/overmind-agent/sources/util"
)

func TestGet(t *testing.T) {
	if _, err := exec.LookPath("cat"); err == nil {
		t.Run("cat a file", func(t *testing.T) {
			if _, err := os.Stat("/etc/hosts"); err == nil {
				s := CommandSource{}

				item, err := s.Get(util.LocalContext, "cat /etc/hosts")

				if err != nil {
					t.Fatal(err)
				}

				discovery.TestValidateItem(t, item)
			} else {
				t.Skip("/etc/hosts does not exist on this platform, skipping")
			}
		})

		t.Run("command that fails", func(t *testing.T) {
			s := CommandSource{}

			_, err := s.Get(util.LocalContext, "cat something doesn't exist")

			if err == nil {
				t.Fatal("expected error but got <nil>")
			}
		})
	} else {
		t.Skip("Skiping test as cat binary not present")
	}
}

func TestSearch(t *testing.T) {
	s := CommandSource{}

	t.Run("with a valid command", func(t *testing.T) {
		query := `{
			"command": "cat",
			"args": ["hosts"],
			"expected_exit": 0,
			"timeout": "5s",
			"dir": "/etc",
			"env": {
				"FOO": "BAR"
			}
		}`

		items, err := s.Search(util.LocalContext, query)

		if err != nil {
			t.Fatal(err)
		}

		discovery.TestValidateItems(t, items)
	})

	t.Run("with invalid json", func(t *testing.T) {
		query := `{
			"command": "cat",
			"args": ["hosts"]
			"expected_exit": 0
			"timeout": "5s"
			"dir": "/etc"
			"env": {
				"FOO": "BAR"
			}
		}`

		_, err := s.Search(util.LocalContext, query)

		if err == nil {
			t.Fatal("expected error but got <nil>")
		}

		expectedError := regexp.MustCompile("could not unmarshal JSON query")

		if !expectedError.MatchString(err.Error()) {
			t.Fatalf("expected error to match %v, got %v", expectedError, err)
		}
	})

	t.Run("with invalid data types", func(t *testing.T) {
		query := `{
			"command": "cat",
			"args": ["hosts"],
			"expected_exit": 0,
			"timeout": "5s",
			"dir": "/etc",
			"env": {
				"FOO": 12
			}
		}`

		_, err := s.Search(util.LocalContext, query)

		if err == nil {
			t.Fatal("expected error but got <nil>")
		}

		expectedError := regexp.MustCompile("could not unmarshal JSON query")

		if !expectedError.MatchString(err.Error()) {
			t.Fatalf("expected error to match %v, got %v", expectedError, err)
		}
	})

	t.Run("with a timeout", func(t *testing.T) {
		query := `{
			"command": "sleep",
			"args": ["60"],
			"expected_exit": 0,
			"timeout": "100ms"
		}`

		_, err := s.Search(util.LocalContext, query)

		if err == nil {
			t.Fatal("expected error but got <nil>")
		}

		expectedError := regexp.MustCompile("command execution timed out")

		if !expectedError.MatchString(err.Error()) {
			t.Fatalf("expected error to match %v, got %v", expectedError, err)
		}
	})
}

func TestFind(t *testing.T) {
	s := CommandSource{}

	_, err := s.Find(util.LocalContext)

	if err == nil {
		t.Fatal("expected error but got <nil>")
	}

	expectedError := regexp.MustCompile("the command source only supports the Get and Search methods")

	if !expectedError.MatchString(err.Error()) {
		t.Errorf("expected error to match %v, got %v", expectedError, err)
	}

}

func TestHidden(t *testing.T) {
	s := CommandSource{}

	if !s.Hidden() {
		t.Error("Source should be hidden")
	}
}
