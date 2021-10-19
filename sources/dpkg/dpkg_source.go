package dpkg

import (
	"fmt"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

// DpkgSource struct on which all methods are registered
type DpkgSource struct{}

// Type is the type of items that this returns (Required)
func (s *DpkgSource) Type() string {
	return "package"
}

// Descriptive name for the source, used in logging and metadata
func (s *DpkgSource) Name() string {
	return "dpkg"
}

// Weighting of duplicate sources
func (s *DpkgSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *DpkgSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *DpkgSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var p Package
	var err error

	p, err = Show(query)

	if e, ok := err.(NotFoundError); ok {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: e.Error(),
		}
	}

	if err != nil {
		return nil, err
	}

	return mapPackageToItem(p)
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *DpkgSource) Find(itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var ps []Package
	var err error
	var item *sdp.Item
	var items []*sdp.Item

	ps, err = ShowAll()

	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		item, err = mapPackageToItem(p)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Search takes a file name or binary name and runs this through rpm -q
// --whatprovides
func (s *DpkgSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var ps []Package
	var err error
	var item *sdp.Item
	var items []*sdp.Item

	ps, err = Search(query)

	if err != nil {
		return nil, err
	}

	for _, p := range ps {
		item, err = mapPackageToItem(p)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

func mapPackageToItem(p Package) (*sdp.Item, error) {
	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"name":         p.Name,
		"status":       p.Status,
		"priority":     p.Priority,
		"section":      p.Section,
		"size":         p.InstalledSize,
		"maintainer":   p.Maintainer,
		"architecture": p.Architecture,
		"version":      p.Version,
		"url":          p.Homepage,
		"summary":      p.Summary,
		"description":  p.Description,
	})

	if err != nil {
		return nil, err
	}

	return &sdp.Item{
		Type:            "package",
		UniqueAttribute: "name",
		Attributes:      attributes,
		Context:         util.LocalContext,
	}, nil
}
