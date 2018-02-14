package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

// A gophermap is a flat slice of gopherItems
type gophermap struct {
	items   []gopherItem
	gopherP bool // Whether to use gopher+ to send the gophermap
}

// reqToGophermap returns the unserialized gophermap for a given gooherReq.
// This function currently ignores the contents of `.gophermap` files, aka only local items are currently supported.
func reqToGophermap(req req) (gophermap gophermap, err error) {
	gophermap.gopherP = req.gopherP
	dirPath, err := selectorToPath(req.selector)
	if err != nil {
		return gophermap, err
	}

	// Make sure the dirPath is readable
	fInfo, err := os.Stat(dirPath)
	if err != nil {
		return gophermap, err
	}
	// Make sure it's a directory
	if !fInfo.IsDir() {
		return gophermap, fmt.Errorf("supplied path %v is not a directory", dirPath)
	}
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return gophermap, err
	}

	// Loop over the slice and fill in our gophermap
	for i := 0; i < len(files); i++ {
		// constructGopherItem expects a selector
		item, err := constructGopherItem(fmt.Sprintf("%v/%v", req.selector, files[i].Name()), req.gopherP)
		if err != nil {
			return gophermap, err
		}
		gophermap.items = append(gophermap.items, item)
	}
	return gophermap, nil
}

// serialize serializes a gophermap.
func (gophermap gophermap) serialize() (serializedGophermap []byte, err error) {
	var itemTerminator string
	// Add a \t+ after the port to indicate item supports gopher+
	if gophermap.gopherP == true {
		itemTerminator = "\t+\r\n"
	} else {
		itemTerminator = "\r\n"
	}

	var serializedString string
	for i := 0; i < len(gophermap.items); i++ {
		nextItem, err := gophermap.items[i].serialize()
		if err != nil {
			return []byte{}, fmt.Errorf("gophermap is invalid: %v", err)
		}
		serializedString = serializedString + nextItem + itemTerminator
	}

	// Response must end with a "."
	serializedString = serializedString + "."
	return []byte(serializedString), nil
}

// send sends the gophermap over conn.
func (conn gopherConn) sendGophermap(gmap gophermap) (err error) {
	sGophermap, err := gmap.serialize()
	if err != nil {
		return err
	}
	_, err = conn.Write(sGophermap)
	if err != nil {
		return err
	}
	return nil
}
