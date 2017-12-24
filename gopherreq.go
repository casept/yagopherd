package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// GopherReq holds all information related to a client request (but not the response, that's gopherItem's job)
type gopherReq struct {
	selector string // Selector the client sent (blank=client requested gopherroot)
	path     string // Path to the requested item
	gopherP  bool   // Whether the client supports gopher+
}

// extractReq extracts data about a request from the receiving gopherConn.
func extractReq(conn gopherConn) (req gopherReq, err error) {
	// This needs to work on \r\n (gohper client) and \r\n\t+ (gopher+ client).
	// Therefore bufio's ReadString() cannot be used.
	// So we read for no more than 0.2 seconds, then check if output ends in "\n" or "+"
	// If it does we're done reading.
	// If it doesn't, repeat until the timeout specified by the selectortimeout viper key is reached.

	var t time.Duration
	// See https://stackoverflow.com/questions/24339660/read-whole-data-with-golang-net-conn-read
	buf := make([]byte, 0, viper.GetInt("selectorlimit"))
	tmp := make([]byte, viper.GetInt("selectorlimit"))
	for {
		// This is needed because there is no other way to timeout io.Copy.
		now := time.Now()
		conn.SetReadDeadline(now.Add(200 * time.Millisecond))
		n, err := conn.Read(tmp)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				// Copy timed out as planned, nothing to see here.
			} else {
				return req, err
			}
		}
		// Increase the time counter, when it exceeds the allowed value the connection times out.
		t = t + 200*time.Millisecond
		// Check if the buf slice would be over capacity by appending.
		if len(tmp)+len(buf) > cap(buf) {
			return req, fmt.Errorf("client selector exceeds maximum allowed length (sent %d bytes, %d bytes allowed)", len(tmp)+len(buf), cap(buf))
		}
		buf = append(buf, tmp[:n]...)
		// Check the last symbol in the buffer to see whether the client has finished.
		if len(buf) > 1 {
			if buf[len(buf)-1] == '+' || buf[len(buf)-1] == '\n' {
				break
			} else {
				// Check for timeout
				if t >= viper.GetDuration("selectortimeout") {
					return req, fmt.Errorf("selector reception timed out (took %v)", t.String())
				}
			}

		}
	}
	// Unset the deadline.
	conn.SetReadDeadline(time.Time{})

	rawStr := string(buf)
	// Remove the \r\n sent at the end of a request
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
