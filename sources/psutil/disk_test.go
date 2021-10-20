package psutil

import (
	"testing"

	"github.com/dylanratcliffe/deviant-agent/sources/util"
	"github.com/dylanratcliffe/sdp-go"
	"github.com/shirou/gopsutil/disk"
)

// Basic test that just checks that the backend is able to get info about the
// first disk
func TestDiskGet(t *testing.T) {
	var err error
	var diskPartitions []disk.PartitionStat
	var diskPath string

	diskPartitions, err = disk.Partitions(false)

	if err != nil {
		t.Errorf("Error during test prep: %v", err)
	}

	diskPath = diskPartitions[0].Device

	tests := []util.SourceTest{
		{
			Name:          "get existing disk",
			ItemContext:   util.LocalContext,
			Query:         diskPath,
			Method:        sdp.RequestMethod_GET,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: 1,
			},
		},
		{
			Name:        "get bad disk",
			ItemContext: util.LocalContext,
			Query:       "/not-real",
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOTFOUND,
				Context: util.LocalContext,
			},
		},
		{
			Name:        "get bad context",
			ItemContext: "bad",
			Query:       diskPath,
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOCONTEXT,
				Context: "bad",
			},
		},
	}

	util.RunSourceTests(t, tests, &DiskSource{})
}

func TestDiskFind(t *testing.T) {
	var err error
	var diskPartitions []disk.PartitionStat
	var diskPath string

	diskPartitions, err = disk.Partitions(true)

	if err != nil {
		t.Errorf("Error during test prep: %v", err)
	}

	diskPath = diskPartitions[0].Device

	tests := []util.SourceTest{
		{
			Name:          "find all disks",
			ItemContext:   util.LocalContext,
			Method:        sdp.RequestMethod_FIND,
			ExpectedError: nil,
			ExpectedItems: &util.ExpectedItems{
				NumItems: len(diskPartitions),
			},
		},
		{
			Name:        "find bad context",
			ItemContext: "bad",
			Query:       diskPath,
			Method:      sdp.RequestMethod_GET,
			ExpectedError: &util.ExpectedError{
				Type:    sdp.ItemRequestError_NOCONTEXT,
				Context: "bad",
			},
		},
	}

	util.RunSourceTests(t, tests, &DiskSource{})
}
