package nas

import (
	"fmt"
	"github.com/hirochachacha/go-smb2"
	"github.com/hultan/softimdb/internal/data"
	"net"
	"path"
)

type Manager struct {
}

func ManagerNew() *Manager {
	return new(Manager)
}

func (m Manager) GetMovies() *[]string {
	session := make(map[string]string)
	session["Username"] = "per"
	session["Password"] = "KnaskimGjwQ6M!"
	session["Domain"] = ""

	client := connectClient("192.168.1.100", "Videos", session)

	fs, err := client.Mount(`Videos`)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = fs.Umount()
	}()

	// Get ignored paths
	db := data.DatabaseNew(false)
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

	for i:= range *dirs {
		dir := (*dirs)[i]

		if !contains(*moviePaths, dir) {
			*result = append(*result, dir)
		}
	}

	return result
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

		currentPath := path.Join(pathName,file.Name())
		ignore := getIgnorePath(ignoredPaths, currentPath)
		if ignore != nil && ignore.IgnoreCompletely {
			continue
		}
		if ignore == nil {
			*dirs = append(*dirs, currentPath)
		}
		readDirectoryEx(fs,path.Join(pathName, file.Name()), ignoredPaths, dirs)
	}
}

func contains(slice []string, find string) bool {
	for _, a := range slice {
		if a == find {
			return true
		}
	}
	return false
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
	//Checks for a connection on port
	conn, err := net.Dial("tcp", host+":445")
	if err != nil {
		panic(err)
	}

	//smb auth
	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     session["Username"],
			Password: session["Password"],
			Domain:   session["Domain"],
		},
	}

	//Returns a client session
	client, err := d.Dial(conn)
	if err != nil {
		fmt.Println("Connection failed")
		client.Logoff()
	} else {
		fmt.Println("Connection Succeeded")
	}
	return client
}
