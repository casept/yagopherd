package main

import (
	"errors"
	"fmt"
	"net"
)

// gopherConn is just used so gopher-connection specific methods can be cleanly declared.
type gopherConn struct {
	net.Conn
}

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

// A gophermap is a flat slice of gopherItems
type gophermap struct {
	items   []gopherItem
	gopherP bool // Whether to use gopher+
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

// send sends the gohpermap over conn.
func (conn gopherConn) sendGophermap(gohpermap gophermap) (err error) {
	sGophermap, err := gophermap.serialize()
	if err != nil {
		return err
	}
	_, err = conn.Write(sGophermap)
	if err != nil {
		return err
	}
	return nil
}
