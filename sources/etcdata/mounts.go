package etcdata

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/overmind-agent/sources/util/parser"
	"github.com/overmindtech/sdp-go"
)

// MountsSource struct on which all methods are registered
type MountsSource struct {
	MountFunction func() (*bufio.Scanner, error)
}

// Scanner Executes the MountFunction and returns the scanner
func (s *MountsSource) Scanner(ctx context.Context) (*bufio.Scanner, error) {
	if s.MountFunction == nil {
		return mountCMD(ctx)
	}

	return s.MountFunction()
}

// MountParser A Parser for the /etc/passwd file
var MountParser = parser.Parser{
	Ignore:  regexp.MustCompile("^#"),
	Capture: regexp.MustCompile(MountRegex),
}

// MountTimeout The timeout before the `mount` command should be canceled
const MountTimeout = (2 * time.Second)

// MountRegex Regex used to parse the output of the `mount` command
const MountRegex = `^(?P<device>.+) on (?P<path>\S+) ((type (?P<fstype_linux>\S+)\s+\()|(\((?P<fstype_bsd>\S+),))(?P<options>.*)\)`

// Type is the type of items that this returns (Required)
func (s *MountsSource) Type() string {
	return "mount"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *MountsSource) Name() string {
	return "etcdata-mount"
}

// Weighting of duplicate sources
func (s *MountsSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *MountsSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *MountsSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var item *sdp.Item
	var scanner *bufio.Scanner
	var err error

	scanner, err = s.Scanner(ctx)

	if err != nil {
		return nil, err
	}

	MountParser.WithParsedResults(ctx, scanner, func(m map[string]string) bool {
		if m["path"] == query {
			item = mountMatchToItem(m)
			return false
		}

		// If the item wasn't found, continue
		return true
	})

	if item == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("user not found: %v", query),
		}
	}

	return item, nil
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *MountsSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	return s.Search(ctx, itemContext, "*")
}

// Search Searches by path or device (exact match). A "*" can also be passed
// which will return all items
func (s *MountsSource) Search(ctx context.Context, itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	isWildcard := (query == "*")

	results := make([]*sdp.Item, 0)

	scanner, err := s.Scanner(ctx)

	if err != nil {
		return nil, err
	}

	MountParser.WithParsedResults(ctx, scanner, func(m map[string]string) bool {
		// Check for a match
		if m["path"] == query || m["device"] == query || isWildcard {
			results = append(results, mountMatchToItem(m))
		}
		return true
	})

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	return results, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (s *MountsSource) Supported() bool {
	// Check that we have the mount executable
	_, err := exec.LookPath("mount")

	return err == nil
}

func mountMatchToItem(m map[string]string) *sdp.Item {
	var path string
	var device string
	var fstype string
	var options []string

	device = m["device"]
	path = m["path"]
	options = make([]string, 0)

	// There are two different way of capturing the filesystem type. On linux
	// mount returns something like this:
	//
	// device on /path type fstype (comma,sep,options)
	//
	// Whereas on many other OSes like Darwin, it's different:
	//
	// /dev/device on /path (fstype, comma, sep, options)
	//
	// The regex accounts for this but we need to grab whichever one matched
	if m["fstype_linux"] != "" {
		fstype = m["fstype_linux"]
	} else {
		fstype = m["fstype_bsd"]
	}

	// Grab the options as a slice
	for _, option := range strings.Split(m["options"], ",") {
		options = append(options, strings.TrimSpace(option))
	}

	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"device":  device,
		"path":    path,
		"fstype":  fstype,
		"options": options,
	})

	if err != nil {
		return nil
	}

	return &sdp.Item{
		Type:            "mount",
		UniqueAttribute: "path",
		Attributes:      attributes,
		Context:         util.LocalContext,
		LinkedItemRequests: []*sdp.ItemRequest{
			{
				Type:    "file",
				Method:  sdp.RequestMethod_GET,
				Query:   m["path"],
				Context: util.LocalContext,
			},
		},
	}
}

// mount executes /bin/mount and returns a scanner for the output
func mountCMD(ctx context.Context) (*bufio.Scanner, error) {
	// Execute mount. This requires a timeout as mount can hang
	cmd := exec.CommandContext(ctx, "mount")
	b, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	return bufio.NewScanner(bytes.NewReader(b)), nil
}
