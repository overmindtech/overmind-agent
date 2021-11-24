package psutil

import (
	"context"
	"fmt"

	"github.com/shirou/gopsutil/disk"

	"github.com/overmindtech/overmind-agent/sources/util"
	"github.com/overmindtech/sdp-go"
)

type DiskSource struct{}

// Type is the type of items that this returns (Required)
func (s *DiskSource) Type() string {
	return "disk"
}

// Descriptive name for the source, used in logging and metadata
func (s *DiskSource) Name() string {
	return "psutil"
}

// Weighting of duplicate sources
func (s *DiskSource) Weight() int {
	return 100
}

// List of contexts that this source is capable of find items for
func (s *DiskSource) Contexts() []string {
	return []string{
		util.LocalContext,
	}
}

// Get information about a disk. device = path
func (s *DiskSource) Get(ctx context.Context, itemContext string, query string) (*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	diskPartitions, err := disk.PartitionsWithContext(ctx, true)

	if err != nil {
		// There was an error getting the disk partitions
		return &sdp.Item{}, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	// Let's check if device is in the diskPartitions slice

	for _, p := range diskPartitions {
		if p.Device == query {
			return getDiskInformation(ctx, p)
		}
	}

	// If it gets to this line it means that there was no matching disk
	return &sdp.Item{}, &sdp.ItemRequestError{
		ErrorType:   sdp.ItemRequestError_NOTFOUND,
		ErrorString: fmt.Sprintf("Disk %v not found among available disks %v", query, diskPartitions),
		Context:     itemContext,
	}
}

// Find information about all of the disks.
func (s *DiskSource) Find(ctx context.Context, itemContext string) ([]*sdp.Item, error) {
	if itemContext != util.LocalContext {
		return nil, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_NOCONTEXT,
			ErrorString: fmt.Sprintf("context %v not available, local context is %v", itemContext, util.LocalContext),
			Context:     itemContext,
		}
	}

	var items []*sdp.Item

	items = make([]*sdp.Item, 0)

	diskPartitions, err := disk.PartitionsWithContext(ctx, false)

	if err != nil {
		// There was an error getting the disk partitions
		return items, &sdp.ItemRequestError{
			ErrorType:   sdp.ItemRequestError_OTHER,
			ErrorString: err.Error(),
			Context:     itemContext,
		}
	}

	for _, partition := range diskPartitions {
		item, err := getDiskInformation(ctx, partition)

		if err == nil {
			items = append(items, item)
		} else {
			return items, &sdp.ItemRequestError{
				ErrorType:   sdp.ItemRequestError_OTHER,
				ErrorString: err.Error(),
				Context:     itemContext,
			}
		}
	}

	return items, nil
}

func getDiskInformation(ctx context.Context, partition disk.PartitionStat) (*sdp.Item, error) {
	var err error

	attributes := make(map[string]interface{})

	item := sdp.Item{}

	item.Type = "disk"
	item.UniqueAttribute = "device"
	item.Context = util.LocalContext

	attributes["device"] = partition.Device
	attributes["mountpoint"] = partition.Mountpoint
	attributes["fstype"] = partition.Fstype
	attributes["opts"] = partition.Opts

	item.LinkedItemRequests = []*sdp.ItemRequest{
		{
			Type:    "file",
			Query:   partition.Mountpoint,
			Method:  sdp.RequestMethod_GET,
			Context: util.LocalContext,
		},
	}

	// Try to get statistics about the disk usage
	var diskUsage *disk.UsageStat

	diskUsage, err = disk.UsageWithContext(ctx, partition.Mountpoint)

	// If there was not error then go ahead
	if err == nil {
		attributes["path"] = diskUsage.Path
		attributes["total"] = diskUsage.Total
		attributes["free"] = diskUsage.Free
		attributes["used"] = diskUsage.Used
		attributes["usedPercent"] = diskUsage.UsedPercent
		attributes["inodesTotal"] = diskUsage.InodesTotal
		attributes["inodesUsed"] = diskUsage.InodesUsed
		attributes["inodesFree"] = diskUsage.InodesFree
		attributes["inodesUsedPercent"] = diskUsage.InodesUsedPercent
	}

	var counters map[string]disk.IOCountersStat

	// Try to get statistics about IO
	counters, err = disk.IOCountersWithContext(ctx, partition.Device)

	// If there was not error then go ahead
	if err == nil {
		var stats disk.IOCountersStat

		// Just get the last element from the map, there should only be one anyway
		for _, v := range counters {
			stats = v
		}

		attributes["readCount"] = stats.ReadCount
		attributes["mergedReadCount"] = stats.MergedReadCount
		attributes["writeCount"] = stats.WriteCount
		attributes["mergedWriteCount"] = stats.MergedWriteCount
		attributes["readBytes"] = stats.ReadBytes
		attributes["writeBytes"] = stats.WriteBytes
		attributes["readTime"] = stats.ReadTime
		attributes["writeTime"] = stats.WriteTime
		attributes["iopsInProgress"] = stats.IopsInProgress
		attributes["ioTime"] = stats.IoTime
		attributes["weightedIO"] = stats.WeightedIO
		attributes["name"] = stats.Name
		attributes["serialNumber"] = stats.SerialNumber
		attributes["label"] = stats.Label
	}

	item.Attributes, err = sdp.ToAttributes(attributes)

	return &item, err
}
