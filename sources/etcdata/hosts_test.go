package etcdata

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources"
	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

func TestDefaultHostsFileExists(t *testing.T) {
	// Test that the default hosts file does actually exist
	if _, err := os.Stat(DefaultHostsLocation); os.IsNotExist(err) {
		t.Errorf("Default hosts file %v does not exist", DefaultHostsLocation)
	}
}

func TestHostsFind(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/hosts")

	s := &HostsSource{
		HostsLocation: exampleFile,
	}

	tests := []sources.SourceTest{
		{
			Name:          "Find",
			ItemContext:   util.LocalContext,
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 10,
				ExpectedAttributes: []map[string]interface{}{
					{
						"address": "127.0.0.1",
						"name":    "localhost",
					},
					{
						"address": "127.0.1.1",
						"name":    "thishost.mydomain.org",
						"aliases": []interface{}{
							"thishost",
						},
					},
					{
						"address": "192.168.1.10",
						"name":    "foo.mydomain.org",
						"aliases": []interface{}{
							"foo",
						},
					},
					{
						"address": "192.168.1.13",
						"name":    "bar.mydomain.org",
						"aliases": []interface{}{
							"bar",
						},
					},
					{
						"address": "146.82.138.7",
						"name":    "master.debian.org",
						"aliases": []interface{}{
							"master",
						},
					},
					{
						"address": "209.237.226.90",
						"name":    "www.opensource.org",
					},
					{
						"address": "::1",
						"name":    "localhost6",
						"aliases": []interface{}{
							"ip6-localhost",
							"ip6-loopback",
						},
					},
					{
						"address": "ff02::1",
						"name":    "ip6-allnodes",
					},
					{
						"address": "ff02::2",
						"name":    "ip6-allrouters",
					},
					{
						"address": "1.1.1.1",
						"name":    "one.one.one.one",
						"comment": "Testing a comment",
					},
				},
			},
		},
	}

	sources.RunSourceTests(t, tests, s)
}

func TestHostsSearch(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/hosts")

	s := &HostsSource{
		HostsLocation: exampleFile,
	}

	tests := []sources.SourceTest{
		{
			Name:          "localhost",
			ItemContext:   util.LocalContext,
			Query:         "localhost",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"address": "127.0.0.1",
						"name":    "localhost",
					},
				},
			},
		},
		{
			Name:          "1.1.1.1",
			ItemContext:   util.LocalContext,
			Query:         "1.1.1.1",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"address": "1.1.1.1",
						"name":    "one.one.one.one",
						"comment": "Testing a comment",
					},
				},
			},
		},
	}

	sources.RunSourceTests(t, tests, s)
}

func TestHostsGet(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/hosts")

	s := &HostsSource{
		HostsLocation: exampleFile,
	}

	tests := []sources.SourceTest{
		{
			Name:          "localhost",
			ItemContext:   util.LocalContext,
			Query:         "localhost",
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"address": "127.0.0.1",
						"name":    "localhost",
					},
				},
			},
		},
		{
			Name:          "1.1.1.1",
			ItemContext:   util.LocalContext,
			Query:         "1.1.1.1",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"address": "1.1.1.1",
						"name":    "one.one.one.one",
						"comment": "Testing a comment",
					},
				},
			},
		},
	}

	sources.RunSourceTests(t, tests, s)
}
