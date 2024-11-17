package path_test

import (
	"io/fs"
	"os"
	gopath "path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/path"
)

func TestIsDir(t *testing.T) {
	assert := assert.New(t)
	t.Parallel()

	// Create testing directories/files.
	dir := t.TempDir()
	file, err := os.CreateTemp(dir, "file")
	assert.Nil(err)
	removedDir := t.TempDir()
	removedFile := gopath.Join(removedDir, "file")
	assert.Nil(err)
	os.Remove(removedDir)

	for _, c := range []struct {
		name        string
		path        string
		expectedErr error
		expectedRes bool
	}{
		{
			name:        "directory-exists",
			path:        dir,
			expectedErr: nil,
			expectedRes: true,
		},
		{
			name:        "file-exists",
			path:        file.Name(),
			expectedErr: nil,
			expectedRes: false,
		},
		{
			name:        "directory-does-not-exist",
			path:        removedDir,
			expectedErr: &fs.PathError{},
			expectedRes: false,
		},
		{
			name:        "file-does-not-exist",
			path:        removedFile,
			expectedErr: &fs.PathError{},
			expectedRes: false,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			isDir, err := path.IsDir(c.path)
			assert.Equal(c.expectedRes, isDir)
			assert.IsType(c.expectedErr, err)
		})
	}
}
