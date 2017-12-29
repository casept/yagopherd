package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// trimRootPath trims a supplied path by trimming off rootPath.
// Both arguments have to be absolute paths.
func trimRootPath(rootPath string, path string) (trimmedPath string, err error) {
	// Expand both paths so that string processing can be used
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(absPath, absRoot), nil
}

// dirToGophermap adds all files and directories in path to gophermap
// This function currently ignores the contents of `.gophermap` files
func dirToGophermap(path string, gopherP bool) (gophermap gophermap, err error) {
	// Indicate it's a gopher+ gophermap if told to
	gophermap.gopherP = gopherP
	// Make sure the path is readable
	fInfo, err := os.Stat(path)
	if err != nil {
		return gophermap, err
	}
	// Make sure it's a directory
	if fInfo.IsDir() != true {
		return gophermap, fmt.Errorf("supplied path %v is not a directory", err)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return gophermap, err
	}
	// Loop over the slice and fill in our gophermap
	for i := 0; i < len(files); i++ {
		var gopherItem gopherItem
		gopherItem.gophertype, err = gophertype(filepath.Join(path, files[i].Name()), gopherP)
		if err != nil {
			return gophermap, err
		}
		gopherItem.displayString = files[i].Name()
		// Have to make sure that the selector isn't the absolute path, but only the part not included in gopherroot.
		// This is kinda messy
		rawSelector, err := trimRootPath(viper.GetString("gopherroot"), filepath.Join(path, files[i].Name()))
		if err != nil {
			return gophermap, err
		}
		// Also has to handle windows "\"-separated paths
		gopherItem.selector = strings.Replace(rawSelector, "\\", "/", -1)
		gopherItem.host = viper.GetString("address")
		gopherItem.port = viper.GetInt("port")
		gopherItem.fsLocation, err = filepath.Abs(filepath.Join(path, files[i].Name()))
		if err != nil {
			return gophermap, err
		}
		gophermap.items = append(gophermap.items, gopherItem)
	}
	return gophermap, nil
}

// appendDir takes a (potentially user-supplied) relative path and an absolute "root" path.
// The relative path gets added to the absolute path in a way which aims to protect against users supplying malicious paths.
func appendDir(root string, userPath string) (path string, err error) {
	// Make sure gopherroot is absolute
	if filepath.IsAbs(root) == false {
		return "", fmt.Errorf("Root path %v is not absolute", root)
	}
	//https://groups.google.com/forum/#!topic/Golang-Nuts/w9qH4rR_XPw
	path = filepath.Join(root, filepath.FromSlash(filepath.Clean("/"+userPath)))
	// Make sure we're still in a child of the root directory, otherwise error
	// path.Clean() *should* have expanded all globs/directory traversals, so the path should be fully expanded now
	// Therefore simply checking the beginning of the string should be enough
	if strings.HasPrefix(path, root) == false {
		return "", fmt.Errorf("Appended path expanded to %v which is outside the root path %v", path, root)
	}
	return path, nil
}

// sendFile sends a file specified by path over the connection
func (conn gopherConn) sendFile(path string, gopherP bool) (err error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}

	// gopher+ requires prepending the filesize.
	if gopherP {
		fInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		io.WriteString(conn, "+"+strconv.Itoa(int(fInfo.Size()))+"\r\n")
	}

	_, err = io.Copy(conn, file)
	if err != nil {
		return fmt.Errorf("Error while sending file %v: %v", path, err.Error())
	}
	return nil
}
