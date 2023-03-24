package nas

import (
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/hirochachacha/go-smb2"

	"github.com/hultan/crypto"
	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

// Manager represents a NAS manager.
type Manager struct {
	database *data.Database
}

// ManagerNew creates a new Manager.
func ManagerNew(database *data.Database) *Manager {
	manager := new(Manager)
	manager.database = database
	return manager
}

// Disconnect closes the connection to the NAS.
func (m Manager) Disconnect() {
	m.database = nil
}

// GetMovies returns a list of movie paths on the NAS.
func (m Manager) GetMovies(config *config.Config) *[]string {
	session := make(map[string]string)
	session["Username"] = config.Nas.User
	session["Password"] = m.getPassword(config.Nas.Password)
	session["Domain"] = ""

	client := connectClient(config.Nas.Address, config.Nas.Folder, session)

	fs, err := client.Mount(config.Nas.Folder)
	if err != nil {
		return nil
	}
	defer func() {
		err = fs.Umount()
	}()

	// Get ignored paths
	db := data.DatabaseNew(false, config)
	ignoredPaths, err := db.GetAllIgnoredPaths()
	if err != nil {
		panic(err)
	}

	var dirs = &[]string{}
	readDirectoryEx(fs, ".", ignoredPaths, dirs)

	// Get movie paths
	moviePaths, err := db.GetAllMoviePaths()
	if err != nil {
		panic(err)
	}

	result := m.removeMoviePaths(dirs, moviePaths)

	db.CloseDatabase()

	return result
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

func readDirectoryEx(fs *smb2.Share, pathName string, ignoredPaths []*data.IgnoredPath, dirs *[]string) {
	files, err := fs.ReadDir(pathName)
	if err != nil {
		return
	}
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		currentPath := path.Join(pathName, file.Name())
		ignore := getIgnorePath(ignoredPaths, currentPath)
		if ignore != nil && ignore.IgnoreCompletely {
			continue
		}
		if ignore == nil {
			*dirs = append(*dirs, currentPath)
		}
		readDirectoryEx(fs, path.Join(pathName, file.Name()), ignoredPaths, dirs)
	}
}

func getIgnorePath(paths []*data.IgnoredPath, name string) *data.IgnoredPath {
	for i := range paths {
		if paths[i].Path == name {
			return paths[i]
		}
	}
	return nil
}

func connectClient(host string, _ string, session map[string]string) *smb2.Client {
	// Checks for a connection on port
	conn, err := net.Dial("tcp", host+":445")
	if err != nil {
		panic(err)
	}

	// Smb auth
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     session["Username"],
			Password: session["Password"],
			Domain:   session["Domain"],
		},
	}

	// Returns a client session
	client, err := d.Dial(conn)
	if err != nil {
		fmt.Println("Connection failed")
		client.Logoff()
	} else {
		fmt.Println("Connection Succeeded")
	}
	return client
}
