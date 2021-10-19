package dpkg

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources"
	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

// TestGet Runs rpm -qa manually and ensures that the backend is able to get all
// of the packages that were returned
func TestGetFind(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	command := exec.Command("dpkg-query", "-f=${Package}\n", "--show")

	// Run the command
	output, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")

	if len(lines) > 0 {
		line := lines[len(lines)-1]

		if line != "" {
			tests := []sources.SourceTest{
				{
					Name:          "Getting last package",
					ItemContext:   util.LocalContext,
					Query:         line,
					Method:        sdp.RequestMethod_GET,
					ExpectedError: nil,
					ExpectedItems: &sources.ExpectedItems{
						NumItems: 1,
					},
				},
				{
					Name:          "Find all packages",
					ItemContext:   util.LocalContext,
					Method:        sdp.RequestMethod_FIND,
					ExpectedError: nil,
					ExpectedItems: &sources.ExpectedItems{
						NumItems: len(lines),
					},
				},
			}

			sources.RunSourceTests(t, tests, &DpkgSource{})
		}
	}
}

func TestBackendSearch(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	if !HasPackages() {
		t.Skip("dpkg-query has no packages, skipping test")
	}

	tests := []sources.SourceTest{
		{
			Name:        "bash",
			ItemContext: util.LocalContext,
			Query:       "/bin/bash",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
			},
		},
	}

	sources.RunSourceTests(t, tests, &DpkgSource{})
}
