package storage_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/butlerwang/project-seed/backend/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorageUploadAndDownload(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()

	err := s.Upload(ctx, "test/file.txt", "text/plain", bytes.NewReader([]byte("hello")))
	require.NoError(t, err)

	rc, err := s.Download(ctx, "test/file.txt")
	require.NoError(t, err)
	defer rc.Close()

	body, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(body))
}

func TestMemoryStorageDownloadNotFound(t *testing.T) {
	s := storage.NewMemoryStorage()
	_, err := s.Download(context.Background(), "nonexistent")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}

func TestMemoryStorageDelete(t *testing.T) {
	s := storage.NewMemoryStorage()
	ctx := context.Background()
	require.NoError(t, s.Upload(ctx, "to-delete", "text/plain", bytes.NewReader([]byte("bye"))))
	require.NoError(t, s.Delete(ctx, "to-delete"))
	_, err := s.Download(ctx, "to-delete")
	assert.ErrorIs(t, err, storage.ErrNotFound)
}
