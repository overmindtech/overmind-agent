//go:build linux || windows
// +build linux windows

package netstat

import (
	"fmt"
	"runtime"
	"strconv"

	ns "github.com/cakturk/go-netstat/netstat"
	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
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
func (s *PortSource) Get(itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var err error
	var sockets []ns.SockTabEntry
	var socket ns.SockTabEntry
	var port int

	port, err = strconv.Atoi(query)

	if err != nil {
		return nil, fmt.Errorf("could not convert %v to uint16", query)
	}

	sockets, _ = ns.TCPSocks(acceptNumber(uint16(port)))

	switch numSockets := len(sockets); numSockets {
	case 0:
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOTFOUND,
			ErrorString: fmt.Sprintf("Port %v not found", query),
		}
	case 1:
		socket = sockets[0]
		return socketToItem(socket)
	default:
		// If more than one was found then we need to narrow it down to one. Thi
		// is possible because there might be many established connections on a
		// single port. In this case I'm going to return the first one that is
		// listening. If this was the search method then I would probably return
		// everything
		for _, sock := range sockets {
			if sock.State.String() == "LISTEN" {
				return socketToItem(sock)
			}
		}

		// If none are listening then fail
		return nil, fmt.Errorf("found many sockets on port %v, none listening. Get() expects to be able to find a listening port, if you want more results use Search()", query)
	}
}

// Find calls back to the FindFunction to find all items
func (s *PortSource) Find(itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var sockets []ns.SockTabEntry
	var items []*sdp.Item
	var err error

	sockets, err = ns.TCPSocks(acceptListening)

	if len(sockets) == 0 || err != nil {
		return make([]*sdp.Item, 0), err
	}

	for _, socket := range sockets {
		item, err := socketToItem(socket)

		if err == nil {
			items = append(items, item)
		} else {
			return items, err
		}
	}

	return items, nil
}

func acceptListening(s *ns.SockTabEntry) bool {
	return s.State.String() == "LISTEN"
}

func acceptNumber(p uint16) func(*ns.SockTabEntry) bool {
	return func(n *ns.SockTabEntry) bool {
		return n.LocalAddr.Port == p
	}
}

func socketToItem(s ns.SockTabEntry) (*sdp.Item, error) {
	var err error
	var item sdp.Item

	attributes := make(map[string]interface{})

	// Prepare the item
	item = sdp.Item{}
	item.Type = "port"
	item.UniqueAttribute = "port"
	item.Context = util.LocalContext
	item.LinkedItemRequests = make([]*sdp.ItemRequest, 0)

	attributes["state"] = s.State.String()

	// Protect against nil pointer dereference
	if s.LocalAddr != nil {
		attributes["port"] = uint32(s.LocalAddr.Port)
		attributes["localIP"] = s.LocalAddr.IP.String()

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Context: "global",
			Method:  sdp.RequestMethod_GET,
			Query:   s.LocalAddr.IP.String(),
			Type:    "ip",
		})
	} else {
		// I'm pretty sure if there is no local address then the socker isn't a port so we should fail
		return nil, fmt.Errorf("socket with UID %v does not have an associated localAddress. Cannot determine port", s.UID)
	}

	if s.Process != nil {
		attributes["pid"] = s.Process.Pid

		item.LinkedItemRequests = append(item.LinkedItemRequests, &sdp.ItemRequest{
			Method: sdp.RequestMethod_GET,
			Query:  strconv.Itoa(s.Process.Pid),
			Type:   "process",
		})
	}

	if s.RemoteAddr != nil {
		if s.RemoteAddr.Port != 0 {
			attributes["remoteIP"] = s.RemoteAddr.IP.String()
			attributes["remotePort"] = s.RemoteAddr.Port
		}
	}

	item.Attributes, err = sdp.ToAttributes(attributes)

	return &item, err
}
