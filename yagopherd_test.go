package main

import (
	"bytes"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

// Test obtaining the gopherroot
func TestRoot(t *testing.T) {
	// Setup connection
	conn, err := net.DialTCP("tcp", nil, Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Request the gopherroot's gophermap
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	receivedGophermap, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("Error while trying to read server response: %v", err)
	}
	// The ./testdata/gophermap file contains a known-good gophermap of the test data directory.
	var rawDiskFile []byte
	rawDiskFile, err = ioutil.ReadFile("./testdata/gophermap")
	if err != nil {
		t.Fatal(err)
	}
	// We have to append the ephemeral port of the gopher server to each line of the file
	// Break down into lines (split on newline)
	result := strings.Split(string(rawDiskFile), "\r\n")
	// Reconstruct the contents of the file, appending the port where needed
	var diskGophermap string
	for i := range result {
		if result[i] == "." {
			diskGophermap = diskGophermap + "."
			break
		} else {
			diskGophermap = diskGophermap + result[i] + "\t" + strconv.Itoa(viper.GetInt("port")) + "\r\n"
		}
	}

	if bytes.Equal(receivedGophermap, []byte(diskGophermap)) == false {
		t.Logf("Received gophermap does not match expected gophermap:\n received:\n%v\n expected:\n%v\n", string(receivedGophermap), string(diskGophermap))
		t.Fail()
	}
}

// Test downloading files
func TestDownload(t *testing.T) {

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
			_, err = conn.Write([]byte(file.Name() + "\r\n"))
			if err != nil {
				t.Fatal(err)
			}
			var receivedFile []byte
			receivedFile, err = ioutil.ReadAll(conn)
			if err != nil {
				t.Fatal(err)
			}

			// Retrieve the same file from disk, check if the received one is identical.
			var diskFile []byte
			diskFile, err = ioutil.ReadFile("./testdata/gopherroot/" + file.Name())
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Equal(receivedFile, diskFile) == false {
				// Dump out both files in hex if they don't match.
				t.Logf("Received file %v does not match expected file:\n received:\n%v\n expected: \n%v\n", file.Name(), string(receivedFile), string(diskFile))
				t.Fail()
			}
		}
	}
}
