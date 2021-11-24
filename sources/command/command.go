package command

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

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

// Get Runs a single command as specified in the query. This should include the
// full command to run, including arguments, as a string. This command will be
// run with default values for all other parametres. For more complex commands,
// use Search()
func (s *CommandSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	params := CommandParams{
		Command: query,
	}

	return params.Run(ctx)
}

// Search runs a command with a given set of parameteres. These paremeteres
// should be supplied as a single JSON string, in the format of a
// `CommandParams` struct e.g.
//
// {
//     "command": "cat /etc/hosts",
//     "expected_exit": 0,
//     "timeout": "5s",
//     "dir": "/etc",
//     "env": {
//         "FOO": "BAR"
//     }
// }
func (s *CommandSource) Search(ctx context.Context, itemContext string, query string) ([]*sdp.Item, error) {
	var params CommandParams
	var item *sdp.Item

	items := make([]*sdp.Item, 0)
	err := json.Unmarshal([]byte(query), &params)

	if err != nil {
		return items, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("could not unmarshal JSON query, error: %v", err.Error()),
			Context:     util.LocalContext,
		}
	}

	item, err = params.Run(ctx)

	items = append(items, item)

	return items, err
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *CommandSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	return nil, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_OTHER,
		ErrorString: "the command source only supports the Get and Search methods",
		Context:     itemContext,
	}
}

// Hidden command items should be hidden as they are only used when needed,
// and should be covered by some higher layer of abstraction like a secondary
// source
func (s *CommandSource) Hidden() bool {
	return true
}
