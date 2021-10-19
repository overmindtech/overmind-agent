package dpkg

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
)

// QueryFormat is used when querying the dpkg database
const QueryFormat = `${Package}\n${db:Status-Status}\n${Priority}\n${Section}\n${Installed-Size}\n${Maintainer}\n${Architecture}\n${Version}\n${Homepage}\n${binary:Summary}\n${Description}\n---DEVIANTEND---\n`

// Package represents information about an dpkg package
type Package struct {
	Name          string
	Status        string
	Priority      string
	Section       string
	InstalledSize int
	Maintainer    string
	Architecture  string
	Version       string
	Homepage      *url.URL
	Summary       string
	Description   string
}

// NotFoundError will be returned when the dpkg wasn't installed
type NotFoundError struct {
	Message string
}

func (m NotFoundError) Error() string {
	return fmt.Sprintf("Package not found. Message: %v", m.Message)
}

// Supported returns true if the current system supports dpkg
func Supported() bool {
	_, err := exec.LookPath("dpkg-query")

	return err == nil
}

// ShowAll shows all packages
func ShowAll() ([]Package, error) {
	var command *exec.Cmd
	var err error
	var output []byte
	var packages []Package
	var args []string

	if !Supported() {
		return nil, errors.New("dpkg-query command not found")
	}

	args = []string{"-f", QueryFormat, "--show"}
	command = exec.Command("dpkg-query", args...)

	// Run the command
	output, err = command.Output()

	if err != nil {
		return nil, fmt.Errorf(
			"Error running command `dpkg-query %v`: %v\nOutput: %v",
			strings.Join(args, " "),
			err,
			string(output),
		)
	}

	packages, err = parseDpkgOutput(string(output))

	if err != nil {
		return nil, err
	}

	return packages, nil
}

// Show Gets info for one specific package
func Show(name string) (Package, error) {
	var command *exec.Cmd
	var err error
	var output []byte
	var packages []Package

	if !Supported() {
		return Package{}, errors.New("dpkg-query command not found")
	}

	args := []string{"-f", QueryFormat, "--show", name}
	command = exec.Command("dpkg-query", args...)

	// Run the command
	output, err = command.Output()

	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			// If the package just wasn't found
			if strings.Contains(string(e.Stderr), "no packages found") {
				return Package{}, NotFoundError{
					Message: string(output),
				}
			}

			return Package{}, fmt.Errorf(
				"Error running command `dpkg-query %v`: %v\nOutput: %v",
				strings.Join(args, " "),
				err,
				string(append(output, e.Stderr...)),
			)
		}

		return Package{}, err
	}

	packages, err = parseDpkgOutput(string(output))

	if err != nil {
		return Package{}, err
	}

	if len(packages) != 1 {
		return Package{}, fmt.Errorf("found >1 package for name %v: %v", name, packages)
	}

	return packages[0], nil
}

// Search searches for what packages provides a given file or filename-search-pattern
func Search(pattern string) ([]Package, error) {
	var searchCommand *exec.Cmd
	var searchOutput []byte
	var searchArgs []string
	var err error
	var packages []Package

	if !Supported() {
		return nil, errors.New("dpkg-query command not found")
	}

	searchArgs = []string{"--search", pattern}
	searchCommand = exec.Command("dpkg-query", searchArgs...)

	// Run the command
	searchOutput, err = searchCommand.Output()

	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			// If the package just wasn't found
			if strings.Contains(string(e.Stderr), "no path found") {
				return nil, NotFoundError{
					Message: string(searchOutput),
				}
			}

			return nil, fmt.Errorf(
				"Error running command `dpkg-query %v`: %v\nOutput: %v",
				strings.Join(searchArgs, " "),
				err,
				string(append(searchOutput, e.Stderr...)),
			)
		}

		return nil, err
	}

	// Parse the output and then find packages for each. The output should be as follows:
	//
	// login: /usr/bin/faillog
	// dpkg-dev: /usr/bin/dpkg-genchanges
	for _, line := range strings.Split(string(searchOutput), "\n") {
		// For each line extract the package name
		fields := strings.Split(line, ":")
		name := fields[0]

		// Get the package details
		pkg, err := Show(name)

		if err == nil {
			packages = append(packages, pkg)
		}
	}

	return packages, nil
}

func parseDpkgOutput(out string) ([]Package, error) {
	var packages []Package

	// Split the output for each package
	for _, infoString := range strings.Split(out, "---DEVIANTEND---\n") {
		// The final line could be "" therefore skip
		if infoString == "" {
			continue
		}

		pkg := Package{}

		// Split on lines
		lines := strings.Split(infoString, "\n")

		// Ensure that we have enough sections
		if len(lines) < 11 {
			return nil, fmt.Errorf("could not parse output, expected at least 11 lines, got %v. Output: %v", len(lines), lines)
		}

		// Extract data
		pkg.Name = lines[0]
		pkg.Status = lines[1]
		pkg.Priority = lines[2]
		pkg.Section = lines[3]

		if size, err := strconv.Atoi(lines[4]); err == nil {
			pkg.InstalledSize = size
		}

		pkg.Maintainer = lines[5]
		pkg.Architecture = lines[6]
		pkg.Version = lines[7]

		if u, err := url.Parse(lines[8]); err == nil {
			pkg.Homepage = u
		}

		pkg.Summary = lines[9]

		// The description that dpkg returns appears to be the summary on the
		// first line, then the description on the other lines, indented by some
		// number of spaces. We will need to trim off the summary then trim off
		// the indentation and re-join to a single string
		descLines := lines[10:]

		// Make sure we actually have description before doing anything drastic
		if len(descLines) > 1 {
			// Trim off the first line as this is actually the summary
			descLines = descLines[1:]

			// Trim off the whitespace
			for i, line := range descLines {
				descLines[i] = strings.TrimSpace(line)
			}

			pkg.Description = strings.Join(descLines, "\n")

		}

		packages = append(packages, pkg)
	}

	return packages, nil

}
