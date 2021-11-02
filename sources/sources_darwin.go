package sources

import "github.com/overmindtech/overmind-agent/sources/unix"

func init() {
	Sources = append(Sources, &unix.FileSource{})
}
