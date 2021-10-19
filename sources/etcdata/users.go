package etcdata

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/deviant-agent/sources/util/parser"
	"github.com/dylanratcliffe/sdp-go"
)

// UsersSource struct on which all methods are registered
type UsersSource struct {
	PasswdLocation string
}

// DefaultPasswdLocation the default location of the passwd file
const DefaultPasswdLocation = "/etc/passwd"

// PasswdFileRegex Regex that is used to parse the /etc/passwd file
const PasswdFileRegex = `^(?P<username>[^:]*):(?P<password>[^:]*):(?P<uid>[^:]*):(?P<gid>[^:]*):(?P<comment>[^:]*):(?P<home>[^:]*):(?P<shell>[^:]*)$`

// PasswdParser A Parser for the /etc/passwd file
var PasswdParser = parser.Parser{
	Ignore:  regexp.MustCompile("^#"),
	Capture: regexp.MustCompile(PasswdFileRegex),
}

// PasswdFile opens and returns the passwd file
func (s *UsersSource) PasswdFile() (*os.File, error) {
	if s.PasswdLocation != "" {
		return os.Open(s.PasswdLocation)
	}

	return os.Open(DefaultPasswdLocation)
}

// Type is the type of items that this returns (Required)
func (s *UsersSource) Type() string {
	return "user"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *UsersSource) Name() string {
	return "etcdata-user"
}

// Weighting of duplicate sources
func (s *UsersSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *UsersSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *UsersSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var item *sdp.Item

	file, err := s.PasswdFile()
	scanner := bufio.NewScanner(file)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	PasswdParser.WithParsedResults(scanner, func(m map[string]string) bool {
		if m["username"] == query {
			item = userMatchToItem(m)

			// If we have found what we are looking for, stop
			return false
		}

		// If the item wasn't found, continue
		return true
	})

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

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
func (s *UsersSource) Find(itemContext string) ([]*sdp.Item, error) {
	return s.Search(itemContext, "*")
}

// Search Searches by UID or Name (exact match). A "*" can also be passed which
// will return all items
func (s *UsersSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	isWildcard := (query == "*")

	results := make([]*sdp.Item, 0)

	file, err := s.PasswdFile()
	scanner := bufio.NewScanner(file)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	PasswdParser.WithParsedResults(scanner, func(m map[string]string) bool {
		// Check for a match
		if m["username"] == query || m["uid"] == query || isWildcard {
			results = append(results, userMatchToItem(m))
		}
		return true
	})

	return results, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (s *UsersSource) Supported() bool {
	// Check that we can open the passwd file
	file, err := s.PasswdFile()

	if err != nil {
		return false
	}

	file.Close()

	return true
}

func userMatchToItem(m map[string]string) *sdp.Item {
	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"username": m["username"],
		"password": m["password"],
		"uid":      m["uid"],
		"gid":      m["gid"],
		"comment":  m["comment"],
		"home":     m["home"],
		"shell":    m["shell"],
	})

	if err != nil {
		return nil
	}

	return &sdp.Item{
		Type:            "user",
		UniqueAttribute: "username",
		Attributes:      attributes,
		Context:         util.LocalContext,
		LinkedItemRequests: []*sdp.ItemRequest{
			{
				Query:  m["gid"],
				Method: sdp.RequestMethod_SEARCH,
				Type:   "group",
			},
			{
				Query:  m["home"],
				Method: sdp.RequestMethod_GET,
				Type:   "file",
			},
			{
				Query:  m["shell"],
				Method: sdp.RequestMethod_GET,
				Type:   "file",
			},
		},
	}
}
