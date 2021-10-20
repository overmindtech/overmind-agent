package sources

import (
	"github.com/dylanratcliffe/deviant-agent/sources/netstat"
	"github.com/dylanratcliffe/deviant-agent/sources/systemd"
)

func init() {
	netstatSource := netstat.PortSource{}

	if netstatSource.Supported() {
		Sources = append(Sources, &netstatSource)
	}

	systemdSource := systemd.ServiceSource{}

	if systemdSource.Supported() {
		Sources = append(Sources, &systemdSource)
	}
}
