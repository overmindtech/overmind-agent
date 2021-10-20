package sources

import "github.com/dylanratcliffe/deviant-agent/sources/unix"

func init() {
	Sources = append(Sources, &unix.FileSource{})
}
