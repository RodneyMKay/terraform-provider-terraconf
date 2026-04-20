package terraconf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
)

// MakeRelative converts an absolute path to a relative path with respect to the
// current working directory. This function might produce an error, since the
// current working directory might not be accessible.
func MakeRelative(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current working directory: %w", err)
	}

	var relativePath string
	if filepath.IsAbs(path) {
		rel, err := filepath.Rel(cwd, path)
		if err == nil {
			relativePath = rel
		} else {
			relativePath = filepath.Base(path)
		}
	} else {
		// keep the supplied relative path as-is
		relativePath = path
	}

	return relativePath, nil
}

// FindGlobFiles finds all files matching the given glob pattern. The pattern
// can include doublestar patterns like "**".
func FindGlobFiles(pattern string) ([]string, error) {
	matches, err := doublestar.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("globbing pattern %s: %w", pattern, err)
	}

	return matches, nil
}
