package rpm

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

// TestGet Runs rpm -qa manually and ensures that the backend is able to get all
// of the packages that were returned
func TestGet(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	tests := []util.SourceTest{
		{
			Name:        "bad context",
			ItemContext: "bad",
			Query:       "bash",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
		{
			Name:        "bad package name",
			ItemContext: util.LocalContext,
			Query:       "bad package name",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOTFOUND,
				Context: util.LocalContext,
			},
		},
	}

	command := exec.Command("rpm", "-qa")

	// Run the command
	output, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")

	if len(lines) > 0 {
		line := lines[len(lines)-1]

		if line != "" {
			tests = append(tests, util.SourceTest{
				Name:        "get real package",
				ItemContext: util.LocalContext,
				Query:       line,
				Method:      sdp.RequestMethod_GET,
				ExpectedItems: &util.ExpectedItems{
					NumItems: 1,
					ExpectedAttributes: []map[string]interface{}{
						{
							"name": line,
						},
					},
				},
			})
		}
	}

	util.RunSourceTests(t, tests, &RPMSource{})
}

func TestFind(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	if !HasPackages() {
		t.Skip("RPM Based system with no packages, skipping test")
	}

	command := exec.Command("rpm", "-qa")

	// Run the command
	output, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(output), "\n")

	tests := []util.SourceTest{
		{
			Name:        "normal find",
			ItemContext: util.LocalContext,
			Method:      sdp.RequestMethod_FIND,
			ExpectedItems: &util.ExpectedItems{
				NumItems: len(lines) - 1,
			},
		},
		{
			Name:        "bad context",
			ItemContext: "bad",
			Method:      sdp.RequestMethod_FIND,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
	}

	util.RunSourceTests(t, tests, &RPMSource{})
}

func TestSearch(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	if !HasPackages() {
		t.Skip("RPM Based system with no packages, skipping test")
	}

	tests := []util.SourceTest{
		{
			Name:        "/etc/hosts search",
			ItemContext: util.LocalContext,
			Query:       "/etc/hosts",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "bash search",
			ItemContext: util.LocalContext,
			Query:       "bash",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "bad context search",
			ItemContext: "bad",
			Query:       "bash",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
	}

	util.RunSourceTests(t, tests, &RPMSource{})
}
