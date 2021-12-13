//go:build linux
// +build linux

package systemd

import (
	"testing"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

func TestServiceFind(t *testing.T) {
	source := ServiceSource{}

	if !source.Supported() {
		t.Skip("DBUS Not supported")
	}

	tests := []util.SourceTest{
		{
			Name:        "normal find",
			ItemContext: util.LocalContext,
			Method:      sdp.RequestMethod_FIND,
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

	util.RunSourceTests(t, tests, &source)
}

func TestServiceGet(t *testing.T) {
	services := []string{
		"dbus.service",
		"ssh.service",
	}

	source := ServiceSource{}

	if !source.Supported() {
		t.Skip("DBUS Not supported, cannot get services")
	}

	tests := []util.SourceTest{
		{
			Name:        "get with bad context",
			ItemContext: "bad",
			Query:       "dbus.service",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
		{
			Name:        "get with bad service name",
			ItemContext: util.LocalContext,
			Query:       "nothing.service.nope",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOTFOUND,
			},
		},
	}

	for _, serviceName := range services {
		tests = append(tests, util.SourceTest{
			Name:        serviceName,
			ItemContext: util.LocalContext,
			Query:       serviceName,
			Method:      sdp.RequestMethod_GET,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"Name": serviceName,
					},
				},
			},
		})
	}

	util.RunSourceTests(t, tests, &source)
}

func TestServiceSearch(t *testing.T) {
	services := []string{
		"dbus.service",
		"ssh.service",
		"dbus",
		"ssh",
		"1",
	}

	source := ServiceSource{}

	if !source.Supported() {
		t.Skip("DBUS Not supported, cannot get services")
	}

	tests := []util.SourceTest{
		{
			Name:        "search with bad context",
			ItemContext: "bad",
			Query:       "dbus.service",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedError: &util.ExpectedError{
				Type: sdp.ItemRequestError_NOCONTEXT,
			},
		},
		{
			Name:        "search with bad service name",
			ItemContext: util.LocalContext,
			Query:       "nothing.service.nope",
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 0,
			},
		},
	}

	for _, serviceName := range services {
		tests = append(tests, util.SourceTest{
			Name:        serviceName,
			ItemContext: util.LocalContext,
			Query:       serviceName,
			Method:      sdp.RequestMethod_SEARCH,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		})
	}

	util.RunSourceTests(t, tests, &source)
}
