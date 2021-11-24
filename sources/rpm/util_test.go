package rpm

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// HasPackages Returns whether this system has any packages on it. Since GitHub
// actions runners are Ubuntu but do have RPM installed you can end up in a
// situation where you do have RPM but there isn't a single RPM installed which
// can make the tests freak out since they are expecting to be able to find
// packages
func HasPackages() bool {
	command := exec.Command("rpm", "-qa")
	b, _ := command.Output()

	lines := strings.Split(string(b), "\n")

	numLines := len(lines) - 1

	return numLines != 0
}

func TestQueryAll(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	ps, err := QueryAll(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	command := exec.Command("rpm", "-qa")
	b, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(b), "\n")

	// Check that we have the correct number of packages
	t.Run("Number of packages", func(t *testing.T) {
		if len(ps) != (len(lines) - 1) {
			t.Fatalf("Mismatched number of results, QueryAll: %v rpm -qa: %v", len(ps), len(lines))
		}
	})

	for _, pkgName := range lines {
		t.Run(pkgName, func(t *testing.T) {
			if pkgName != "" {
				_, err := Query(context.Background(), pkgName)

				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestQuery(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	if !HasPackages() {
		t.Skip("RPM Based system with no packages, skipping test")
	}

	// Get a package for testing
	ps, _ := QueryAll(context.Background())

	packageName := ps[len(ps)-1].Name

	p, err := Query(context.Background(), packageName)

	if err != nil {
		t.Fatal(err)
	}

	// Test against what `rpm -qi` tells us
	command := exec.Command("rpm", "-qi", packageName)
	b, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	out := string(b)

	// Check that all of the basic attributes were present
	if strings.Contains(out, p.Name) == false {
		t.Errorf("Could not find reported %v Name (%v) in info: %v", packageName, p.Name, out)
	}
	if strings.Contains(out, p.Version) == false {
		t.Errorf("Could not find reported %v Version (%v) in info: %v", packageName, p.Version, out)
	}
	if strings.Contains(out, p.Release) == false {
		t.Errorf("Could not find reported %v Release (%v) in info: %v", packageName, p.Release, out)
	}
	if strings.Contains(out, p.Arch) == false {
		t.Errorf("Could not find reported %v Arch (%v) in info: %v", packageName, p.Arch, out)
	}
	if strings.Contains(out, p.Group) == false {
		t.Errorf("Could not find reported %v Group (%v) in info: %v", packageName, p.Group, out)
	}
	if strings.Contains(out, p.License) == false {
		t.Errorf("Could not find reported %v License (%v) in info: %v", packageName, p.License, out)
	}
	if strings.Contains(out, p.SourceRPM) == false {
		t.Errorf("Could not find reported %v SourceRPM (%v) in info: %v", packageName, p.SourceRPM, out)
	}
	if strings.Contains(out, p.Vendor) == false {
		t.Errorf("Could not find reported %v Vendor (%v) in info: %v", packageName, p.Vendor, out)
	}
	if strings.Contains(out, p.Summary) == false {
		t.Errorf("Could not find reported %v Summary (%v) in info: %v", packageName, p.Summary, out)
	}
	if strings.Contains(out, p.Description) == false {
		t.Errorf("Could not find reported %v Description (%v) in info: %v", packageName, p.Description, out)
	}
}

func TestEmptyQuery(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	// Get a package for testing
	_, err := Query(context.Background(), "thisShouldNotExist")

	_, ok := err.(NotFoundError)

	if ok == false {
		t.Errorf("Expected NotFoundError, got %v", err)
	}
}

func TestWhatProvides(t *testing.T) {
	if !Supported() {
		t.Skip("Non-RPM based system, skipping tests")
	}

	if !HasPackages() {
		t.Skip("RPM Based system with no packages, skipping test")
	}

	// Test a query that should work
	packages, err := WhatProvides(context.Background(), "bash")

	if err != nil {
		t.Fatalf("Error getting what provides bash: %v", err)
	}

	if len(packages) != 1 {
		t.Fatalf("Expected 1 package, got %v", len(packages))
	}

	bash := packages[0]

	if bash.Name != "bash" {
		t.Fatalf("Expected bash package name to be bash, got %v", bash.Name)
	}

	// Test a query that shouldn't work
	_, err = WhatProvides(context.Background(), "notarealcommand")

	if _, ok := err.(NotFoundError); ok == false {
		t.Fatalf("Expected NotFoundError, got %T: %v", err, err)
	}

	// Test a bad query
	_, err = WhatProvides(context.Background(), "/var/foo/bad/path.txt")

	if err == nil {
		t.Fatal("Found no error with bad query")
	}
}
