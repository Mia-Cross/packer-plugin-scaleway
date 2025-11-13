package cleanup

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*SnapshotCleanup)(nil)

type SnapshotCleanup struct {
	zone         scw.Zone
	snapshotName string
}

func Snapshot(zone scw.Zone, name string) *SnapshotCleanup {
	return &SnapshotCleanup{
		zone:         zone,
		snapshotName: name,
	}
}

func (i *SnapshotCleanup) Cleanup(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := instance.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&instance.ListSnapshotsRequest{
		Name:    &i.snapshotName,
		Zone:    i.zone,
		Project: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(resp.Snapshots) == 0 {
		return fmt.Errorf("could not find any snapshot by the name %q", i.snapshotName)
	}

	err = api.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		Zone:       i.zone,
		SnapshotID: resp.Snapshots[0].ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}
