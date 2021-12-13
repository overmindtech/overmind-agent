//go:build linux
// +build linux

package systemd

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	systemdUtil "github.com/coreos/go-systemd/v22/util"

	"github.com/overmindtech/overmind-agent/sources/util"

	"github.com/overmindtech/sdp-go"
)

// This backend gathers systemd information though the the systemd D-Bus API.
// See http://www.freedesktop.org/wiki/Software/systemd/dbus/
//
// Note that since this is a C API and uses CGO under the hood, it won't compile
// on Darwin or Windows for example. For developers running non-linux platforms
// looking to work on this code, I would recommend connecting to a compatible
// system remotely: https://code.visualstudio.com/docs/remote/remote-overview or
// just developing directly on a compatible system

// ServiceSource struct on which all methods are registered
type ServiceSource struct {
	Timeout time.Duration
}

// DEFAULT_TIMEOUT Default DBUS query timeout
const DEFAULT_TIMEOUT = (10 * time.Second)

var serviceRegex = regexp.MustCompile(`^(.*)\.service$`)

// ServiceBasicProperties The list of properties to include if they are
// non-zero. The "basic" properties have simple values types that can be
// compared easily to their zero values
var ServiceBasicProperties = []string{
	"BusName",
	"ExecCondition",
	"ExecMainPID",
	"FileDescriptorStoreMax",
	"GuessMainPID",
	"MemoryCurrent",
	"NonBlocking",
	"NotifyAccess",
	"OOMPolicy",
	"PIDFile",
	"RemainAfterExit",
	"Restart",
	"RestartSec",
	"RootDirectoryStartOnly",
	"RuntimeMaxSec",
	"Sockets",
	"TimeoutAbortSec",
	"TimeoutSec",
	"TimeoutStartFailureMode",
	"TimeoutStartSec",
	"TimeoutStopSec",
	"Type",
	"USBFunctionDescriptors",
	"USBFunctionStrings",
	"WatchdogSec",
}

// ServiceExecProperties A list of properties that contain commands to execute.
// These are stored as [][]interface{} and need special logic to extract
var ServiceExecProperties = []string{
	"ExecCondition",
	"ExecReload",
	"ExecStart",
	"ExecStartPre",
	"ExecStop",
	"ExecStopPost",
}

// ServiceStatusProperties Properties that list exit statuses which need special
// logic
var ServiceStatusProperties = []string{
	"RestartForceExitStatus",
	"RestartPreventExitStatus",
	"SuccessExitStatus",
}

// Type is the type of items that this returns (Required)
func (bc *ServiceSource) Type() string {
	return "service"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (bc *ServiceSource) Name() string {
	return "systemd"
}

// Weighting of duplicate sources
func (s *ServiceSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *ServiceSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (bc *ServiceSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	// TODO: Should we be re-using the DBUS connection?
	var c *dbus.Conn
	var err error
	var units []dbus.UnitStatus
	var item *sdp.Item

	c, err = dbus.NewWithContext(ctx)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error connecting to dbus to get systemd services: %v", err),
			Context:     itemContext,
		}
	}

	// Cacnel once we're done
	defer c.Close()

	// This matches the LoadState of "loaded"
	// https://www.freedesktop.org/wiki/Software/systemd/dbus/
	units, err = c.ListUnitsByNamesContext(
		ctx,
		[]string{query},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error getting services from systemd via dbus: %v", err),
			Context:     itemContext,
		}
	}

	if len(units) < 1 {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Service %v not found", query),
			Context:     itemContext,
		}
	}

	if len(units) > 1 {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf(">1 service for %v found: %v", query, units),
			Context:     itemContext,
		}
	}

	item, err = mapUnitToItem(ctx, units[0], c)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error when mapping service %v: %v", query, err),
			Context:     itemContext,
		}
	}

	return item, nil
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (bc *ServiceSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var c *dbus.Conn
	var err error
	var units []dbus.UnitStatus
	var item *sdp.Item
	var items []*sdp.Item

	c, err = dbus.NewWithContext(ctx)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error connecting to dbus to get systemd services: %v", err),
			Context:     itemContext,
		}
	}

	defer c.Close()

	units, err = c.ListUnitsContext(ctx)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error getting services from systemd via dbus: %v", err),
			Context:     itemContext,
		}
	}

	for _, unit := range units {
		item, err = mapUnitToItem(ctx, unit, c)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Search searches by a glob pattern, but also appends the following default
// glob patterns:
//   - {input}.*
//
// If the query is an integer, it will be converted to the name of the units
// which owns a process with that PID
//
func (bc *ServiceSource) Search(ctx context.Context, itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var c *dbus.Conn
	var err error
	var units []dbus.UnitStatus
	var item *sdp.Item
	var items []*sdp.Item

	c, err = dbus.NewWithContext(ctx)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error connecting to dbus to get systemd services: %v", err),
			Context:     itemContext,
		}
	}

	defer c.Close()

	if i, err := strconv.Atoi(query); err == nil {
		if name, err := c.GetUnitNameByPID(ctx, uint32(i)); err == nil {
			// If the query is an integer, and that integer corresponds to a
			// PID, which corresponds to a unit, search by that name instead
			query = name
		}
	}

	// This matches the LoadState of "loaded"
	// https://www.freedesktop.org/wiki/Software/systemd/dbus/
	units, err = c.ListUnitsByPatternsContext(
		ctx,
		[]string{"loaded"},
		[]string{
			query,
			fmt.Sprintf("%v.*", query),
		},
	)

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error getting services from systemd via dbus: %v", err),
			Context:     itemContext,
		}
	}

	for _, unit := range units {
		item, err = mapUnitToItem(ctx, unit, c)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (bc *ServiceSource) Supported() bool {
	return systemdUtil.IsRunningSystemd()
}

// mapUnitToItem Maps a unit to a "service" item
func mapUnitToItem(ctx context.Context, u dbus.UnitStatus, c *dbus.Conn) (*sdp.Item, error) {
	var a map[string]interface{}
	var attributes *sdp.ItemAttributes
	var err error
	var linkedItemRequests []*sdp.ItemRequest
	var binaries map[string]bool

	a = make(map[string]interface{})
	binaries = make(map[string]bool)

	match := serviceRegex.MatchString(u.Name)

	// Filter for services, since a "unit" can be much more than a service
	if !match {
		return nil, errors.New("unit was not a service")
	}

	// Get basic details
	a["Name"] = u.Name
	a["Description"] = u.Description
	a["LoadState"] = u.LoadState
	a["ActiveState"] = u.ActiveState
	a["Path"] = u.Path

	if u.SubState != "" {
		a["SubState"] = u.SubState
	}

	if p, e := c.GetUnitPropertyContext(ctx, u.Name, "FragmentPath"); e == nil {
		a["FragmentPath"] = p.Value
	}

	// Loop over the ServiceProperties and check if they are non-zero
	for _, propName := range ServiceBasicProperties {
		prop, err := c.GetUnitTypePropertyContext(ctx, u.Name, "Service", propName)

		if err == nil {
			v := reflect.ValueOf(prop.Value.Value())

			if !v.IsZero() {
				a[prop.Name] = prop.Value.Value()
			}
		}
	}

	// Loop over exec properties and extract them. This requires different logic
	// since they are store in a much more complex data structure
	for _, propName := range ServiceExecProperties {
		prop, err := c.GetUnitTypePropertyContext(ctx, u.Name, "Service", propName)

		if err == nil {
			// Check the data is the type we are expecting
			val, ok := prop.Value.Value().([][]interface{})

			// Extract useful info
			if ok && len(val) > 0 {
				execDetails := make(map[string]interface{})

				for _, v := range val {
					// ExecStartPre, ExecStart, ExecStartPost, ExecReload,
					// ExecStop, ExecStop each are arrays of structures each
					// containing: the binary path to execute; an array with all
					// arguments to pass to the executed command, starting with
					// argument 0; a boolean whether it should be considered a
					// failure if the process exits uncleanly; two pairs of
					// CLOCK_REALTIME/CLOCK_MONOTONIC usec timestamps when the
					// process began and finished running the last time, or 0 if
					// it never ran or never finished running; the PID of the
					// process, or 0 if it has not run yet; the exit code and
					// status of the last run. This field hence maps more or
					// less to the corresponding setting in the service unit
					// file but is augmented with runtime data.
					//
					// https://www.freedesktop.org/wiki/Software/systemd/dbus/
					execDetails["binary"] = v[0]
					execDetails["args"] = v[1].([]string)[1:]
					execDetails["fullCMD"] = strings.Join(v[1].([]string), " ")

					// Note down the binary for linking
					binaries[v[0].(string)] = true
				}

				a[prop.Name] = execDetails
			}
		}
	}

	for _, propName := range ServiceStatusProperties {
		prop, err := c.GetUnitTypePropertyContext(ctx, u.Name, "Service", propName)

		if err == nil {
			slices, ok := prop.Value.Value().([]interface{})

			// I can't find any documentation about the return format of this.
			// For some reason it's [][]int32 where the first slice of int32s is
			// the actual values as specified in the config file, and I have no
			// idea what the other slice is for
			if ok && len(slices) == 2 {
				// The first slice contains the data we want
				statuses, ok := slices[0].([]int32)

				if ok && len(statuses) > 0 {
					a[prop.Name] = statuses
				}
			}
		}
	}

	attributes, err = sdp.ToAttributes(a)

	if err != nil {
		return nil, err
	}

	// Create linked item requests

	// Link to the file representing the binary that the service runs as
	for binary := range binaries {
		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:    "file",
			Method:  sdp.RequestMethod_GET,
			Query:   binary,
			Context: util.LocalContext,
		})
	}

	// Link to the PID of the service
	if pid, err := attributes.Get("ExecMainPID"); err == nil {
		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:    "process",
			Method:  sdp.RequestMethod_GET,
			Query:   fmt.Sprint(pid),
			Context: util.LocalContext,
		})
	}

	// Link to the user
	if name, err := attributes.Get("User"); err == nil {
		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:    "user",
			Method:  sdp.RequestMethod_SEARCH,
			Query:   fmt.Sprint(name),
			Context: util.LocalContext,
		})
	}

	// Link to group
	if name, err := attributes.Get("Group"); err == nil {
		linkedItemRequests = append(linkedItemRequests, &sdp.ItemRequest{
			Type:    "group",
			Method:  sdp.RequestMethod_SEARCH,
			Query:   fmt.Sprint(name),
			Context: util.LocalContext,
		})
	}

	return &sdp.Item{
		Type:               "service",
		UniqueAttribute:    "Name",
		Attributes:         attributes,
		Context:            util.LocalContext,
		LinkedItemRequests: linkedItemRequests,
	}, nil
}
