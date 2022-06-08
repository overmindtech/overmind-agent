//go:build linux || windows
// +build linux windows

package netstat

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strconv"

	ns "github.com/cakturk/go-netstat/netstat"
	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

// PortSource struct on which all methods are registered
type PortSource struct{}

// Type is the type of items that this returns
func (s *PortSource) Type() string {
	return "port"
}

// Name Returns the name of the backend
func (s *PortSource) Name() string {
	return "netstat"
}

// Supported returns whether the backend is supported on this platform
func (s *PortSource) Supported() bool {
	return (runtime.GOOS == "linux" || runtime.GOOS == "windows")
}

// Weighting of duplicate sources
func (s *PortSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *PortSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get calls back to the GetFunction to get a single item
func (s *PortSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var err error
	var sockets4 []ns.SockTabEntry
	var sockets6 []ns.SockTabEntry
	var sockets []ns.SockTabEntry
	var port int

	port, err = strconv.Atoi(query)

	if err != nil {
		return nil, fmt.Errorf("could not convert %v to uint16", query)
	}

	sockets4, _ = ns.TCPSocks(acceptNumber(uint16(port)))
	sockets6, _ = ns.TCP6Socks(acceptNumber(uint16(port)))
	sockets = append(sockets4, sockets6...)

	switch numSockets := len(sockets); numSockets {
	case 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Port %v not found", query),
		}
	default:
		return socketsToItem(sockets...)
	}
}

// Find calls back to the FindFunction to find all items
func (s *PortSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var items []*sdp.Item
	var errString string
	var err error

	ipv4Sockets, ipv4Err := ns.TCPSocks(acceptListening)

	if ipv4Err != nil {
		errString = fmt.Sprintf("error getting ipv4 sockets: %v ", ipv4Err)
	}

	ipv6Sockets, ipv6Err := ns.TCP6Socks(acceptListening)

	if ipv6Err != nil {
		errString = errString + fmt.Sprintf("error getting ipv6 sockets: %v", ipv6Err)
	}

	if errString != "" {
		err = errors.New(errString)
	}

	allSockets := append(ipv4Sockets, ipv6Sockets...)
	socketMap := make(map[uint16][]ns.SockTabEntry)

	for _, socket := range allSockets {
		if socket.LocalAddr != nil {
			socketMap[socket.LocalAddr.Port] = append(socketMap[socket.LocalAddr.Port], socket)
		}
	}

	for _, sockets := range socketMap {
		item, err := socketsToItem(sockets...)

		if err == nil {
			items = append(items, item)
		}
	}

	return items, err
}

func acceptListening(s *ns.SockTabEntry) bool {
	return s.State.String() == "LISTEN"
}

// acceptNumber Accepts a given port nuymber that is also listening
func acceptNumber(p uint16) func(*ns.SockTabEntry) bool {
	return func(n *ns.SockTabEntry) bool {
		return n.LocalAddr.Port == p && acceptListening(n)
	}
}

// socketsToItem Will merge details of sockets into a single item, this assumes
// that all sockets already have the same state and port and will simply merge
// the local and remote addresses
func socketsToItem(s ...ns.SockTabEntry) (*sdp.Item, error) {
	var err error
	var item sdp.Item

	attributes := make(map[string]interface{})

	if len(s) == 0 {
		return nil, errors.New("no sockets provided")
	}

	// Prepare the item
	item = sdp.Item{}
	item.Type = "port"
	item.UniqueAttribute = "port"
	item.Context = util.LocalContext
	item.LinkedItemRequests = make([]*sdp.ItemRequest, 0)

	// Get most of the details from the first socket since they should be the same
	attributes["state"] = s[0].State.String()
	attributes["port"] = uint32(s[0].LocalAddr.Port)
	localIPs := make([]string, 0)

	if s[0].Process != nil {
		attributes["pid"] = s[0].Process.Pid

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Method:  sdp.RequestMethod_GET,
			Query:   strconv.Itoa(s[0].Process.Pid),
			Type:    "process",
			Context: util.LocalContext,
		})
	}

	// Merge the IPs and anything else that might be different
	for _, socket := range s {
		if socket.LocalAddr != nil {
			localIPs = append(localIPs, socket.LocalAddr.IP.String())

			// If the IP is unspecified (0.0.0.0 or ::) it means that the port will
			// be listening on all IPs, therefore we need to do some expansion
			if socket.LocalAddr.IP.IsUnspecified() {
				unicastAddresses, err := net.InterfaceAddrs()

				if err == nil {
					for _, address := range unicastAddresses {
						// Extract just the IP address
						if ipNet, ok := address.(*net.IPNet); ok {
							item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
								Context: "global",
								Method:  sdp.RequestMethod_GET,
								Query:   net.JoinHostPort(ipNet.IP.String(), fmt.Sprint(socket.LocalAddr.Port)),
								Type:    "networksocket",
							})
						}
					}
				}
			} else {
				item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
					Context: "global",
					Method:  sdp.RequestMethod_GET,
					Query:   net.JoinHostPort(socket.LocalAddr.IP.String(), fmt.Sprint(socket.LocalAddr.Port)),
					Type:    "networksocket",
				})
			}
		} else {
			return nil, fmt.Errorf("socket with UID %v does not have an associated localAddress. Cannot determine port", socket.UID)
		}

	}

	attributes["localIPs"] = localIPs

	item.Attributes, err = sdp.ToAttributes(attributes)

	return &item, err
}
