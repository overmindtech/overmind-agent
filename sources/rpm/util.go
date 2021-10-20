package rpm

import (
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// QueryFormat is used when querying the RPM database
const QueryFormat = `%{NAME}\n%{EPOCH}\n%{VERSION}\n%{RELEASE}\n%{ARCH}\n%{INSTALLTIME}\n%{GROUP}\n%{SIZE}\n%{LICENSE}\n%{SOURCERPM}\n%{VENDOR}\n%{URL}\n%{SUMMARY}\n%{DESCRIPTION}\n---DEVIANTEND---\n`

// Package represents information about an RPM package
type Package struct {
	Name        string
	Epoch       int32
	Version     string
	Release     string
	Arch        string
	InstallTime time.Time
	Group       string
	Size        int32
	License     string
	SourceRPM   string
	Vendor      string
	URL         *url.URL
	Summary     string
	Description string
}

// NotFoundError will be returned when the RPM wasn't installed
type NotFoundError struct {
	Message string
}

func (m NotFoundError) Error() string {
	return fmt.Sprintf("Package not found. Message: %v", m.Message)
}

// Supported returns true if the current system supports RPM
func Supported() bool {
	_, err := exec.LookPath("rpm")

	return err == nil
}

// QueryAll Runs `rpm -qa`
func QueryAll() ([]Package, error) {
	var command *exec.Cmd
	var err error
	var output []byte
	var packages []Package
	var args []string

	if !Supported() {
		return nil, errors.New("RPM command not found")
	}

	args = []string{"-qa", "--queryformat", fmt.Sprintf("%v", QueryFormat)}
	command = exec.Command("rpm", args...)

	// Run the command
	output, err = command.Output()

	if err != nil {
		return nil, fmt.Errorf(
			"Error running command `rpm %v`: %v\nOutput: %v",
			strings.Join(args, " "),
			err,
			string(output),
		)
	}

	packages, err = parseRPMOutput(string(output))

	if err != nil {
		return nil, err
	}

	return packages, nil
}

// Query Gets info for one specific package
func Query(name string) (Package, error) {
	var command *exec.Cmd
	var err error
	var output []byte
	var packages []Package

	if !Supported() {
		return Package{}, errors.New("RPM command not found")
	}

	args := []string{"-q", name, "--queryformat", fmt.Sprintf("%v", QueryFormat)}
	command = exec.Command("rpm", args...)

	// Run the command
	output, err = command.Output()

	if err != nil {
		// If the package just wasn't found
		if strings.Contains(string(output), "not installed") {
			return Package{}, NotFoundError{
				Message: string(output),
			}
		}

		return Package{}, fmt.Errorf(
			"Error running command `rpm %v`: %v\nOutput: %v",
			strings.Join(args, " "),
			err,
			string(output),
		)
	}

	packages, err = parseRPMOutput(string(output))

	if err != nil {
		return Package{}, err
	}

	if len(packages) != 1 {
		return Package{}, fmt.Errorf("found >1 package for name %v: %v", name, packages)
	}

	return packages[0], nil
}

// WhatProvides runs rpm -q --whatprovides with the given query
func WhatProvides(query string) ([]Package, error) {
	var command *exec.Cmd
	var err error
	var output []byte

	if !Supported() {
		return nil, errors.New("RPM command not found")
	}

	args := []string{"-q", "--whatprovides", query, "--queryformat", fmt.Sprintf("%v", QueryFormat)}
	command = exec.Command("rpm", args...)

	// Run the command
	output, err = command.Output()

	if err != nil {
		// If the package just wasn't found
		if strings.Contains(string(output), "no package provides") || strings.Contains(string(output), "not owned by") {
			return nil, NotFoundError{
				Message: string(output),
			}
		}

		return nil, fmt.Errorf(
			"Error running command `rpm %v`: %v\nOutput: %v",
			strings.Join(args, " "),
			err,
			string(output),
		)
	}

	return parseRPMOutput(string(output))
}

func parseRPMOutput(out string) ([]Package, error) {
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
		if len(lines) < 14 {
			return nil, fmt.Errorf("could not parse output, expected at least 14 lines, got %v. Output: %v", len(lines), lines)
		}
		// Extract data
		pkg.Name = lines[0]

		if epoch, err := strconv.Atoi(lines[1]); err == nil {
			pkg.Epoch = int32(epoch)
		}

		pkg.Version = lines[2]
		pkg.Release = lines[3]
		pkg.Arch = lines[4]

		if unix, err := strconv.Atoi(lines[5]); err == nil {
			pkg.InstallTime = time.Unix(int64(unix), 0)
		}

		pkg.Group = lines[6]

		if size, err := strconv.Atoi(lines[7]); err == nil {
			pkg.Size = int32(size)
		}

		pkg.License = lines[8]
		pkg.SourceRPM = lines[9]
		pkg.Vendor = lines[10]

		if u, err := url.Parse(lines[11]); err == nil {
			pkg.URL = u
		}

		pkg.Summary = lines[12]

		// Description need to be re-joined
		pkg.Description = strings.Join(lines[13:], "\n")

		packages = append(packages, pkg)
	}

	return packages, nil
}
