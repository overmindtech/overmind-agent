package system

import (
	"context"
	"fmt"
	"net"

	"github.com/elastic/go-sysinfo"
	"github.com/elastic/go-sysinfo/types"
	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

// SystemSource struct on which all methods are registered
type SystemSource struct{}

//
// Below are the actual methods for the backend itself
//

// Type is the type of items that this returns (Required)
func (s *SystemSource) Type() string {
	return "system"
}

// Name Returns the name of the backend package. This is used for
// debugging and logging (Required)
func (s *SystemSource) Name() string {
	return "system"
}

// Weighting of duplicate sources
func (s *SystemSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *SystemSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get takes the UniqueAttribute value as a parameter (also referred to as the
// "name" of the item) and returns a full item will all details. This function
// must return an item whose UniqueAttribute value exactly matches the supplied
// parameter. If the item cannot be found it should return an ItemNotFoundError
// (Required)
func (s *SystemSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	return nil, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_NOTFOUND,
		ErrorString: "backend does not respond to Get() since this only every returns one thing. Use Find() instead",
		Context:     itemContext,
	}
}

// IgnoredRanges IP Ranges thatare ignored when generating linked items for a
// host's IPs
var IgnoredRanges = []string{
	"169.254.0.0/16", // Link-local
	"fe80::/10",      // Link-local
	"127.0.0.0/8",    // Loopback
	"::1/128",        // Loopback
}

func shouldIgnoreIP(ip net.IP) bool {
	for _, cidr := range IgnoredRanges {
		_, ignoreNet, err := net.ParseCIDR(cidr)

		if err == nil {
			if ignoreNet.Contains(ip) {
				return true
			}
		}
	}

	return false
}

// Find Gets information about all item that the source can possibly find. If
// nothing is found then just return an empty list (Required)
func (s *SystemSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var foundItems []*sdp.Item
	var systemItem *sdp.Item
	var attributes *sdp.ItemAttributes
	var attrMap map[string]interface{}
	var host types.Host
	var i types.HostInfo
	var m *types.HostMemoryInfo
	var err error

	host, err = sysinfo.Host()

	if err != nil {
		return nil, err
	}

	i = host.Info()

	attrMap = map[string]interface{}{
		"architecture":      i.Architecture,      // Hardware architecture (e.g. x86_64, arm, ppc, mips).
		"bootTime":          i.BootTime,          // Host boot time.
		"containerized":     i.Containerized,     // Is the process containerized.
		"hostname":          i.Hostname,          // Hostname
		"iPs":               i.IPs,               // List of all IPs.
		"kernelVersion":     i.KernelVersion,     // Kernel version.
		"mACs":              i.MACs,              // List of MAC addresses.
		"oS":                i.OS,                // OS information.
		"timezone":          i.Timezone,          // System timezone.
		"timezoneOffsetSec": i.TimezoneOffsetSec, // Timezone offset (seconds from UTC).
		"uniqueID":          i.UniqueID,          // Unique ID of the host (optional).
	}

	m, err = host.Memory()

	if err == nil {
		attrMap["Memory"] = m
	}

	attributes, err = sdp.ToAttributes(attrMap)

	if err != nil {
		return nil, err
	}

	systemItem = &sdp.Item{
		Type:            "system",
		UniqueAttribute: "hostname",
		Attributes:      attributes,
		Context:         util.LocalContext,
	}

	// Create a linked item for each IP address
	for _, ipString := range i.IPs {
		ip, _, err := net.ParseCIDR(ipString)

		if err != nil {
			continue
		}

		// Check if the IP should be ignored
		if shouldIgnoreIP(ip) {
			continue
		}

		systemItem.LinkedItemRequests = append(systemItem.LinkedItemRequests, &sdp.ItemRequest{
			Type:    "ip",
			Method:  sdp.RequestMethod_GET,
			Query:   ip.String(),
			Context: "global", // IPs live at the global context since they *should* be globally unique, or at least for our purposes
		})
	}

	foundItems = []*sdp.Item{
		systemItem,
	}

	return foundItems, nil
}
