package sources

import (
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/overmind-agent/sources/command"
	"github.com/overmindtech/overmind-agent/sources/dpkg"
	"github.com/overmindtech/overmind-agent/sources/etcdata"
	"github.com/overmindtech/overmind-agent/sources/file_content"
	"github.com/overmindtech/overmind-agent/sources/psutil"
	"github.com/overmindtech/overmind-agent/sources/rpm"
	"github.com/overmindtech/overmind-agent/sources/system"
)

var Sources []discovery.Source

// Load sources that are abe to compile on all operating systems, burt check
// that they are supported before actually loading them
func init() {
	Sources = append(Sources, &etcdata.HostsSource{})
	Sources = append(Sources, &psutil.DiskSource{})
	Sources = append(Sources, &psutil.ProcessSource{})
	Sources = append(Sources, &system.SystemSource{})
	Sources = append(Sources, &file_content.FileContentSource{})
	Sources = append(Sources, &command.CommandSource{})

	dpkgSource := dpkg.DpkgSource{}

	if dpkgSource.Supported() {
		Sources = append(Sources, &dpkgSource)
	}

	rpmSource := rpm.RPMSource{}

	if rpmSource.Supported() {
		Sources = append(Sources, &rpmSource)
	}

	groupsSource := etcdata.GroupsSource{}

	if groupsSource.Supported() {
		Sources = append(Sources, &groupsSource)
	}

	mountsSource := etcdata.MountsSource{}

	if mountsSource.Supported() {
		Sources = append(Sources, &mountsSource)
	}

	usersSource := etcdata.UsersSource{}

	if usersSource.Supported() {
		Sources = append(Sources, &usersSource)
	}
}
