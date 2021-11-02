package sources

import "github.com/overmindtech/overmind-agent/sources/netstat"

func init() {
	netstatSource := netstat.PortSource{}

	if netstatSource.Supported() {
		Sources = append(Sources, &netstatSource)
	}
}
