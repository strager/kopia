package content

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kopia/kopia/internal/blobtesting"
	"github.com/kopia/kopia/internal/cache"
	"github.com/kopia/kopia/internal/gather"
	"github.com/kopia/kopia/internal/testlogging"
	"github.com/kopia/kopia/repo/blob"
)

func TestContentCacheForData(t *testing.T) {
	ctx := testlogging.Context(t)

	underlyingData := blobtesting.DataMap{}
	underlying := blobtesting.NewMapStorage(underlyingData, nil, nil)

	cacheData := blobtesting.DataMap{}
	metadataCacheStorage := blobtesting.NewMapStorage(cacheData, nil, nil).(cache.Storage)

	dataCache, err := newContentCacheForData(ctx, underlying, metadataCacheStorage, cache.SweepSettings{
		MaxSizeBytes: 100,
	}, []byte{1, 2, 3, 4})
	require.NoError(t, err)

	var tmp gather.WriteBuffer
	defer tmp.Close()

	// get something we don't have in the underlying storage
	require.ErrorIs(t, dataCache.getContent(ctx, "key1", "blob1", 0, 3, &tmp), blob.ErrBlobNotFound)

	require.NoError(t, underlying.PutBlob(ctx, "blob1", gather.FromSlice([]byte{1, 2, 3, 4, 5, 6}), blob.PutOptions{}))

	require.NoError(t, dataCache.getContent(ctx, "xkey1", "blob1", 0, 3, &tmp))
	require.Equal(t, []byte{1, 2, 3}, tmp.ToByteSlice())

	require.NoError(t, dataCache.getContent(ctx, "xkey2", "blob1", 3, 3, &tmp))
	require.Equal(t, []byte{4, 5, 6}, tmp.ToByteSlice())

	// cache has 2 entries, one for each section of the blob.
	require.Len(t, cacheData, 2)

	// keys are mangled so that last character (which is always 0..9 a..f) is the first.
	require.Contains(t, cacheData, blob.ID("key1x"))
	require.Contains(t, cacheData, blob.ID("key2x"))

	dataCache.close(ctx)

	// get slice with cache miss
	require.NoError(t, underlying.PutBlob(ctx, "blob2", gather.FromSlice([]byte{1, 2, 3, 4, 5, 6}), blob.PutOptions{}))
	require.NoError(t, dataCache.getContent(ctx, "aaa1", "blob2", 3, 3, &tmp))
	require.Equal(t, []byte{4, 5, 6}, tmp.ToByteSlice())

	// even-length content IDs are not mangled
	require.Len(t, cacheData, 3)
	require.Contains(t, cacheData, blob.ID("aaa1"))
}

func TestContentCacheForData_Passthrough(t *testing.T) {
	underlyingData := blobtesting.DataMap{}
	underlying := blobtesting.NewMapStorage(underlyingData, nil, nil)

	ctx := testlogging.Context(t)

	dataCache, err := newContentCacheForData(ctx, underlying, nil, cache.SweepSettings{
		MaxSizeBytes: 100,
	}, []byte{1, 2, 3, 4})

	require.NoError(t, err)
	require.NoError(t, underlying.PutBlob(ctx, "blob1", gather.FromSlice([]byte{1, 2, 3, 4, 5, 6}), blob.PutOptions{}))

	var tmp gather.WriteBuffer
	defer tmp.Close()

	require.NoError(t, dataCache.getContent(ctx, "key1", "blob1", 0, 5, &tmp))
	require.Equal(t, []byte{1, 2, 3, 4, 5}, tmp.ToByteSlice())
}
