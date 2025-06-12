package nas

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

// Manager represents a NAS manager.
type Manager struct {
	database     *data.Database
	dirs         []string
	ignoredPaths []*data.IgnoredPath
}

// ManagerNew creates a new Manager.
func ManagerNew(database *data.Database) *Manager {
	manager := new(Manager)
	manager.database = database
	return manager
}

// GetMovies returns a list of movie paths on the NAS.
func (m *Manager) GetMovies(config *config.Config) ([]string, error) {
	ignoredPaths, err := m.database.GetAllIgnoredPaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get ignored paths: %w", err)
	}
	m.ignoredPaths = ignoredPaths

	dir, err := os.Open(config.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open root dir: %w", err)
	}
	defer func() {
		if cerr := dir.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close failed: %w", cerr)
		}
	}()

	entries, err := dir.Readdirnames(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir entries: %w", err)
	}

	var pathsOnNAS []string
	for _, entry := range entries {
		if !getIgnorePath(m.ignoredPaths, entry) {
			pathsOnNAS = append(pathsOnNAS, entry)
		}
	}

	// Get movie paths to exclude
	pathsInDB, err := m.database.GetAllMoviePaths()
	if err != nil {
		return nil, fmt.Errorf("failed to get movie paths: %w", err)
	}

	result := m.removeExistingPaths(pathsOnNAS, pathsInDB)
	slices.Sort(result)
	return result, nil
}

func (m *Manager) removeExistingPaths(pathsOnNAS []string, pathsInDB []string) []string {
	pathsInDBMap := make(map[string]struct{}, len(pathsInDB))
	for _, path := range pathsInDB {
		pathsInDBMap[path] = struct{}{}
	}

	var result []string
	for _, dir := range pathsOnNAS {
		if dir == "" {
			continue
		}
		if _, exists := pathsInDBMap[dir]; !exists {
			result = append(result, dir)
		}
	}

	return result
}

func getIgnorePath(paths []*data.IgnoredPath, name string) bool {
	for i := range paths {
		if strings.HasSuffix(paths[i].Path, name) {
			return true
		}
	}
	return false
}
