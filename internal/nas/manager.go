package nas

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/hultan/crypto"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

const basePath = "/home/per/media/videos/"

// Manager represents a NAS manager.
type Manager struct {
	database *data.Database
}

var dirs = &[]string{}
var ignoredPaths []*data.IgnoredPath

// ManagerNew creates a new Manager.
func ManagerNew(database *data.Database) *Manager {
	manager := new(Manager)
	manager.database = database
	return manager
}

// GetMovies returns a list of movie paths on the NAS.
func (m Manager) GetMovies(config *config.Config) *[]string {
	var err error

	// Get ignored paths
	db := data.DatabaseNew(false, config)
	ignoredPaths, err = db.GetAllIgnoredPaths()
	if err != nil {
		panic(err)
	}

	err = filepath.WalkDir(basePath, walk)
	if err != nil {
		panic(err)
	}

	// Get movie paths
	moviePaths, err := db.GetAllMoviePaths()
	if err != nil {
		panic(err)
	}

	result := m.removeMoviePaths(dirs, moviePaths)

	db.CloseDatabase()

	return result
}

func walk(_ string, d fs.DirEntry, err error) error {
	if d.Name() == "videos" { // Skip main directory
		return nil
	}
	if err != nil { // Skip on errors
		return err
	}
	if !d.IsDir() { // Skip files
		return nil
	}

	// Skip ignored paths
	ignore := getIgnorePath(ignoredPaths, d.Name())
	if ignore != nil {
		return filepath.SkipDir
	}
	*dirs = append(*dirs, d.Name())

	return nil
}

func (m Manager) removeMoviePaths(dirs *[]string, moviePaths *[]string) *[]string {
	var result = &[]string{}

	for i := range *dirs {
		dir := (*dirs)[i]

		if !containsString(*moviePaths, dir) {
			*result = append(*result, dir)
		}
	}

	return result
}

// containsString : Returns true if the slice contains the string
//
//	in find, otherwise returns false.
func containsString(slice []string, find string) bool {
	for _, a := range slice {
		if a == find {
			return true
		}
	}
	return false
}

func (m Manager) getPassword(encrypted string) string {
	c := &crypto.Crypto{}
	password, err := c.Decrypt(encrypted)
	if err != nil {
		panic(err)
	}
	return strings.Replace(password, "\n", "", -1)
}

func getIgnorePath(paths []*data.IgnoredPath, name string) *data.IgnoredPath {
	for i := range paths {
		if strings.Contains(paths[i].Path, name) {
			return paths[i]
		}
	}
	return nil
}
