package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// A gopherItem is one item that shows up on the client's menu for selection
type gopherItem struct {
	gophertype    string // Gophertype of the item
	displayString string // Display string that the client will render
	selector      string // Selector that the client sends to retrieve the item
	host          string // Hostname of the server the item resides on
	port          int    // Port of the server the item resides on
	isFS          bool   // Whether item resides on FS
	isRemote      bool   // Whether item resides on remote server
	fsLocation    string // Absolute path to the location on the filesystem (blank if item is remote)
	mimetype      string // Mimetype of the item
	gopherP       bool   // Whether item is a gopher+ item
}

// constructGopherItem constructs a gopherItem based on a supplied gopher request.
// Note that the item currently has to reside on the local filesystem, remote items are currently unsupported.
func constructGopherItem(selector string, gopherP bool) (item gopherItem, err error) {
	item.fsLocation, err = selectorToPath(selector)
	if err != nil {
		return gopherItem{}, err
	}
	item.gophertype, err = gophertype(item.fsLocation, gopherP)
	if err != nil {
		return gopherItem{}, err
	}
	// Display string = the last part of the selector.
	// yagohperd Selectors are built like filepaths, so the filepath package can be used to extract a "filename".
	item.displayString = filepath.Base(selector)

	// TODO: Implement support for remote hosts
	item.host = viper.GetString("address")
	item.port = viper.GetInt("port")

	// Check whether the item exists on the FS and is readable.
	if isReadable(item.fsLocation) {
		item.isFS = true
	} else {
		// Return an error if item isn't readable
		// TODO: Remove once remote hosts are supported
		return gopherItem{}, os.ErrNotExist
	}
	// Have to make sure that the selector isn't the absolute path, but only the part not included in gopherroot.
	// This is kinda messy
	trimmedSelector := strings.TrimPrefix(item.fsLocation, viper.GetString("gopherroot"))
	// Also has to handle windows "\"-separated paths
	item.selector = strings.Replace(trimmedSelector, "\\", "/", -1)

	return item, nil
}

// serialize serializes a gopherItem into a string.
// Note that the serialized string isn't terminated by \r\n or any other gopher(+) gophermap item terminator.
func (g *gopherItem) serialize() (serializedGopherItem string, err error) {
	// Check whether all required fields are filled out
	if len(g.gophertype) == 0 {
		return "", errors.New("gopherItem is invalid (gophertype missing)")
	}
	if len(g.displayString) == 0 {
		return "", errors.New("gopherItem is invalid (displayString missing)")
	}
	if len(g.selector) == 0 {
		return "", errors.New("gopherItem is invalid (selector missing)")
	}
	if len(g.host) == 0 {
		return "", errors.New("gopherItem is invalid (host missing)")
	}
	if g.port == 0 {
		return "", errors.New("gopherItem is invalid (port missing)")
	}

	return fmt.Sprintf("%v%v\t%v\t%v\t%d", g.gophertype, g.displayString, g.selector, g.host, g.port), nil
}
