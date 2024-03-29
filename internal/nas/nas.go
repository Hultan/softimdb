package nas

import (
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

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
		log.Fatal(err)
	}

	dirs = &[]string{}

	err = filepath.WalkDir(config.RootDir, walk)
	if err != nil {
		log.Fatal(err)
	}

	// Get movie paths
	moviePaths, err := db.GetAllMoviePaths()
	if err != nil {
		log.Fatal(err)
	}

	result := m.removeMoviePaths(dirs, moviePaths)

	db.CloseDatabase()

	return result
}

func walk(_ string, d fs.DirEntry, err error) error {
	if d.Name() == "videos" { // Skip main directory
		return nil
	}
	//if d.Name() == "SEARCH FOR ME!" { // Skip main directory
	//	fmt.Println("SEARCH")
	//}
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

func getIgnorePath(paths []*data.IgnoredPath, name string) *data.IgnoredPath {
	for i := range paths {
		if strings.HasSuffix(paths[i].Path, name) {
			return paths[i]
		}
	}
	return nil
}
