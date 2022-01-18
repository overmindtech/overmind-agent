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
	"sync"
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
	getFunction           func(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	getFunctionMutex      sync.Mutex
	getFunctionDetermined bool

	searchFunction           func(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	searchFunctionMutex      sync.Mutex
	searchFunctionDetermined bool

	dbusConnection      *dbus.Conn
	dbusConnectionMutex sync.Mutex
}

// DBusConnection Returns the current shared dBus connection, creating a new one
// if required
func (s *ServiceSource) DBusConnection() (*dbus.Conn, error) {
	s.dbusConnectionMutex.Lock()
	defer s.dbusConnectionMutex.Unlock()

	if s.dbusConnection == nil {
		var err error

		s.dbusConnection, err = dbus.NewWithContext(context.Background())

		if err != nil {
			return s.dbusConnection, err
		}
	}

	return s.dbusConnection, nil
}

// getFunction Returns the function that should be used to get units. This will
// end up calling different methods of dBus depending on availability
func (s *ServiceSource) GetFunction() (func(ctx context.Context, units []string) ([]dbus.UnitStatus, error), error) {
	s.getFunctionMutex.Lock()
	defer s.getFunctionMutex.Unlock()

	if s.getFunctionDetermined {
		return s.getFunction, nil
	}

	conn, err := s.DBusConnection()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// See if we can use ListUnitsByNamesContext (preferred)
	if _, err := conn.ListUnitsByNamesContext(ctx, []string{"*"}); err == nil {
		s.getFunction = conn.ListUnitsByNamesContext
	} else {
		s.getFunction = s.bruteSearch
	}
	s.getFunctionDetermined = true

	return s.getFunction, nil
}

// searchFunction Returns the function that should be used to get units. This will
// end up calling different methods of dBus depending on availability
func (s *ServiceSource) SearchFunction() (func(ctx context.Context, units []string) ([]dbus.UnitStatus, error), error) {
	s.searchFunctionMutex.Lock()
	defer s.searchFunctionMutex.Unlock()

	if s.searchFunctionDetermined {
		return s.searchFunction, nil
	}

	conn, err := s.DBusConnection()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if _, err := conn.ListUnitsByPatternsContext(ctx, []string{"loaded"}, []string{"foobar"}); err == nil {
		s.searchFunction = func(ctx context.Context, patterns []string) ([]dbus.UnitStatus, error) {
			conn, err := s.DBusConnection()
			if err != nil {
				return nil, err
			}
			return conn.ListUnitsByPatternsContext(ctx, []string{"loaded"}, patterns)
		}
	} else {
		s.searchFunction = s.bruteSearch
	}
	s.searchFunctionDetermined = true

	return s.searchFunction, nil
}

// bruteSearch Replicates the functionality of ListUnitsByPatternsContext but
// using a brute force approach of getting all units and then filtering by the
// given patterns. This is required on old versions of systemd that don't
// implement better methods like ListUnitsByPatternsContext
func (s *ServiceSource) bruteSearch(ctx context.Context, patterns []string) ([]dbus.UnitStatus, error) {
	conn, err := s.DBusConnection()

	if err != nil {
		return nil, err
	}

	var allUnits []dbus.UnitStatus
	var filteredUnits []dbus.UnitStatus

	allUnits, err = conn.ListUnitsContext(ctx)

	if err != nil {
		return allUnits, err
	}

	var patternRegexes []*regexp.Regexp

	// Convert paterns to regexes
	for _, pattern := range patterns {
		patternRegexes = append(patternRegexes, regexp.MustCompile(wildCardToRegexp(pattern)))
	}

	for _, unit := range allUnits {
		if unit.LoadState != "loaded" {
			continue
		}

		var patternMatched bool

		for _, r := range patternRegexes {
			if patternMatched {
				continue
			}

			if r.MatchString(unit.Name) {
				filteredUnits = append(filteredUnits, unit)
				patternMatched = true
			}
		}
	}

	return filteredUnits, nil
}

// wildCardToRegexp converts a wildcard pattern to a regular expression pattern.
func wildCardToRegexp(pattern string) string {
	var result strings.Builder
	for i, literal := range strings.Split(pattern, "*") {

		// Replace * with .*
		if i > 0 {
			result.WriteString(".*")
		}

		// Quote any regular expression meta characters in the
		// literal text.
		result.WriteString(regexp.QuoteMeta(literal))
	}
	return result.String()
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
func (s *ServiceSource) Type() string {
	return "service"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *ServiceSource) Name() string {
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
func (s *ServiceSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var units []dbus.UnitStatus
	var item *sdp.Item
	var getFunc func(ctx context.Context, units []string) ([]dbus.UnitStatus, error)
	var err error

	getFunc, err = s.GetFunction()

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error connecting to dbus to get systemd services: %v", err),
			Context:     itemContext,
		}
	}

	// This matches the LoadState of "loaded"
	// https://www.freedesktop.org/wiki/Software/systemd/dbus/
	units, err = getFunc(
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

	item, err = s.mapUnitToItem(ctx, units[0])

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
func (s *ServiceSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
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
		item, err = s.mapUnitToItem(ctx, unit)

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
func (s *ServiceSource) Search(ctx context.Context, itemContext string, query string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var err error
	var units []dbus.UnitStatus
	var item *sdp.Item
	var items []*sdp.Item
	var searchFunc func(ctx context.Context, units []string) ([]dbus.UnitStatus, error)

	searchFunc, err = s.SearchFunction()

	if err != nil {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("Error connecting to dbus to get systemd services: %v", err),
			Context:     itemContext,
		}
	}

	if i, err := strconv.Atoi(query); err == nil {
		if conn, err := s.DBusConnection(); err == nil {
			if name, err := conn.GetUnitNameByPID(ctx, uint32(i)); err == nil {
				// If the query is an integer, and that integer corresponds to a
				// PID, which corresponds to a unit, search by that name instead
				query = name
			}
		}
	}

	// This matches the LoadState of "loaded"
	// https://www.freedesktop.org/wiki/Software/systemd/dbus/
	units, err = searchFunc(
		ctx,
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
		item, err = s.mapUnitToItem(ctx, unit)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, nil
}

// Supported A function that can be executed to see if the backend is supported
// in the current environment, if it returns false the backend simply won't be
// loaded (Optional)
func (s *ServiceSource) Supported() bool {
	return systemdUtil.IsRunningSystemd()
}

// mapUnitToItem Maps a unit to a "service" item
func (s *ServiceSource) mapUnitToItem(ctx context.Context, u dbus.UnitStatus) (*sdp.Item, error) {
	var a map[string]interface{}
	var attributes *sdp.ItemAttributes
	var err error
	var linkedItemRequests []*sdp.ItemRequest
	var binaries map[string]bool
	var c *dbus.Conn

	a = make(map[string]interface{})
	binaries = make(map[string]bool)
	c, err = s.DBusConnection()

	if err != nil {
		return nil, err
	}

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
