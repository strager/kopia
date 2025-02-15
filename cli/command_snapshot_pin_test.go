package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/testutil"
	"github.com/kopia/kopia/snapshot"
	"github.com/kopia/kopia/tests/testenv"
)

func TestSnapshotPin(t *testing.T) {
	t.Parallel()

	runner := testenv.NewInProcRunner(t)
	e := testenv.NewCLITest(t, testenv.RepoFormatNotImportant, runner)

	defer e.RunAndExpectSuccess(t, "repo", "disconnect")
	e.RunAndExpectSuccess(t, "repo", "create", "filesystem", "--path", e.RepoDir)

	srcdir := testutil.TempDirectory(t)
	require.NoError(t, os.WriteFile(filepath.Join(srcdir, "some-file2"), []byte{1, 2, 3}, 0o755))

	var man snapshot.Manifest

	e.RunAndExpectSuccess(t, "policy", "set", srcdir, "--keep-latest=3", "--keep-hourly=0", "--keep-daily=0", "--keep-monthly=0", "--keep-weekly=0", "--keep-annual=0")

	testutil.MustParseJSONLines(t, e.RunAndExpectSuccess(t, "snapshot", "create", srcdir, "--pin=a", "--pin=b", "--json"), &man)
	require.Equal(t, []string{"a", "b"}, man.Pins)

	e.RunAndExpectSuccess(t, "snapshot", "list")

	// create more unpinned snapshots
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)

	var snapshots []*snapshot.Manifest

	testutil.MustParseJSONLines(t, e.RunAndExpectSuccess(t, "snapshot", "list", "--json"), &snapshots)
	snapshots = snapshot.SortByTime(snapshots, false)

	// make sure the pinned one is on top.
	require.Len(t, snapshots, 4)
	require.Equal(t, []string{"a", "b"}, snapshots[0].Pins)
	require.Empty(t, snapshots[1].Pins)
	require.Empty(t, snapshots[2].Pins)
	require.Empty(t, snapshots[3].Pins)

	// neither --add nor --remove were provided
	e.RunAndExpectFailure(t, "snapshot", "pin", string(snapshots[3].ID))
	e.RunAndExpectSuccess(t, "snapshot", "pin", string(snapshots[0].ID), "--add=c", "--remove=b")
	e.RunAndExpectSuccess(t, "snapshot", "pin", string(snapshots[3].ID), "--add=d")

	var snapshots2 []*snapshot.Manifest

	testutil.MustParseJSONLines(t, e.RunAndExpectSuccess(t, "snapshot", "list", "--json"), &snapshots2)
	snapshots2 = snapshot.SortByTime(snapshots2, false)
	require.Len(t, snapshots2, 4)

	require.Equal(t, []string{"a", "c"}, snapshots2[0].Pins)
	require.Empty(t, snapshots2[1].Pins)
	require.Empty(t, snapshots2[2].Pins)
	require.Equal(t, []string{"d"}, snapshots2[3].Pins)

	// create more unpinned snapshots
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)
	e.RunAndExpectSuccess(t, "snapshot", "create", srcdir)

	var snapshots3 []*snapshot.Manifest

	testutil.MustParseJSONLines(t, e.RunAndExpectSuccess(t, "snapshot", "list", "--json"), &snapshots3)
	snapshots3 = snapshot.SortByTime(snapshots3, false)
	require.Len(t, snapshots3, 5)

	require.Equal(t, []string{"a", "c"}, snapshots3[0].Pins)
	require.Equal(t, []string{"d"}, snapshots3[1].Pins)
	require.Empty(t, snapshots3[2].Pins)
	require.Empty(t, snapshots3[3].Pins)
	require.Empty(t, snapshots3[4].Pins)
}
