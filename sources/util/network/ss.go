package network

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
)

// Port represents the details of a port
type Port struct {
	Number int
	Listen string
	PID    int
}

// Compile the regex for looking at ports once only then store in memory
var tcpRegex = regexp.MustCompile(`(?m)^\d+\s+\d+\s+(?P<listenAddress>.*?):(?P<port>\d+)\s.*?pid=(?P<pid>\d+)`)

func ssCmd(mode string) []string {
	switch mode {
	case "tcp":
		return []string{"ss", "--tcp", "--listening", "--numeric", "--no-header", "--process", "--all", "state", "listening"}
	}

	// Return an empty string if nothing is  matched
	return []string{}
}

// Supported tells us if Puppet works on the current machine
func Supported() bool {
	_, err := exec.LookPath("ss")
	return err == nil
}

// ListeningPorts returns the TCP ports that are listingin on the system
func ListeningPorts() []Port {
	var ports []Port
	// tcpRegex := regexp.MustCompile(`(?m)^\d+\s+\d+\s+(?P<listenAddress>.*?):(?P<port>\d+)\s.*?pid=(?P<pid>\d+)`)

	// Run the ss command
	out := string(util.Run(ssCmd("tcp")))

	// Match the output
	caps := util.RegexpNamedCaptures(tcpRegex, out)

	// Convert to []Port
	for _, details := range caps {
		// Convert integers
		number, err := strconv.Atoi(details["port"])

		if err != nil {
			panic(fmt.Sprintf("Could not convert port details to integers: %v. Capture: %v", details["port"], details))
		}

		pid, err := strconv.Atoi(details["pid"])

		if err != nil {
			panic(fmt.Sprintf("Could not convert port details to integers: %v. Capture: %v", details["pid"], details))
		}

		ports = append(ports, Port{
			Number: number,
			Listen: details["listenAddress"],
			PID:    pid,
		})
	}
	return ports
}
