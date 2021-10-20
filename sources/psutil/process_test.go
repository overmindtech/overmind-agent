package psutil

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

func TestProcessGet(t *testing.T) {
	var childProcess *sdp.Item
	var childPID float64
	var pi interface{}
	var err error

	source := ProcessSource{}

	childPID = float64(os.Getpid())

	// Get the current process
	childProcess, err = source.Get(util.LocalContext, fmt.Sprint(childPID))

	if err != nil {
		t.Error(err)
	}

	// Get parent
	pi, err = childProcess.GetAttributes().Get("parent")

	if err != nil {
		t.Error(err)
	}

	tests := []util.SourceTest{
		{
			Name:        "get existing pid",
			ItemContext: util.LocalContext,
			Query:       fmt.Sprint(childPID),
			Method:      sdp.RequestMethod_GET,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "get parent pid",
			ItemContext: util.LocalContext,
			Query:       fmt.Sprint(pi),
			Method:      sdp.RequestMethod_GET,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "bad context",
			ItemContext: "bad",
			Query:       fmt.Sprint(pi),
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
		{
			Name:        "bad pid format",
			ItemContext: util.LocalContext,
			Query:       "stringPid",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:             sdp.ItemRequestError_OTHER,
				ErrorStringRegex: regexp.MustCompile("could not be converted to integer"),
			},
		},
		{
			Name:        "non existent pid",
			ItemContext: util.LocalContext,
			Query:       "891726304",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOTFOUND,
				Context: util.LocalContext,
			},
		},
	}

	util.RunSourceTests(t, tests, &source)

}
