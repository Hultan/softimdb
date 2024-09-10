package nas

import (
	"log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

// Manager represents a NAS manager.
type Manager struct {
	database *data.Database
}

var dirs []string
var ignoredPaths []*data.IgnoredPath

// ManagerNew creates a new Manager.
func ManagerNew(database *data.Database) *Manager {
	manager := new(Manager)
	manager.database = database
	return manager
}

// GetMovies returns a list of movie paths on the NAS.
func (m Manager) GetMovies(config *config.Config) ([]string, error) {
	var err error
	dirs = make([]string, 3000)

	// Get ignored paths
	db := data.DatabaseNew(false, config)
	ignoredPaths, err = db.GetAllIgnoredPaths()
	if err != nil {
		return nil, err
	}

	scanDir(config.RootDir, "", "")

	for i, dir := range dirs {
		if getIgnorePath(ignoredPaths, dir) {
			dirs[i] = ""
		}
	}

	// Get movie paths
	moviePaths, err := db.GetAllMoviePaths()
	if err != nil {
		return nil, err
	}

	result := m.removeMoviePaths(dirs, moviePaths)

	db.CloseDatabase()

	slices.Sort(result)

	return result, nil
}

func scanDir(root, base, dir string) {
	file, err := os.Open(path.Join(root, base, dir))
	if err != nil {
		log.Fatal(err)
	}

	if dir != "" && getIgnorePath(ignoredPaths, dir) {
		return
	}

	foundDirs, err := file.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	dirs = append(dirs, dir)

	for _, foundDir := range foundDirs {
		if foundDir.IsDir() {
			scanDir(root, path.Join(base, dir), foundDir.Name())
		}
	}
}

func (m Manager) removeMoviePaths(dirs []string, moviePaths []string) []string {
	var result []string

	for i := range dirs {
		dir := dirs[i]

		if dir == "" {
			continue
		}

		if !slices.Contains(moviePaths, dir) {
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
