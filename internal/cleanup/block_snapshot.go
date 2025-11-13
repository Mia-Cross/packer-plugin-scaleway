package cleanup

import (
	"context"
	"fmt"

	"github.com/scaleway/packer-plugin-scaleway/internal/tester"
	"github.com/scaleway/scaleway-sdk-go/api/block/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

var _ tester.PackerCleanup = (*BlockSnapshotCleanup)(nil)

type BlockSnapshotCleanup struct {
	zone         scw.Zone
	snapshotName string
}

func BlockSnapshot(zone scw.Zone, name string) *BlockSnapshotCleanup {
	return &BlockSnapshotCleanup{
		zone:         zone,
		snapshotName: name,
	}
}

func (i *BlockSnapshotCleanup) Cleanup(ctx context.Context) error {
	testCtx := tester.ExtractCtx(ctx)
	api := block.NewAPI(testCtx.ScwClient)

	resp, err := api.ListSnapshots(&block.ListSnapshotsRequest{
		Name:      &i.snapshotName,
		Zone:      i.zone,
		ProjectID: &testCtx.ProjectID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to list block snapshots: %w", err)
	}

	if len(resp.Snapshots) == 0 {
		return fmt.Errorf("could not find any block snapshot by the name %q", i.snapshotName)
	}

	err = api.DeleteSnapshot(&block.DeleteSnapshotRequest{
		Zone:       i.zone,
		SnapshotID: resp.Snapshots[0].ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete block snapshot: %w", err)
	}

	return nil
}
