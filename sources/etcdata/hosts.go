package etcdata

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

// TODO: Refactor this to use a Parser

// HostsSource struct on which all methods are registered
type HostsSource struct {
	// Where to find the hosts file if not the default location (optional)
	HostsLocation string
}

// Entry matches a hosts file entry
var Entry = regexp.MustCompile(`^(?P<address>[0-9a-f:]\S+)\s+(?P<name>[^#\s+]\S+)\s*(?P<aliases>.*?)?(?:\s*#\s*(?P<comment>.*))?$`)

// DefaultHostsLocation Default path to the hosts file
var DefaultHostsLocation string

// HostsFile Opens and returns the hosts file
func (s *HostsSource) HostsFile() (*os.File, error) {
	if s.HostsLocation != "" {
		return os.Open(s.HostsLocation)
	}

	return os.Open(DefaultHostsLocation)
}

// Type is the type of items that this returns (Required)
func (s *HostsSource) Type() string {
	return "host"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *HostsSource) Name() string {
	return "etcdata-hosts"
}

// Weighting of duplicate sources
func (s *HostsSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *HostsSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *HostsSource) Get(itemContext string, query string) (*sdp.Item, error) {
	items, err := s.Search(itemContext, query)

	if err != nil {
		return nil, err
	}

	if len(items) != 1 {
		return nil, fmt.Errorf("found >1 item from Get() with query %v", query)
	}

	return items[0], nil
}

// Find finds all items
func (s *HostsSource) Find(itemContext string) ([]*sdp.Item, error) {
	return s.Search(itemContext, "*")
}

// Search Search for a host entry, the query can be an ip, a name or '*'
func (s *HostsSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var file *os.File
	var err error
	var names []string
	var matches []string
	var line map[string]interface{}
	var items []*sdp.Item
	var wildcard bool

	items = make([]*sdp.Item, 0)
	line = make(map[string]interface{})
	wildcard = (query == "*")

	file, err = s.HostsFile()

	if err != nil {
		return nil, fmt.Errorf("could not open hosts file, error: %v", err)
	}

	defer file.Close()

	// Load capturing group names
	names = Entry.SubexpNames()

	// Read the file one line at a time
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches = Entry.FindStringSubmatch(scanner.Text())

		if matches != nil {
			var attributes *sdp.ItemAttributes

			// Populate the matches into a map so that we can look things up
			// e.g.
			//
			// line["address"]
			for i, match := range matches {
				matchName := names[i]
				if matchName != "" && match != "" {
					if matchName != "aliases" {
						line[matchName] = match
					} else {
						var aliases []interface{}
						fields := strings.Fields(match)

						for _, f := range fields {
							aliases = append(aliases, f)
						}

						line[matchName] = aliases
					}
				}
			}

			// Check if the host matches the query
			if !(query == line["address"] || query == line["name"] || wildcard) {
				// If the address or the name doesn't match this host and we
				// don't have a wildcard search, then discard
				continue
			}

			attributes, err = sdp.ToAttributes(line)

			if err != nil {
				return nil, err
			}

			items = append(items, &sdp.Item{
				Type:            "host",
				UniqueAttribute: "name",
				Attributes:      attributes,
				Context:         util.LocalContext,
				LinkedItemRequests: []*sdp.ItemRequest{
					{
						Type:    "ip",
						Method:  sdp.RequestMethod_GET,
						Query:   fmt.Sprint(line["address"]),
						Context: "global",
					},
				},
			})
		}
	}

	return items, scanner.Err()
}

// Set the variable when the package is loaded. Ideally this would be a constant
// but on Windows we need to see which drive the OS is installed on so that will
// have to be caluclated at runtime
func init() {
	switch runtime.GOOS {
	case "windows":
		DefaultHostsLocation = filepath.Join(
			os.Getenv("SystemRoot"),
			"System32/drivers/etc/hosts",
		)
	case "solaris":
		DefaultHostsLocation = "/etc/inet/hosts"
	default:
		DefaultHostsLocation = "/etc/hosts"
	}
}
