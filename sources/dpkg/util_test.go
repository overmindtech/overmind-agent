package dpkg

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

// HasPackages Returns whether this system has any packages on it
func HasPackages() bool {
	command := exec.Command("dpkg-query", "-l")
	b, _ := command.Output()

	lines := strings.Split(string(b), "\n")

	numLines := len(lines) - 1

	return numLines != 0
}

func TestShowAll(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	ps, err := ShowAll()

	if err != nil {
		t.Fatal(err)
	}

	command := exec.Command("dpkg-query", "-f=${Package}\n", "--show")
	b, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(string(b), "\n")

	// Check that we have the correct number of packages
	t.Run("Number of packages", func(t *testing.T) {
		if len(ps) != (len(lines) - 1) {
			t.Fatalf("Mismatched number of results, ShowAll: %v rpm -qa: %v", len(ps), len(lines))
		}
	})

	for _, pkgName := range lines {
		t.Run(pkgName, func(t *testing.T) {
			if pkgName != "" {
				_, err := Show(pkgName)

				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestShow(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	if !HasPackages() {
		t.Skip("dpkg-query has no packages, skipping test")
	}

	// Get a package for testing
	ps, _ := ShowAll()

	packageName := ps[len(ps)-1].Name

	p, err := Show(packageName)

	if err != nil {
		t.Fatal(err)
	}

	// Test against what `rpm -qi` tells us
	command := exec.Command("dpkg-query", "-s", packageName)
	b, err := command.Output()

	if err != nil {
		t.Fatal(err)
	}

	out := string(b)

	// Check that all of the basic attributes were present
	if strings.Contains(out, p.Name) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Name, out)
	}
	if strings.Contains(out, p.Status) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Status, out)
	}
	if strings.Contains(out, p.Priority) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Priority, out)
	}
	if strings.Contains(out, p.Section) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Section, out)
	}
	if strings.Contains(out, fmt.Sprint(p.InstalledSize)) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.InstalledSize, out)
	}
	if strings.Contains(out, p.Maintainer) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Maintainer, out)
	}
	if strings.Contains(out, p.Architecture) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Architecture, out)
	}
	if strings.Contains(out, p.Version) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Version, out)
	}
	if strings.Contains(out, p.Homepage.String()) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Homepage, out)
	}
	if strings.Contains(out, p.Summary) == false {
		t.Errorf("Could not find reported %v name (%v) in info: %v", packageName, p.Summary, out)
	}
}

func TestEmptyShow(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	// Get a package for testing
	_, err := Show("thisShouldNotExist")

	_, ok := err.(NotFoundError)

	if ok == false {
		t.Errorf("Expected NotFoundError, got %v", err)
	}
}

func TestSearch(t *testing.T) {
	if !Supported() {
		t.Skip("dpkg-query not present, skipping tests")
	}

	if !HasPackages() {
		t.Skip("dpkg-query has no packages, skipping test")
	}

	// Test a query that should work
	packages, err := Search("/bin/bash")

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
	_, err = Search("notarealcommand")

	if _, ok := err.(NotFoundError); ok == false {
		t.Fatalf("Expected NotFoundError, got %T: %v", err, err)
	}

	// Test a bad query
	_, err = Search("/var/foo/bad/path.txt")

	if err == nil {
		t.Fatal("Found no error with bad query")
	}
}
