package file_content

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

const DefaultMaxSize = (100 * 1024 * 1024)

// FileContentSource struct on which all methods are registered
type FileContentSource struct {
	// Maximum file size in bytes to be read
	MaxSize int64
}

// RealMaxSize returns the max size of file that should be read. If the store
// MaxSize is 0, uses the default
func (bc *FileContentSource) realMaxSize() int64 {
	if bc.MaxSize == 0 {
		return DefaultMaxSize
	}

	return bc.MaxSize
}

// Type is the type of items that this returns
func (bc *FileContentSource) Type() string {
	return "filecontent"
}

// Hidden File_content items should be hidden as they are only used when needed,
// and should be covered by some higher layer of abstraction like a secondary
// source
func (bc *FileContentSource) Hidden() bool {
	return true
}

// Name Returns the name of the backend
func (bc *FileContentSource) Name() string {
	return "unix"
}

// Weighting of duplicate sources
func (s *FileContentSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *FileContentSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get calls back to the GetFunction to get a single item
func (bc *FileContentSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var info fs.FileInfo
	var content []byte
	var err error
	var item sdp.Item

	// Check that we can see the file
	info, err = os.Stat(query)

	if err != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	// Check the file size
	if info.Size() > bc.realMaxSize() {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("File size %v exceeded maximum %v ", byteCountIEC(info.Size()), byteCountIEC(bc.MaxSize)),
			Context:     util.LocalContext,
		}
	}

	// Read the file into memory
	content, err = os.ReadFile(query)

	if err != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: err.Error(),
			Context:     util.LocalContext,
		}
	}

	// Encode with base64.
	//
	// This is required since although protocol buffers are
	// capable of transporting raw binary, the attributes of an item are stored
	// in a google.protobuf.Struct. This is capable of holding the same data
	// types that JSON is, and therefore is *not* capable of handling raw
	// binary, meaning that we need an encoding.
	//
	// We could have chosen any string encoding here, including something with
	// great efficiency since we aren't limited to a certain character set in
	// protocol buffers (such as yEnc), but I've chosen base64 for now since
	// it's very common and the encoding overhead isn't ridiculous
	base64Content := base64.StdEncoding.EncodeToString(content)

	item = sdp.Item{
		Type:            "filecontent",
		UniqueAttribute: "path",
		Context:         util.LocalContext,
	}

	// Reset err
	err = nil

	item.Attributes, err = sdp.ToAttributes(map[string]interface{}{
		"path":           query,
		"size":           info.Size(),
		"content_base64": base64Content,
	})

	return &item, err
}

// Find calls back to the FindFunction to find all items
func (bc *FileContentSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	return make([]*sdp.Item, 0), nil
}

func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
