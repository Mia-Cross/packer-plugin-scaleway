package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*InstanceVolumeCheck)(nil)

type InstanceVolumeCheck struct {
	zone      scw.Zone
	imageName string
	tags      []string

	//volumeType string
	size *scw.Size
}

func (i *InstanceVolumeCheck) SizeInGB(size uint64) *InstanceVolumeCheck {
	i.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return i
}

func (i *InstanceVolumeCheck) SetExpectedTags(tags []string) *InstanceVolumeCheck {
	i.tags = tags

	return i
}

func findInstanceVolume(ctx context.Context, zone scw.Zone, imageName string) (*instance.Volume, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListVolumes(&instance.ListVolumesRequest{
		Zone:    zone,
		Name:    &imageName,
		Project: &testCtx.ProjectID,
	}, scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list instance volumes: %w", err)
	}

	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("instance volume with prefix %q not found, no volumes found", imageName)
	}

	if len(resp.Volumes) > 1 {
		return nil, fmt.Errorf("multiple instance volumes found with name %s", imageName)
	}

	return resp.Volumes[0], nil
}

func (i *InstanceVolumeCheck) Check(ctx context.Context) error {
	volume, err := findInstanceVolume(ctx, i.zone, i.imageName)
	if err != nil {
		return err
	}

	//if volume.Name != i.imageName {
	//	return fmt.Errorf("volume name %s does not match expected volume name %s", volume.Name, i.imageName)
	//}

	if i.size != nil && volume.Size != *i.size {
		return fmt.Errorf("volume size %d does not match expected size %d", volume.Size, *i.size)
	}

	if len(i.tags) > 0 {
		for _, expectedTag := range i.tags {
			found := false

			for _, actualTag := range volume.Tags {
				if actualTag == expectedTag {
					found = true

					break
				}
			}

			if !found {
				return fmt.Errorf("expected tag %q not found on volume %s", expectedTag, i.imageName)
			}
		}
	}

	return nil
}

func InstanceVolume(zone scw.Zone, name string) *InstanceVolumeCheck {
	return &InstanceVolumeCheck{
		zone:      zone,
		imageName: name,
	}
}
