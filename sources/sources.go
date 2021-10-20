package sources

import (
	"github.com/dylanratcliffe/deviant-agent/sources/dpkg"
	"github.com/dylanratcliffe/deviant-agent/sources/etcdata"
	"github.com/dylanratcliffe/discovery"
)

var Sources []discovery.Source

// Load sources that are abe to compile on all operating systems, burt check
// that they are supported before actually loading them
func init() {
	if dpkg.Supported() {
		Sources = append(Sources, &dpkg.DpkgSource{})
	}

	groupsSource := etcdata.GroupsSource{}

	if groupsSource.Supported() {
		Sources = append(Sources, &groupsSource)
	}

	Sources = append(Sources, &etcdata.HostsSource{})

	mountsSource := etcdata.MountsSource{}

	if mountsSource.Supported() {
		Sources = append(Sources, &mountsSource)
	}

	usersSource := etcdata.UsersSource{}

	if usersSource.Supported() {
		Sources = append(Sources, &usersSource)
	}
}
