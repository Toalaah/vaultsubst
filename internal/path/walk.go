package path

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// Normalize normalizes a path. This includes converting the path to its
// absolute representation, stripping leading and trailing spaces, and removing
// trailing filepath separators.
func Normalize(path string) (string, error) {
	path, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return "", err
	}
	if path == string(filepath.Separator) {
		return path, nil
	}
	return strings.TrimSuffix(path, string(filepath.Separator)), nil
}

// Depth returns the depth of a sub-path relative to a base-path.
func Depth(base, path string) (int, error) {
	path, err := Normalize(path)
	if err != nil {
		return -1, err
	}
	base, err = Normalize(base)
	if err != nil {
		return -1, err
	}
	if base == string(filepath.Separator) {
		goto end
	}
	if !strings.HasPrefix(path, base) {
		return -1, fmt.Errorf("%s is not a subpath of %s", path, base)
	}
	path = strings.TrimPrefix(path, base)
end:
	return strings.Count(path, string(filepath.Separator)), nil
}

// WalkDir walks a directory tree and applies `fn` to each file/directory. A
// depth of 0 or less implies limitless recursion when walking.
func WalkDir(dir string, depth int, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, e error) error {
		currentDepth, err := Depth(dir, path)
		if err != nil {
			return err
		}
		if depth > 0 && currentDepth > depth {
			return nil
		}
		return fn(path, d, e)
	})
}
