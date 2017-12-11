package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
func dirToGophermap(path string) (gophermap gophermap, err error) {
	// Make sure the path is readable
	fInfo, err := os.Stat(path)
	if err != nil {
		return gophermap, fmt.Errorf("failed to stat path %v: %v", path, err)
	}
	// Make sure it's a directory
	if fInfo.IsDir() != true {
		return gophermap, fmt.Errorf("supplied path %v is not a directory", err)
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return gophermap, fmt.Errorf("failed to read directory %v: %v", path, err)
	}
	// Loop over the slice and fill in our gophermap
	for i := 0; i < len(files); i++ {
		var gopherItem gopherItem
		gopherItem.gophertype, err = gophertype(filepath.Join(path, files[i].Name()))
		if err != nil {
			return gophermap, err
		}
		gopherItem.displayString = files[i].Name()
		// Have to make sure that the selector isn't the absolute path, but only the part not included in config.gopherroot.
		// This is kinda messy
		rawSelector, err := trimRootPath(Config.Gopherroot, filepath.Join(path, files[i].Name()))
		if err != nil {
			return gophermap, err
		}
		// Also has to handle windows "\"-separated paths
		gopherItem.selector = strings.Replace(rawSelector, "\\", "/", -1)
		gopherItem.host = Config.Address
		gopherItem.port = Config.Port
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

// gophertype gets the gopher response code via file extension
// Might get replaced with MIME type sniffing in the future, although it will be slower due to having to open the file to check magic numbers (and possibly require cgo).
// Some clients might also depend on file extensions.
// TODO: Implement some of the more obscure content types and gopher+/nonstandart server extensions
func gophertype(path string) (gophertype string, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("Failed to stat file %v: %v", path, err)
	}
	if fi.IsDir() == true {
		return "1", nil
	}
	fileExt := strings.ToLower(filepath.Ext(path))
	switch fileExt {
	//Image formats
	//Might move these to a global array later
	case ".jpg", ".jpeg", ".jp2", ".jpx", ".tiff", ".tif", ".bmp", ".png", ".webp", ".pbm", ".pgm", ".ppm", ".pnm", "heif", ".heic", ".bpg", ".ecw", ".fits", ".fit", ".fts", ".flif", ".ico", ".jxr", ".hdp", ".svg":
		return "I", nil
	// Spec treats GIF specially
	case ".gif":
		return "g", nil
	// DOS/Windows executables
	case ".com", ".exe":
		return "5", nil
	// Text files of various kinds
	// TODO: Expand this list
	case ".txt", ".md", ".rtf":
		return "0", nil

	// Other types of files
	// We assume unknown files are binaries, that way the user is prompted to download them instead of seeing a bunch of gibberish
	default:
		return "9", nil
	}
}

// sendFile sends a file specified by path over the connection
func (conn gopherConn) sendFile(path string) (err error) {
	file, err := os.Open(path)
	log.Print("File:" + path)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, file)
	if err != nil {
		return fmt.Errorf("Error while sending file %v: %v", path, err.Error())
	}
	return nil
}
