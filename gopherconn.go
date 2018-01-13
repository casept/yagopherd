package main

import "net"

// gopherConn is just used so gopher-connection specific methods can be cleanly declared.
type gopherConn struct {
	net.Conn
}
