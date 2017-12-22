package main

import (
	"errors"
	"net"
	"strconv"
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
	fsLocation    string // Absolute path to the location on the filesystem (blank if item is remote)
	mimetype      string // Mimetype of the item
}

// A gophermap is a flat slice of gopherItems
type gophermap struct {
	items   []gopherItem
	gopherP bool // Whether to use gopher+
}

// serialize serializes a gophermap.
func (gophermap *gophermap) serialize() (serializedGophermap []byte, err error) {
	// Check whether all required fields are filled out
	for i := 0; i < len(gophermap.items); i++ {
		if len(gophermap.items[i].gophertype) == 0 || len(gophermap.items[i].displayString) == 0 || len(gophermap.items[i].selector) == 0 || len(gophermap.items[i].host) == 0 || gophermap.items[i].port == 0 {
			return []byte{}, errors.New("gophermap is invalid")
		}
	}

	var itemTerminator string
	if gophermap.gopherP == true {
		itemTerminator = "\r\n\t+"
	} else {
		itemTerminator = "\r\n"
	}

	var serializedString string
	for i := 0; i < len(gophermap.items); i++ {
		serializedString = serializedString + gophermap.items[i].gophertype + gophermap.items[i].displayString + "\t" + gophermap.items[i].selector + "\t" + gophermap.items[i].host + "\t" + strconv.Itoa(gophermap.items[i].port) + itemTerminator
	}

	// Response must end with a "."
	serializedString = serializedString + "."
	return []byte(serializedString), nil
}

// send sends the gohpermap over conn.
func (gophermap gophermap) send(conn gopherConn) (err error) {
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
