package main

import (
	"bufio"
	"io"
	"log"
	"strings"
)

// GopherReq holds all information related to a client request (but not the response, that's gopherItem's job)
type gopherReq struct {
	selector string // Selector the client sent (blank=client requested gopherroot)
	path     string // Path to the requested item
	gopherP  bool   // Whether the client supports gopher+
}

// extractReq extracts data about a request from the receiving gopherConn.
func extractReq(iord io.Reader) (req gopherReq, err error) {
	// Extract the selector from request
	rd := bufio.NewReader(iord)

	// TODO: Simplify
	// This needs to work on \r\n (gohper client) and \r\n\t+ (gopher+ client).
	// So it's unknown up to which delimiter to read, as the client could always send the "/t+" part at an unknown point in the future.

	if  rawStr, err := rd.ReadString('\n')
	log.Printf("rawStr: %q\n", rawStr)
	if err != nil {
		return req, err
	}
	// Remove the \r\n sent at the end of a request (this one was fun to debug... not.)
	// Kinda messy, fuck it
	if strings.HasSuffix(rawStr, "\r\n\t+") {
		req.selector = strings.TrimRight(rawStr, "\r\n\t+")
		req.gopherP = true
	} else {
		if strings.HasSuffix(rawStr, "\r\n") {
			req.selector = strings.TrimRight(rawStr, "\r\n")
			req.gopherP = false
		}
	}
	return req, nil
}
