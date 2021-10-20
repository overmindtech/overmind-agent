package sources

import "github.com/dylanratcliffe/deviant-agent/sources/netstat"

func init() {
	netstatSource := netstat.PortSource{}

	if netstatSource.Supported() {
		Sources = append(Sources, &netstatSource)
	}
}
