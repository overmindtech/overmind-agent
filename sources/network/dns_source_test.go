package network

import (
	"net"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

func TestDnsGet(t *testing.T) {
	var conn net.Conn
	var err error

	// Check that we actually have an inertnet connection, if not there is not
	// point running this test
	conn, err = net.Dial("tcp", "one.one.one.one:443")

	if err != nil {
		t.Skip("No internet connection detected")
	}

	conn.Close()

	tests := []util.SourceTest{
		{
			Name:          "one.one.one.one",
			ItemContext:   "global",
			Query:         "one.one.one.one",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "bad dns entry",
			ItemContext: "global",
			Query:       "something.does.not.exist.please.testing",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOTFOUND,
				Context: "global",
			},
		},
		{
			Name:        "bad context",
			ItemContext: "something.local",
			Query:       "one.one.one.one",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOCONTEXT,
				Context: "something.local",
			},
		},
	}

	util.RunSourceTests(t, tests, &DNSSource{})
}

func TestDnsFind(t *testing.T) {
	tests := []util.SourceTest{
		{
			Name:          "find all",
			ItemContext:   "global",
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 0,
			},
		},
	}

	util.RunSourceTests(t, tests, &DNSSource{})
}
