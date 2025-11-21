package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

type InstanceSnapshotCheck struct {
	zone      scw.Zone
	imageName string

	//tags         []string
	size *scw.Size
}

var _ SnapshotCheck = &InstanceSnapshotCheck{}

func (c *InstanceSnapshotCheck) SizeInGB(size uint64) *InstanceSnapshotCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

//func (c *InstanceSnapshotCheck) Tags(tags []string) *InstanceSnapshotCheck {
//	c.tags = tags
//
//	return c
//}

func findInstanceSnapshot(ctx context.Context, zone scw.Zone, imageName string) (*instance.Snapshot, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&instance.ListSnapshotsRequest{
		Zone:    zone,
		Name:    &imageName,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("error listing instance snapshots: %v", err)
	}

	if len(resp.Snapshots) == 0 {
		return nil, fmt.Errorf("could not find any instance snapshot prefixed with %q", imageName)
	}

	if len(resp.Snapshots) > 1 {
		return nil, fmt.Errorf("multiple instance snapshots found with prefix %q", imageName)
	}

	return resp.Snapshots[0], nil
}

func (c *InstanceSnapshotCheck) Check(ctx context.Context) error {
	snapshot, err := findInstanceSnapshot(ctx, c.zone, c.imageName)
	if err != nil {
		return err
	}

	//if snapshot.Name != c.snapshotName {
	//	return fmt.Errorf("snapshot name %s does not match expected snapshot name %s", snapshot.Name, c.snapshotName)
	//}

	if c.size != nil && snapshot.Size != *c.size {
		return fmt.Errorf("snapshot size %d does not match expected size %d", snapshot.Size, *c.size)
	}

	//if len(c.tags) > 0 {
	//	for _, expectedTag := range c.tags {
	//		found := false
	//
	//		for _, actualTag := range snapshot.Tags {
	//			if actualTag == expectedTag {
	//				found = true
	//
	//				break
	//			}
	//		}
	//
	//		if !found {
	//			return fmt.Errorf("expected tag %q not found on snapshot %s", expectedTag, c.snapshotName)
	//		}
	//	}
	//}

	return nil
}

func InstanceSnapshot(zone scw.Zone, imageName string) *InstanceSnapshotCheck {
	return &InstanceSnapshotCheck{
		zone:      zone,
		imageName: imageName,
		//snapshotName: snapshotName,
	}
}
