//go:build linux || windows
// +build linux windows

package netstat

import (
	"net"
	"strings"
	"testing"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

func TestPortGet(t *testing.T) {
	var testPort string
	var err error

	// Listen for incoming connections on a random port
	l, err := net.Listen("tcp", "localhost:0")

	if err != nil {
		t.Errorf("Error listening: %v", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	testPort = strings.Split(l.Addr().String(), ":")[1]

	tests := []util.SourceTest{
		{
			Name:          "get listening port",
			ItemContext:   util.LocalContext,
			Query:         testPort,
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"state": "LISTEN",
					},
				},
			},
		},
	}

	util.RunSourceTests(t, tests, &PortSource{})
}
