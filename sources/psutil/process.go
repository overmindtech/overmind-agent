package psutil

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// ProcessSource struct on which all methods are registered
type ProcessSource struct{}

// Type is the type of items that this returns (Required)
func (s *ProcessSource) Type() string {
	return "process"
}

// Descriptive name for the source, used in logging and metadata
func (s *ProcessSource) Name() string {
	return "psutil"
}

// Weighting of duplicate sources
func (s *ProcessSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *ProcessSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

func (s *ProcessSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var pid int
	var err error
	var p *process.Process
	var item sdp.Item

	// Convert PID to an integer
	pid, err = strconv.Atoi(query)

	if err != nil {
		// I'm counting this is not found since if the PID isn't an integer it
		// definitely won't exist
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: fmt.Sprintf("PID could not be converted to integer, encountered error: %v", err),
			Context:     itemContext,
		}
	}

	// Create the process object so that we can inspect it
	p, err = process.NewProcessWithContext(ctx, int32(pid))

	if err != nil {
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	// Create initial item details
	item.Type = "process"
	item.UniqueAttribute = "pid"
	item.Context = util.LocalContext
	item.LinkedItemRequests = make([]*sdp.ItemRequest, 0)
	attributes := make(map[string]interface{})

	// Process attributes
	var cpuPercent float64
	var cmdline string
	var connections []net.ConnectionStat
	var createTime int64
	var cwd string
	var exe string
	var groups []int32
	var groupsInterface []interface{}
	var ioCounters *process.IOCountersStat
	var ionice int32
	var isRunning bool
	var memoryPercent float32
	var name string
	var nice int32
	var pageFaults *process.PageFaultsStat
	var ppid int32
	var status string
	var terminal string
	var username string

	attributes["pid"] = p.Pid

	// Look for the service that owns the pid
	item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
		Type:    "service",
		Method:  sdp.RequestMethod_SEARCH,
		Query:   fmt.Sprint(p.Pid),
		Context: itemContext,
	})

	if cpuPercent, err = p.CPUPercentWithContext(ctx); err == nil {
		attributes["cpuPercent"] = cpuPercent
	}

	if cmdline, err = p.CmdlineWithContext(ctx); err == nil {
		attributes["cmdline"] = cmdline
	}

	if connections, err = p.ConnectionsWithContext(ctx); err == nil {
		// TODO: See if there is a better way to get connection info without it
		// being super transient and overwhelming. i.e. if we were to get the
		// details of every connection there might be thousands and they might
		// change completely every few seconds
		attributes["numConnections"] = len(connections)
	}

	if createTime, err = p.CreateTimeWithContext(ctx); err == nil {
		attributes["createTime"] = time.Unix(0, (createTime * 1000000)).String()
	}

	if cwd, err = p.CwdWithContext(ctx); err == nil {
		attributes["cwd"] = cwd

		// We can create a link to the file here
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "file",
			Method:  sdp.RequestMethod_GET,
			Query:   cwd,
			Context: itemContext,
		})
	}

	if exe, err = p.ExeWithContext(ctx); err == nil {
		attributes["exe"] = exe

		// We can create a link to the executable here
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "file",
			Method:  sdp.RequestMethod_GET,
			Query:   exe,
			Context: itemContext,
		})

		// Check what package provies the executable
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "package",
			Method:  sdp.RequestMethod_SEARCH,
			Query:   exe,
			Context: itemContext,
		})
	}

	if groups, err = p.GroupsWithContext(ctx); err == nil {
		// Convert int32 to interface{} since the protobuf library expects a
		// slice of interfaces
		for _, group := range groups {
			groupsInterface = append(groupsInterface, group)
		}
		attributes["groups"] = groupsInterface

		for _, gid := range groups {
			// Create linked item requests for the groups
			item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
				Type:    "group",
				Method:  sdp.RequestMethod_GET,
				Query:   strconv.Itoa(int(gid)),
				Context: itemContext,
			})
		}
	}

	if ioCounters, err = p.IOCountersWithContext(ctx); err == nil {
		attributes["ioReadBytes"] = ioCounters.ReadBytes
		attributes["ioWriteBytes"] = ioCounters.WriteBytes
		attributes["ioReadCount"] = ioCounters.ReadCount
		attributes["ioWriteCount"] = ioCounters.WriteCount
	}

	if ionice, err = p.IOniceWithContext(ctx); err == nil {
		attributes["ionice"] = ionice
	}

	if isRunning, err = p.IsRunningWithContext(ctx); err == nil {
		attributes["isRunning"] = isRunning
	}

	if memoryPercent, err = p.MemoryPercentWithContext(ctx); err == nil {
		attributes["memoryPercent"] = memoryPercent
	}

	if name, err = p.NameWithContext(ctx); err == nil {
		attributes["name"] = name
	}

	if nice, err = p.NiceWithContext(ctx); err == nil {
		attributes["nice"] = nice
	}

	if pageFaults, err = p.PageFaultsWithContext(ctx); err == nil {
		// We probably don't care about minor page faults since it doesn't hit
		// disk, however major page faults require reding from disk and could
		// seriously impact performance, so we probably want to be tracking
		// these
		attributes["pageFaultsMajor"] = pageFaults.MajorFaults
	}

	if ppid, err = p.PpidWithContext(ctx); err == nil {
		attributes["parent"] = ppid

		// Link to the parent process
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "process",
			Method:  sdp.RequestMethod_GET,
			Query:   strconv.Itoa(int(ppid)),
			Context: itemContext,
		})
	}

	if status, err = p.StatusWithContext(ctx); err == nil {
		var statusLong string

		switch status {
		case "R":
			statusLong = "Running"
		case "S":
			statusLong = "Sleep"
		case "T":
			statusLong = "Stop"
		case "I":
			statusLong = "Idle"
		case "Z":
			statusLong = "Zombie"
		case "W":
			statusLong = "Wait"
		case "L":
			statusLong = "Lock"
		default:
			statusLong = "Unknown"
		}

		attributes["status"] = statusLong
	}

	if terminal, err = p.TerminalWithContext(ctx); err == nil {
		attributes["terminal"] = terminal
	}

	if username, err = p.UsernameWithContext(ctx); err == nil {
		attributes["username"] = username

		// Link to the username
		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "user",
			Method:  sdp.RequestMethod_GET,
			Query:   username,
			Context: itemContext,
		})
	}

	// Save the attributes
	item.Attributes, err = sdp.ToAttributes(attributes)

	if err != nil {
		// I'm counting this is not found since if the PID isn't an integer it
		// definitely won't exist
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	return &item, err
}

// BlankFind always returns an empty list of items
func (s *ProcessSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	return make([]*sdp.Item, 0), nil
}
