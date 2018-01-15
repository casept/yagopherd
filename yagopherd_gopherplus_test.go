package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"strconv"
	"testing"
)

// Test downloading files via gopher+
func TestGopherPDownload(t *testing.T) {

	testFiles, err := ioutil.ReadDir("./testdata/gopherroot")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range testFiles {
		// This only handles files at the root of the gopherroot, but that's sufficient for test purposes.
		// Only handle files, directory listing retrieval testing handled separately
		if file.IsDir() == false {
			// Connections cannot be reused in the gopher protocol.
			// Therefore create a new one each time.
			conn, err := net.DialTCP("tcp", nil, Addr)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			_, err = conn.Write([]byte(file.Name() + "\r\n\t+"))
			if err != nil {
				t.Fatal(err)
			}
			var receivedFile []byte
			receivedFile, err = ioutil.ReadAll(conn)
			if err != nil {
				t.Fatal(err)
			}

			// Retrieve the same file from disk.
			diskFile, err := ioutil.ReadFile("./testdata/gopherroot/" + file.Name())
			// Prepend the length of the file (as it should be returned by a gopher+ server if known).
			diskFile = append([]byte("\r\n"), diskFile...)
			// Subtract the already appended "\r\n" from the total length of the file
			diskFile = append([]byte(strconv.Itoa(len(diskFile)-2)), diskFile...)
			diskFile = append([]byte("+"), diskFile...)
			if bytes.Equal(receivedFile, diskFile) == false {
				// Dump out both files in hex if they don't match.
				t.Logf("Received file %v does not match expected file:\n received:\n%v\n expected: \n%v\n", file.Name(), string(receivedFile), string(diskFile))
				t.Fail()
			}
		}
	}
}
