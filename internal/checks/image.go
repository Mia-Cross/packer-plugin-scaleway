package checks

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCheck = (*ImageCheck)(nil)

func Image(zone scw.Zone, name string) *ImageCheck {
	return &ImageCheck{
		zone:      zone,
		imageName: name,
	}
}

type ImageCheck struct {
	zone      scw.Zone
	imageName string
	tags      []string

	size *scw.Size
	//rootVolumeType        *string
	rootVolumeSnapshot SnapshotCheck
	//extraVolumesTypes     map[string]string
	extraVolumesSnapshots map[string]SnapshotCheck
}

//func (c *ImageCheck) RootVolumeType(rootVolumeType string) *ImageCheck {
//	c.rootVolumeType = &rootVolumeType
//
//	return c
//}

//func (c *ImageCheck) RootVolumeBlockSnapshot(snapshotCheck *BlockSnapshotCheck) *ImageCheck {
//	c.rootVolumeSnapshot = snapshotCheck
//
//	return c
//}

func (c *ImageCheck) RootVolumeSnapshot(snapshotCheck SnapshotCheck) *ImageCheck {
	c.rootVolumeSnapshot = snapshotCheck

	return c
}

//func (c *ImageCheck) ExtraVolumeType(index string, extraVolumeType string) *ImageCheck {
//	//if c.extraVolumesSnapshots == nil {
//	c.extraVolumesTypes = map[string]string{index: extraVolumeType}
//	//}
//
//	//c.extraVolumesSnapshots = append()
//	return c
//}

func (c *ImageCheck) ExtraVolumeSnapshot(index string, snapshotCheck SnapshotCheck) *ImageCheck {
	//if c.extraVolumesSnapshots == nil {
	c.extraVolumesSnapshots = map[string]SnapshotCheck{index: snapshotCheck}
	//}

	//c.extraVolumesSnapshots = append()
	return c
}

//func (c *ImageCheck) ExtraVolumeSize(key string, volumeSize uint64) *ImageCheck {
//	if c.extraVolumesSnapshots == nil {
//		c.extraVolumesSnapshots = map[string]SnapshotCheck{}
//	}
//
//	c.extraVolumesSnapshots[key].SetExpectedSizeInGB(volumeSize)
//
//	return c
//}
//
//func (c *ImageCheck) ExtraVolumeType(key string, volumeType string) *ImageCheck {
//	if c.extraVolumesSnapshots == nil {
//		c.extraVolumesSnapshots = map[string]SnapshotCheck{}
//	}
//
//	c.extraVolumesSnapshots[key].SetExpectedType(volumeType)
//
//	return c
//}

func (c *ImageCheck) SizeInGB(size uint64) *ImageCheck {
	c.size = scw.SizePtr(scw.Size(size) * scw.GB)

	return c
}

func (c *ImageCheck) Tags(tags []string) *ImageCheck {
	c.tags = tags

	return c
}

func findImage(ctx context.Context, zone scw.Zone, imageName string) (*instance.Image, error) {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)
	//images := []*instance.Image(nil)

	resp, err := api.ListImages(&instance.ListImagesRequest{
		Name:    &imageName,
		Zone:    zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	//for _, img := range resp.Images {
	//	if img.Name == imageName {
	//		images = append(images, img)
	//	}
	//}

	if len(resp.Images) == 0 {
		return nil, fmt.Errorf("image %s not found, no images found", imageName)
	}

	if len(resp.Images) > 1 {
		return nil, fmt.Errorf("multiple images found with name %s", imageName)
	}

	return resp.Images[0], nil
}

func (c *ImageCheck) Check(ctx context.Context) error {
	image, err := findImage(ctx, c.zone, c.imageName)
	if err != nil {
		return err
	}

	if image.Name != c.imageName {
		return fmt.Errorf("image name %s does not match expected %s", image.Name, c.imageName)
	}

	//if c.rootVolumeType != nil && string(image.RootVolume.VolumeType) != *c.rootVolumeType {
	//	return fmt.Errorf("image root volume type %s does not match expected %s", image.RootVolume.VolumeType, *c.rootVolumeType)
	//}
	if c.rootVolumeSnapshot != nil {
		err = c.rootVolumeSnapshot.Check(ctx)
		if err != nil {
			return err
		}
	}

	if c.size != nil && image.RootVolume.Size != *c.size {
		return fmt.Errorf("image size %d does not match expected %d", image.RootVolume.Size, *c.size)
	}

	if c.extraVolumesSnapshots != nil {
		for key, snapshotCheck := range c.extraVolumesSnapshots {
			_, exists := image.ExtraVolumes[key]

			if !exists {
				return fmt.Errorf("extra volume %s does not exist", key)
			}

			volumeErr := snapshotCheck.Check(ctx)
			if volumeErr != nil {
				return fmt.Errorf("extra volume %s check failed: %w", key, volumeErr)
				//if string(vol.VolumeType) != volumeCheck. {
				//	return fmt.Errorf("extra volume %s type %s does not match expected %s", key, vol.VolumeType, v)
			}
		}
	}

	if c.tags != nil {
		for _, expectedTag := range c.tags {
			found := false

			for _, actualTag := range image.Tags {
				if actualTag == expectedTag {
					found = true

					break
				}
			}

			if !found {
				return fmt.Errorf("expected tag %q not found on image %s", expectedTag, c.imageName)
			}
		}
	}

	return nil
}
