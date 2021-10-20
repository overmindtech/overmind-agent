package system

import (
	"regexp"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

func TestFind(t *testing.T) {
	tests := []util.SourceTest{
		{
			Name:        "regular find",
			ItemContext: util.LocalContext,
			Method:      sdp.RequestMethod_FIND,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					{
						"hostname": util.LocalContext,
					},
				},
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

	util.RunSourceTests(t, tests, &SystemSource{})
}

func TestGet(t *testing.T) {
	tests := []util.SourceTest{
		{
			Name:        "regular get",
			ItemContext: util.LocalContext,
			Method:      sdp.RequestMethod_GET,
			Query:       util.LocalContext,
			ExpectedError: &util.ExpectedError{
				Type:             sdp.ItemRequestError_NOTFOUND,
				ErrorStringRegex: regexp.MustCompile("backend does not respond to Get"),
			},
		},
	}

	util.RunSourceTests(t, tests, &SystemSource{})
}
