package etcdata

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources"
	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
)

var testMountsBackend = &MountsSource{
	MountFunction: testMountFunc,
}

// testMountFunc Reads the test mount output instead of actually running mount
func testMountFunc() (*bufio.Scanner, error) {
	_, filename, _, _ := runtime.Caller(0)
	exampleFile := path.Join(path.Dir(filename), "test/mount")

	file, err := os.Open(exampleFile)

	if err != nil {
		return nil, err
	}

	// defer file.Close()

	return bufio.NewScanner(file), nil
}

var AllMounts = sources.ExpectedItems{
	NumItems: 20,
	ExpectedAttributes: []map[string]interface{}{
		{
			"path":   "/",
			"device": "/dev/disk0s2",
			"fstype": "hfs",
			"options": []interface{}{
				"local",
				"journaled",
			},
		},
		{
			"device": "devfs",
			"path":   "/dev",
			"fstype": "devfs",
			"options": []interface{}{
				"local",
				"nobrowse",
			},
		},
		{
			"device": "map -hosts",
			"path":   "/net",
			"fstype": "autofs",
			"options": []interface{}{
				"nosuid",
				"automounted",
				"nobrowse",
			},
		},
		{
			"device": "map auto_home",
			"path":   "/home",
			"fstype": "autofs",
			"options": []interface{}{
				"automounted",
				"nobrowse",
			},
		},
		{
			"device": "/dev/ad0s1a",
			"path":   "/",
			"fstype": "ufs",
			"options": []interface{}{
				"local",
				"soft-updates",
			},
		},
		{
			"device": "/dev/ad0s1d",
			"path":   "/ghost",
			"fstype": "ufs",
			"options": []interface{}{
				"local",
				"soft-updates",
			},
		},
		{
			"device": "tmpfs",
			"path":   "/run",
			"fstype": "tmpfs",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"seclabel",
				"mode=755",
			},
		},
		{
			"device": "sysfs",
			"path":   "/sys",
			"fstype": "sysfs",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"noexec",
				"relatime",
				"seclabel",
			},
		},
		{
			"device": "devtmpfs",
			"path":   "/dev",
			"fstype": "devtmpfs",
			"options": []interface{}{
				"rw",
				"nosuid",
				"seclabel",
				"size=395340k",
				"nr_inodes=98835",
				"mode=755",
			},
		},
		{
			"device": "tmpfs",
			"path":   "/sys/fs/cgroup",
			"fstype": "tmpfs",
			"options": []interface{}{
				"ro",
				"nosuid",
				"nodev",
				"noexec",
				"seclabel",
				"mode=755",
			},
		},
		{
			"device": "cgroup",
			"path":   "/sys/fs/cgroup/systemd",
			"fstype": "cgroup",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"noexec",
				"relatime",
				"seclabel",
				"xattr",
				"release_agent=/usr/lib/systemd/systemd-cgroups-agent",
				"name=systemd",
			},
		},
		{
			"device": "cgroup",
			"path":   "/sys/fs/cgroup/net_cls,net_prio",
			"fstype": "cgroup",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"noexec",
				"relatime",
				"seclabel",
				"net_cls",
				"net_prio",
			},
		},
		{
			"device": "/dev/mapper/cl_centos8-root",
			"path":   "/",
			"fstype": "xfs",
			"options": []interface{}{
				"rw",
				"relatime",
				"seclabel",
				"attr2",
				"inode64",
				"logbufs=8",
				"logbsize=32k",
				"noquota",
			},
		},
		{
			"device": "systemd-1",
			"path":   "/proc/sys/fs/binfmt_misc",
			"fstype": "autofs",
			"options": []interface{}{
				"rw",
				"relatime",
				"fd=35",
				"pgrp=1",
				"timeout=0",
				"minproto=5",
				"maxproto=5",
				"direct",
				"pipe_ino=13712",
			},
		},
		{
			"device": "/dev/ad0s1a",
			"path":   "/",
			"fstype": "ufs",
			"options": []interface{}{
				"local",
			},
		},
		{
			"device": "devfs",
			"path":   "/dev",
			"fstype": "devfs",
			"options": []interface{}{
				"local",
			},
		},
		{
			"device": "/dev/ad0s1e",
			"path":   "/tmp",
			"fstype": "ufs",
			"options": []interface{}{
				"local",
				"soft-updates",
			},
		},
		{
			"device": "tmpfs",
			"path":   "/run",
			"fstype": "tmpfs",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"seclabel",
				"mode=755",
			},
		},
		{
			"device": "/dev/wd0a",
			"path":   "/",
			"fstype": "ffs",
			"options": []interface{}{
				"local",
			},
		},
		{
			"device": "tmpfs",
			"path":   "/run",
			"fstype": "tmpfs",
			"options": []interface{}{
				"rw",
				"nosuid",
				"nodev",
				"seclabel",
				"mode=755",
			},
		},
	},
}

func TestMountsFind(t *testing.T) {
	tests := []sources.SourceTest{
		{
			Name:          "normal find",
			ItemContext:   util.LocalContext,
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &AllMounts,
		},
	}

	sources.RunSourceTests(t, tests, testMountsBackend)
}

func TestMountsSearch(t *testing.T) {
	slashMounts := make([]map[string]interface{}, 0)

	for _, expected := range AllMounts.ExpectedAttributes {
		if expected["path"] == "/" {
			slashMounts = append(slashMounts, expected)
		}
	}

	tests := []sources.SourceTest{
		{
			Name:          "* search",
			ItemContext:   util.LocalContext,
			Query:         "*",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &AllMounts,
		},
		{
			Name:          "/ search",
			ItemContext:   util.LocalContext,
			Query:         "/",
			Method:        sdp.RequestMethod_SEARCH,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems:           len(slashMounts),
				ExpectedAttributes: slashMounts,
			},
		},
	}

	sources.RunSourceTests(t, tests, testMountsBackend)
}

func TestMountsGet(t *testing.T) {
	// Deduplicate the expected mounts since the the behavior should be to
	// return the first one that matches. Normally you wouldn't be able to have
	// duplicates but in my test data we have them so we need to deduplicate
	// here
	expectedGetMounts := make(map[string]map[string]interface{})

	for _, expected := range AllMounts.ExpectedAttributes {
		_, exists := expectedGetMounts[fmt.Sprint(expected["path"])]
		if !exists {
			expectedGetMounts[fmt.Sprint(expected["path"])] = expected
		}
	}

	tests := make([]sources.SourceTest, 0)

	for _, expected := range expectedGetMounts {
		path := fmt.Sprint(expected["path"])
		tests = append(tests, sources.SourceTest{
			Name:          path,
			ItemContext:   util.LocalContext,
			Query:         path,
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &sources.ExpectedItems{
				NumItems: 1,
				ExpectedAttributes: []map[string]interface{}{
					expected,
				},
			},
		})
	}

	sources.RunSourceTests(t, tests, testMountsBackend)
}
