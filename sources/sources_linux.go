package sources

import (
	"github.com/dylanratcliffe/deviant-agent/sources/netstat"
	"github.com/dylanratcliffe/deviant-agent/sources/systemd"
	"github.com/dylanratcliffe/deviant-agent/sources/unix"
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
