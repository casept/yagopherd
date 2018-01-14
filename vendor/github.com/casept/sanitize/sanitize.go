// Package sanitize provides functions which check whether a path expands to be outside the allowed root directory.
package sanitize

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathNotAbsoluteError is returned if a path isn't absolute.
type PathNotAbsoluteError struct {
	path string
}

// Err serializes the PathNotAbsoluteError.
func (p PathNotAbsoluteError) Error() string {
	return fmt.Sprintf("the path %v is not absolute", p.path)
}

// IsSane checks whether a user-supplied path expands to be outside the provided absolute rootPath.
// The provided root path has to be an absolute path in order to avoid surprises during expansion.
// The function will return a PathNotAbsoluteError otherwise.
func IsSane(rootPath string, path string) (bool, error) {
	// Check if root path is absolute
	if filepath.IsAbs(rootPath) == false {
		return false, PathNotAbsoluteError{rootPath}
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(absPath, rootPath) {
		return true, nil
	}
	return false, nil
}

// PathNotSaneError is returned by ErrorIfNotSane() if a path isn't absolute.
type PathNotSaneError struct {
	rootPath string // The root path.
	path     string // The untrusted path.
}

// Err serializes the PathNotSaneError.
func (p PathNotSaneError) Error() string {
	return fmt.Sprintf("the path %v expands to be outside the root directory %v", p.path, p.rootPath)
}

// ErrorIfNotSane is like IsSane, except that it returns a PathNotSaneError if the path is outside the root path.
// It will also return any errors encountered by IsSane.
// This function is useful if you wish to pass the error along in your own program, as you don't have to come up with a custom error type, message etc.
func ErrorIfNotSane(rootPath, path string) error {
	sane, err := IsSane(rootPath, path)
	if err != nil {
		return err
	}
	if sane == false {
		return PathNotSaneError{rootPath, path}
	}
	return nil
}
