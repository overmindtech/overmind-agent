//go:build linux || windows
// +build linux windows

package netstat

import (
	"context"
	"net"
	"testing"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

func TestPortGet(t *testing.T) {
	var testPort string
	var err error

	// Listen for incoming connections on a random port
	l, err := net.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		t.Errorf("Error listening: %v", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	_, testPort, _ = net.SplitHostPort(l.Addr().String())

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

func TestPortGetSocketLinks(t *testing.T) {
	var testPort string
	var err error

	// Listen for incoming connections on a random port
	l, err := net.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		t.Errorf("Error listening: %v", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	_, testPort, _ = net.SplitHostPort(l.Addr().String())

	src := PortSource{}

	item, err := src.Get(context.Background(), util.LocalContext, testPort)

	if err != nil {
		t.Fatal(err)
	}

	if len(item.LinkedItemRequests) < 2 {
		t.Error("expected at least 2 linked item requests")
	}

	// Make sure that the IP has been expanded.
	for _, lir := range item.LinkedItemRequests {
		if lir.Type == "networksocket" {
			host, _, err := net.SplitHostPort(lir.Query)

			if err != nil {
				t.Error(err)
			}

			if ip := net.ParseIP(host); ip != nil {
				if ip.IsUnspecified() {
					t.Errorf("found unspecififed IP in linked items: %v", ip.String())
				}
			}
		}
	}
}

func TestPortFind(t *testing.T) {
	var testPort string
	var err error

	// Listen for incoming connections on a random port
	l, err := net.Listen("tcp", "0.0.0.0:0")

	if err != nil {
		t.Errorf("Error listening: %v", err.Error())
	}

	// Close the listener when the application closes.
	defer l.Close()

	_, testPort, _ = net.SplitHostPort(l.Addr().String())

	src := PortSource{}

	items, err := src.Find(context.Background(), util.LocalContext)

	if err != nil {
		t.Error(err)
	}

	var found bool

	for _, item := range items {
		if item.UniqueAttributeValue() == testPort {
			found = true
		}
	}

	if !found {
		t.Errorf("couldn't find port %v", testPort)
	}
}
