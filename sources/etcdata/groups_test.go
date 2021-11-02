package etcdata

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

func TestDefaultGroupsFileExists(t *testing.T) {
	if runtime.GOOS != "windows" {
		// Test that the default hosts file does actually exist
		if _, err := os.Stat(DefaultGroupsLocation); os.IsNotExist(err) {
			t.Errorf("Default Group file %v does not exist", DefaultHostsLocation)
		}
	}

}

func TestGroupsFind(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/group")

	s := &GroupsSource{
		GroupsLocation: exampleFile,
	}

	tests := []util.SourceTest{
		{
			Name:          "running find",
			ItemContext:   util.LocalContext,
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 5,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "0",
						"name": "wheel",
						"members": []interface{}{
							"root",
						},
					},
					{
						"gid":  "1",
						"name": "daemon",
					},
					{
						"gid":  "6",
						"name": "mail",
						"members": []interface{}{
							"_teamsserver",
						},
					},
					{
						"gid":  "13",
						"name": "_taskgated",
						"members": []interface{}{
							"_taskgated",
						},
					},
					{
						"gid":  "29",
						"name": "certusers",
						"members": []interface{}{
							"root",
							"_jabber",
							"_postfix",
							"_cyrus",
							"_calendar",
							"_dovecot",
						},
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, s)
}

func TestGroupsSearch(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/group")

	s := &GroupsSource{
		GroupsLocation: exampleFile,
	}

	tests := []util.SourceTest{
		{
			Name:          "search wheel",
			ItemContext:   util.LocalContext,
			Query:         "wheel",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "0",
						"name": "wheel",
						"members": []interface{}{
							"root",
						},
					},
				},
			},
		},
		{
			Name:          "search 1",
			ItemContext:   util.LocalContext,
			Query:         "1",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "1",
						"name": "daemon",
					},
				},
			},
		},
		{
			Name:          "search 29",
			ItemContext:   util.LocalContext,
			Query:         "29",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "29",
						"name": "certusers",
						"members": []interface{}{
							"root",
							"_jabber",
							"_postfix",
							"_cyrus",
							"_calendar",
							"_dovecot",
						},
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, s)
}

func TestGroupsGet(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/group")

	s := &GroupsSource{
		GroupsLocation: exampleFile,
	}

	tests := []util.SourceTest{
		{
			Name:          "get wheel",
			ItemContext:   util.LocalContext,
			Query:         "wheel",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "0",
						"name": "wheel",
						"members": []interface{}{
							"root",
						},
					},
				},
			},
		},
		{
			Name:          "get daemon",
			ItemContext:   util.LocalContext,
			Query:         "daemon",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "1",
						"name": "daemon",
					},
				},
			},
		},
		{
			Name:          "get mail",
			ItemContext:   util.LocalContext,
			Query:         "mail",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "6",
						"name": "mail",
						"members": []interface{}{
							"_teamsserver",
						},
					},
				},
			},
		},
		{
			Name:          "get _taskgated",
			ItemContext:   util.LocalContext,
			Query:         "_taskgated",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "13",
						"name": "_taskgated",
						"members": []interface{}{
							"_taskgated",
						},
					},
				},
			},
		},
		{
			Name:          "get certusers",
			ItemContext:   util.LocalContext,
			Query:         "certusers",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"gid":  "29",
						"name": "certusers",
						"members": []interface{}{
							"root",
							"_jabber",
							"_postfix",
							"_cyrus",
							"_calendar",
							"_dovecot",
						},
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, s)
}
