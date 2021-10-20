//go:build linux || darwin
// +build linux darwin

package unix

import (
	"io/ioutil"
	"os"
	"os/user"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

func TestFileFind(t *testing.T) {
	tests := []util.SourceTest{
		{
			Name:        "normal find",
			ItemContext: util.LocalContext,
			Method:      sdp.RequestMethod_FIND,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 0,
			},
		},
		{
			Name:        "find with bad context",
			ItemContext: "bad",
			Method:      sdp.RequestMethod_FIND,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
	}

	util.RunSourceTests(t, tests, &FileSource{})
}

func TestFileGet(t *testing.T) {
	var file *os.File
	var err error

	// Create a file for testing on
	file, err = ioutil.TempFile("", "tempTestFile")

	if err != nil {
		t.Error(err)
	}

	t.Cleanup(func() {
		os.Remove(file.Name())
	})

	user, err := user.Current()

	if err != nil {
		t.Fatal(err)
	}

	tests := []util.SourceTest{
		{
			Name:        "get bad context",
			ItemContext: "bad",
			Query:       file.Name(),
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
		{
			Name:        "get bad file path",
			ItemContext: util.LocalContext,
			Query:       "/tmp/please-no-exist-1723632",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOTFOUND,
				Context: util.LocalContext,
			},
		},
		{
			Name:        "get real file",
			ItemContext: util.LocalContext,
			Query:       file.Name(),
			Method:      sdp.RequestMethod_GET,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"path":  file.Name(),
						"type":  "file",
						"owner": user.Username,
						"size":  float64(0),
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, &FileSource{})
}
