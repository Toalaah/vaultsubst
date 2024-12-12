package path_test

import (
	"fmt"
	"io/fs"
	"os"
	gopath "path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toalaah/vaultsubst/internal/path"
)

func TestIsDir(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

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

func TestNormalizePath(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	assert.Nil(os.Chdir(path.Root))

	for _, c := range []struct {
		name        string
		path        string
		expectedRes string
		expectedErr error
	}{
		{
			name:        "strip-spaces-1",
			path:        "   ./../foo/bar/baz   ",
			expectedRes: "/foo/bar/baz",
			expectedErr: nil,
		},
		{
			name:        "strip-spaces-2",
			path:        "   ./../../../foo   /     bar/baz   ",
			expectedRes: "/foo   /     bar/baz",
			expectedErr: nil,
		},
		{
			name:        "strip-trailing-slash",
			path:        "foo/bar/baz/",
			expectedRes: "/foo/bar/baz",
			expectedErr: nil,
		},
		{
			name:        "root-dir",
			path:        path.Root,
			expectedRes: path.Root,
			expectedErr: nil,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			res, err := path.Normalize(c.path)
			assert.Equal(c.expectedRes, res)
			assert.IsType(c.expectedErr, err)
		})
	}
}

func TestDepth(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	tmpDir := t.TempDir()
	assert.Nil(os.Chdir(tmpDir))

	for _, c := range []struct {
		name        string
		subpath     string
		root        string
		expectedRes int
		expectedErr error
	}{
		{
			name:        "single-depth",
			subpath:     filepath.Join(tmpDir, "foo"),
			expectedRes: 1,
			expectedErr: nil,
		},
		{
			name:        "single-depth-up-down",
			subpath:     filepath.Join(tmpDir, "foo/bar/baz/../baz/../.."),
			expectedRes: 1,
			expectedErr: nil,
		},
		{
			name:        "dirty-path",
			subpath:     filepath.Join(tmpDir, "foo////bar/baz/../..   "),
			expectedRes: 1,
			expectedErr: nil,
		},
		{
			name:        "not-subpath",
			subpath:     "/etc/passwd",
			expectedRes: -1,
			expectedErr: fmt.Errorf("%s is not a subpath of %s", "/etc/passwd", tmpDir),
		},
		{
			name:        "root",
			subpath:     "/etc/passwd",
			root:        path.Root,
			expectedRes: 2,
			expectedErr: nil,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := tmpDir
			if c.root != "" {
				root = c.root
			}
			depth, err := path.Depth(root, c.subpath)
			assert.Equal(c.expectedRes, depth)
			assert.Equal(c.expectedErr, err)
		})
	}
}
