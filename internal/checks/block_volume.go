package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

//var _ VolumeCheck = &BlockVolumeCheck{}

type BlockVolumeCheck struct {
	zone      scw.Zone
	imageName string

	//volumeType string   // no need because if we call BlockVolumeCheck we know it's a block volume
	volumeName *string
	//tags       []string // no need because tags are not used in block volumes
	size *scw.Size
	iops *uint32
}

func (c *BlockVolumeCheck) SizeInGB(size uint64) *BlockVolumeCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *BlockVolumeCheck) IOPS(iops uint32) *BlockVolumeCheck {
	c.iops = &iops

	return c
}

func (c *BlockVolumeCheck) Name(volumeName string) *BlockVolumeCheck {
	c.volumeName = &volumeName

	return c
}

//func (c *BlockVolumeCheck) Tags(tags []string) *BlockVolumeCheck {
//	c.tags = tags
//
//	return c
//}

func findBlockVolume(ctx context.Context, zone scw.Zone, imageName string) (*block.Volume, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&block.ListVolumesRequest{
		Zone:      zone,
		Name:      &imageName,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list block volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("block volume with prefix %q not found, no volumes found", imageName)
	}

	if len(resp.Volumes) > 1 {
		return nil, fmt.Errorf("multiple block volumes found with name %s", imageName)
	}

	return resp.Volumes[0], nil
}

func (c *BlockVolumeCheck) Check(ctx context.Context) error {
	volume, err := findBlockVolume(ctx, c.zone, c.imageName)
	if err != nil {
		return err
	}

	if c.size != nil && volume.Size != *c.size {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	if c.iops != nil && volume.Specs != nil && volume.Specs.PerfIops != nil && *volume.Specs.PerfIops != *c.iops {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *c.size)
	}

	if c.volumeName != nil && volume.Name != *c.volumeName {
		return fmt.Errorf("volume name %s does not match expected volume name %s", volume.Name, c.volumeName)
	}

	//if len(c.tags) > 0 {
	//	for _, expectedTag := range c.tags {
	//		found := false
	//
	//		for _, actualTag := range volume.Tags {
	//			if actualTag == expectedTag {
	//				found = true
	//
	//				break
	//			}
	//		}
	//
	//		if !found {
	//			return fmt.Errorf("expected tag %q not found on volume %s", expectedTag, c.imageName)
	//		}
	//	}
	//}

	return nil
}

func BlockVolume(zone scw.Zone, imageName string) *BlockVolumeCheck {
	return &BlockVolumeCheck{
		zone:      zone,
		imageName: imageName,
		//volumeType: "sbs_volume",
	}
}
