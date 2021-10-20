package etcdata

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/deviant-agent/sources/util/parser"
	"github.com/dylanratcliffe/sdp-go"
)

// GroupsSource struct on which all methods are registered
type GroupsSource struct {
	GroupsLocation string
}

// DefaultGroupsLocation the default location of the group file
const DefaultGroupsLocation = "/etc/group"

// GroupFileRegex Regex that is used to parse the /etc/passwd file
const GroupFileRegex = `^(?P<name>[^:]*):(?P<password>[^:]*):(?P<gid>[^:]*):(?P<members>[^:]*)$`

// GroupParser A Parser for the /etc/passwd file
var GroupParser = parser.Parser{
	Ignore:  regexp.MustCompile("^#"),
	Capture: regexp.MustCompile(GroupFileRegex),
}

// GroupFile opens and returns the passwd file
func (s *GroupsSource) GroupFile() (*os.File, error) {
	if s.GroupsLocation != "" {
		return os.Open(s.GroupsLocation)
	}

	return os.Open(DefaultGroupsLocation)
}

// Type is the type of items that this returns (Required)
func (s *GroupsSource) Type() string {
	return "group"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *GroupsSource) Name() string {
	return "etcdata-groups"
}

// Weighting of duplicate sources
func (s *GroupsSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *GroupsSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *GroupsSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var item *sdp.Item

	file, err := s.GroupFile()
	scanner := bufio.NewScanner(file)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	GroupParser.WithParsedResults(scanner, func(m map[string]string) bool {
		if m["name"] == query {
			item = groupMatchToItem(m)

			// If we have found what we are looking for, stop
			return false
		}

		// If the item wasn't found, continue
		return true
	})

	if item == nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("group not found: %v", query),
		}
	}

	return item, nil
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *GroupsSource) Find(itemContext string) ([]*sdp.Item, error) {
	return s.Search(itemContext, "*")
}

// Search Searches by UID or Name (exact match). A "*" can also be passed which
// will return all items
func (s *GroupsSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	isWildcard := (query == "*")

	results := make([]*sdp.Item, 0)

	file, err := s.GroupFile()
	scanner := bufio.NewScanner(file)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	GroupParser.WithParsedResults(scanner, func(m map[string]string) bool {
		// Check for a match
		if m["name"] == query || m["gid"] == query || isWildcard {
			results = append(results, groupMatchToItem(m))
		}
		return true
	})

	return results, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (s *GroupsSource) Supported() bool {
	// Check that we can open the passwd file
	file, err := s.GroupFile()

	if err != nil {
		return false
	}

	file.Close()

	return true
}

func groupMatchToItem(m map[string]string) *sdp.Item {
	var members []string
	var linkedItemRequests []*sdp.ItemRequest

	// Split the members on commas
	if membersString, ok := m["members"]; ok {
		if membersString != "" {
			members = strings.Split(membersString, ",")
		}
	}

	for _, member := range members {
		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:    "user",
			Method:  sdp.RequestMethod_SEARCH,
			Query:   member,
			Context: util.LocalContext,
		})
	}

	attributes, err := sdp.ToAttributes(map[string]interface{}{
		"name":    m["name"],
		"gid":     m["gid"],
		"members": members,
	})

	if err != nil {
		return nil
	}

	return &sdp.Item{
		Type:               "group",
		UniqueAttribute:    "name",
		Attributes:         attributes,
		Context:            util.LocalContext,
		LinkedItemRequests: linkedItemRequests,
	}
}
