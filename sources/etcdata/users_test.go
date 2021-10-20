package etcdata

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

// ExpectedUsers This iis the array of hosts that is expected to be found in the
// test/hosts file. Note that this needs to be in the same order as the
// hosts file
var ExpectedUsers = []map[string]interface{}{
	{
		"uid":      "65535",
		"gid":      "65535",
		"username": "nobody",
		"comment":  "Unprivileged User",
		"home":     "/var/empty",
	},
	{
		"uid":      "0",
		"gid":      "0",
		"username": "root",
		"comment":  "System Administrator",
		"home":     "/var/root",
	},
	{
		"uid":      "1",
		"gid":      "1",
		"username": "daemon",
		"comment":  "System Services",
		"home":     "/var/root",
	},
	{
		"uid":      "4",
		"gid":      "4",
		"username": "_uucp",
		"comment":  "Unix to Unix Copy Protocol",
		"home":     "/var/spool/uucp",
	},
}

func TestDefaultPasswdFileExists(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Test that the default hosts file does actually exist
		if _, err := os.Stat(DefaultPasswdLocation); os.IsNotExist(err) {
			t.Errorf("Default passwd file %v does not exist", DefaultHostsLocation)
		}
	}

}

func TestUsersFind(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/passwd")

	s := &UsersSource{
		PasswdLocation: exampleFile,
	}

	tests := []util.SourceTest{
		{
			Name:          "normal find",
			ItemContext:   util.LocalContext,
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems:           4,
				ExpectedAttributes: ExpectedUsers,
			},
		},
	}

	util.RunSourceTests(t, tests, s)
}

func TestUsersSearch(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/passwd")

	s := &UsersSource{
		PasswdLocation: exampleFile,
	}

	tests := []util.SourceTest{
		{
			Name:          "root search",
			ItemContext:   util.LocalContext,
			Query:         "root",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"uid":      "0",
						"gid":      "0",
						"username": "root",
						"comment":  "System Administrator",
						"home":     "/var/root",
					},
				},
			},
		},
		{
			Name:          "1 search",
			ItemContext:   util.LocalContext,
			Query:         "1",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"uid":      "1",
						"gid":      "1",
						"username": "daemon",
						"comment":  "System Services",
						"home":     "/var/root",
					},
				},
			},
		},
		{
			Name:          "65535 search",
			ItemContext:   util.LocalContext,
			Query:         "65535",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"uid":      "65535",
						"gid":      "65535",
						"username": "nobody",
						"comment":  "Unprivileged User",
						"home":     "/var/empty",
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, s)
}

func TestUsersGet(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/passwd")

	s := &UsersSource{
		PasswdLocation: exampleFile,
	}

	tests := make([]util.SourceTest, 0)

	for _, expected := range ExpectedUsers {
		name := fmt.Sprint(expected["username"])
		tests = append(tests, util.SourceTest{
			Name:          name,
			ItemContext:   util.LocalContext,
			Query:         name,
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					expected,
				},
			},
		})
	}

	util.RunSourceTests(t, tests, s)
}
