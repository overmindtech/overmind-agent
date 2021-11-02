package rpm

import (
	"fmt"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

// RPMSource struct on which all methods are registered
type RPMSource struct{}

// Type is the type of items that this returns (Required)
func (bc *RPMSource) Type() string {
	return "package"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (bc *RPMSource) Name() string {
	return "rpm"
}

// Weighting of duplicate sources
func (s *RPMSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *RPMSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (bc *RPMSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var p Package
	var err error

	p, err = Query(query)

	if e, ok := err.(NotFoundError); ok {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: e.Error(),
			Context:     itemContext,
		}
	}

	if err != nil {
		return nil, err
	}

	return mapPackageToItem(p)
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (bc *RPMSource) Find(itemContext string) ([]*sdp.Item, error) {
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

	ps, err = QueryAll()

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
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
func (bc *RPMSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
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

	ps, err = WhatProvides(query)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	for _, p := range ps {
		item, err = mapPackageToItem(p)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (bc *RPMSource) Supported() bool {
	return Supported()
}

func mapPackageToItem(p Package) (*sdp.Item, error) {
	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"name":         p.Name,
		"epoch":        p.Epoch,
		"version":      p.Version,
		"release":      p.Release,
		"architecture": p.Arch,
		"installtime":  p.InstallTime,
		"group":        p.Group,
		"size":         p.Size,
		"license":      p.License,
		"sourcerpm":    p.SourceRPM,
		"vendor":       p.Vendor,
		"url":          p.URL,
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
