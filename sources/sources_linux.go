package sources

import (
	"github.com/overmindtech/overmind-agent/sources/netstat"
	"github.com/overmindtech/overmind-agent/sources/systemd"
	"github.com/overmindtech/overmind-agent/sources/unix"
)

func init() {
	Sources = append(Sources, &unix.FileSource{})

	netstatSource := netstat.PortSource{}

	if netstatSource.Supported() {
		Sources = append(Sources, &netstatSource)
	}

	systemdSource := systemd.ServiceSource{}

	if systemdSource.Supported() {
		Sources = append(Sources, &systemdSource)
	}
}
