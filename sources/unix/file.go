//go:build linux || darwin
// +build linux darwin

package unix

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

// FileSource struct on which all methods are registered
type FileSource struct{}

// Type is the type of items that this returns
func (bc *FileSource) Type() string {
	return "file"
}

// Name Returns the name of the backend
func (bc *FileSource) Name() string {
	return "unix"
}

// Weighting of duplicate sources
func (s *FileSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *FileSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get calls back to the GetFunction to get a single item
func (bc *FileSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var file *os.File
	var typ string
	var info os.FileInfo
	var stat *syscall.Stat_t
	var uid uint32
	var owner string
	var gid uint32
	var group string
	var size uint64
	var mode os.FileMode
	var modeString string
	var err error
	var item sdp.Item

	file, err = os.Open(query)

	if err != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	info, err = file.Stat()

	if err != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	stat = info.Sys().(*syscall.Stat_t)

	// Pull all the info that we can from stat
	uid = uint32(stat.Uid)
	gid = uint32(stat.Gid)
	size = uint64(info.Size())
	mode = info.Mode()

	// Conversions

	// Convert UID to username
	u, err := user.LookupId(strconv.Itoa(int(uid)))

	if err != nil {
		owner = strconv.Itoa(int(uid))
	} else {
		owner = u.Username
	}

	// Convert GID to groupname
	g, err := user.LookupGroupId(strconv.Itoa(int(gid)))

	if err != nil {
		group = strconv.Itoa(int(gid))
	} else {
		group = g.Name
	}

	// Convert mode to string
	modeString = mode.String()

	// Get the type
	if mode.IsDir() {
		typ = "directory"
	} else {
		typ = "file"
	}

	item = sdp.Item{}
	item.Type = "file"
	item.UniqueAttribute = "path"
	item.Context = util.LocalContext

	// Reset err
	err = nil

	item.Attributes, err = sdp.ToAttributes(map[string]interface{}{
		"type":  typ,
		"owner": owner,
		"group": group,
		"mode":  modeString,
		"size":  size,
		"path":  query,
	})

	item.LinkedItemRequests = []*sdp.ItemRequest{
		{
			Type:    "group",
			Query:   group,
			Method:  sdp.RequestMethod_GET,
			Context: util.LocalContext,
		},
		{
			Type:    "user",
			Query:   owner,
			Method:  sdp.RequestMethod_GET,
			Context: util.LocalContext,
		},
		{
			Type:    "package",
			Method:  sdp.RequestMethod_SEARCH,
			Query:   query,
			Context: util.LocalContext,
		},
	}

	return &item, err
}

// Find calls back to the FindFunction to find all items
func (bc *FileSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	return make([]*sdp.Item, 0), nil
}
