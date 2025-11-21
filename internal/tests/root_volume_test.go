package tests_test

import (
	"fmt"
	"testing"

	"github.com/scaleway/packer-plugin-scaleway/internal/checks"
	"github.com/scaleway/packer-plugin-scaleway/internal/cleanup"
	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// Test : Root volume local
// OK
func TestRootVolumeLocal(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-root-volume-local"
	rootVolumeType := "l_ssd"
	rootVolumeSize := 20

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "GP1-XS"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = "%s"
			  ssh_username = "root"
			  remove_volume = true
			
			  root_volume {
			    type = "%s"
			    size_in_gb = %d
			  }
			}
			
			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, rootVolumeType, rootVolumeSize),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				//RootVolumeType(rootVolumeType).
				SizeInGB(uint64(rootVolumeSize)).
				RootVolumeSnapshot(
					checks.InstanceSnapshot(zone, imageName).
						SizeInGB(uint64(rootVolumeSize)),
				),
			checks.NoVolume(zone),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.InstanceSnapshot(zone, imageName),
		},
	})
}

// Test : root volume sbs
// Almost OK, see TODO
func TestRootVolumeSBS(t *testing.T) {
	zone := scw.ZoneFrPar1
	imageName := "packer-e2e-root-volume-sbs"
	//rootVolumeType := "sbs_snapshot"
	rootVolumeSize := 50

	tester.Test(t, &tester.TestConfig{
		Config: fmt.Sprintf(`
			source "scaleway" "basic" {
			  communicator = "none"
			  commercial_type = "PLAY2-PICO"
			  zone = "%s"
			  image = "ubuntu_jammy"
			  image_name = %q
			  ssh_username = "root"
			  remove_volume = true

			  root_volume {
			    size_in_gb = %d
			    iops = 15000
			  }
			}

			build {
			  sources = ["source.scaleway.basic"]
			}
			`, zone, imageName, rootVolumeSize),
		Checks: []tester.PackerCheck{
			checks.Image(zone, imageName).
				//SizeInGB(uint64(rootVolumeSize)).  // TODO: fix size = 0
				RootVolumeSnapshot(
					checks.BlockSnapshot(zone, imageName).
						SizeInGB(uint64(rootVolumeSize)),
				),
			checks.NoVolume(zone),
		},
		Cleanup: []tester.PackerCleanup{
			cleanup.Image(zone, imageName),
			cleanup.BlockSnapshot(zone, imageName),
		},
	})
}
