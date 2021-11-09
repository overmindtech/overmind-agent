package command

import (
	"fmt"
	"strings"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

const GetTimeout = 10 * time.Second

// CommandSource struct on which all methods are registered
type CommandSource struct{}

//
// Below are the actual methods for the backend itself
//

// Type is the type of items that this returns (Required)
func (s *CommandSource) Type() string {
	return "command"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *CommandSource) Name() string {
	return "command"
}

// Weighting of duplicate sources
func (s *CommandSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *CommandSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

func (s *CommandSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	splitQuery := strings.Split(query, " ")

	// commandName := splitQuery[0]
	args := make([]string, len(splitQuery)-1)

	// Loop over the remaining arguments and copy into new slice
	for i := 1; i != len(splitQuery)-1; i++ {
		args[i-1] = splitQuery[i]
	}

	return nil, nil

}

func (s *CommandSource) Search(itemContext string, query string) ([]*sdp.Item, error) {
	return []*sdp.Item{}, nil
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *CommandSource) Find(itemContext string) ([]*sdp.Item, error) {
	return nil, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_OTHER,
		ErrorString: "the command source only supports the Get and Search methods",
		Context:     itemContext,
	}
}
