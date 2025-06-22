// my-go-shell/fsutil/fsutil.go
package fsutil

import (
	"os"
	"path/filepath" // Use filepath for path manipulation
	"strings"
)

// Finder is a struct to manage path-related operations.
type Finder struct {
	paths []string
}

// NewFinder creates and initializes a new Finder.
func NewFinder(paths []string) *Finder {
	return &Finder{paths: paths}
}

func IsValidPath(fullPath string) bool {
	_, err := os.Stat(fullPath)
	return err == nil
}

// IsValidPath checks if a given full path exists and is accessible.
func (f *Finder) IsValidPath(fullPath string) bool {
	return IsValidPath(fullPath)
}

// GetAbsolutePath converts a relative or special path to an absolute path.
func (f *Finder) GetAbsolutePath(path string) string {
	if path == "" {
		return path
	} else if path == "~" {
		homeDir := os.Getenv("HOME")
		return homeDir
	} else if path[0] == '/' {
		return path
	} else {
		directory_changes := strings.Split(path, "/")
		cwd, _ := os.Getwd()
		curDirectories := strings.Split(cwd, "/")

		for _, change := range directory_changes {
			if change == ".." {
				// Pop the last directory if not at the root
				if len(curDirectories) > 0 {
					curDirectories = curDirectories[:len(curDirectories)-1]
				}
			} else if change != "." && change != "" {
				// Push the directory to the current path
				curDirectories = append(curDirectories, change)
			}
		}
		// Join the directories back into a single path
		absolutePath := strings.Join(curDirectories, "/")
		if !strings.HasPrefix(absolutePath, "/") {
			absolutePath = "/" + absolutePath
		}
		return absolutePath
	}
}

// FindExecutablePath searches for an executable in the configured paths.
func (f *Finder) FindExecutablePath(command string) (string, bool) {
	for _, dir := range f.paths {
		fullPath := filepath.Join(dir, command) // Use filepath.Join
		if f.IsValidPath(fullPath) {
			return fullPath, true
		}
	}
	return "", false
}

func (f *Finder) GetExecutables() []string {
	executables := []string{}
	for _, dir := range f.paths {
		files, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories that cannot be read
		}
		for _, file := range files {
			if file.IsDir() {
				continue // Skip directories
			}
			if info, err := file.Info(); err == nil && info.Mode()&0111 != 0 { // Check if executable
				executables = append(executables, file.Name()) // Use filepath.Join for full path
				// executables = append(executables, filepath.Join(dir, file.Name())) // Use filepath.Join for full path
			}
		}
	}
	return executables
}
