package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/casept/sanitize"
	"github.com/spf13/viper"
)

// isReadable reports whether the named file or directory exists and is accessible.
func isReadable(dirPath string) bool {
	if _, err := os.Stat(dirPath); err != nil {
		return false
	}
	return true
}

// sendFile sends a file specified by dirPath over the connection
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

// selectorToPath returns the path on the FS to an item requested by the given selector.
func selectorToPath(selector string) (retPath string, err error) {
	if selector == "" {
		retPath = viper.GetString("gopherroot")
	} else {
		err = sanitize.ErrorIfNotSane(viper.GetString("gopherroot"), path.Join(viper.GetString("gopherroot"), selector))
		if err != nil {
			return "", err
		}
		retPath = filepath.Join(viper.GetString("gopherroot"), selector)
	}
	return retPath, err
}
