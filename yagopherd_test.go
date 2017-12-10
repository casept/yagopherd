package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/phayes/freeport"
)

// Addr ist the Address of the server under test.
var Addr *net.TCPAddr

// Start the server on a random high port assigned by the kernel
func TestMain(m *testing.M) {
	Config.IsTesting = true
	var err error
	Config.Gopherroot, err = filepath.Abs("./testdata/gopherroot")
	if err != nil {
		log.Fatal(err)
	}
	Config.Port, err = freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
	}
	Config.Address = "localhost"

	// Call main() to start the server
	// A goroutine is used so the tests can be run at the same time
	go main()

	Addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", Config.Address, strconv.Itoa(Config.Port)))
	if err != nil {
		log.Fatal(err)
	}
	// Wait for the server to start, deadline is 5 seconds.
	for i := 0; i < 5; i++ {
		_, err = net.DialTCP("tcp", nil, Addr)
		if i > 4 {
			log.Fatal("Server bringup timed out.")
		}
		// Error -> server is not up and listening, retry
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}
		// No error -> succesfull connection -> server is running
		if err == nil {
			break
		}
	}
	retCode := m.Run()

	// Do some teardown
	// os.Exit() does not send a signal to the server process, which means it won't terminate cleanly.
	// Therefore we signal ourselves, invoking the shutdown handler.
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatal(err)
	}
	process.Signal(os.Interrupt)
	// The shutdown handler has no way to know whether the tests were successfull, therefore we do our own os.Exit() with our own return code.
	os.Exit(retCode)
}

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
			diskGophermap = diskGophermap + result[i] + "\t" + strconv.Itoa(Config.Port) + "\r\n"
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

			if bytes.Equal(receivedFile, diskFile) == false {
				// Dump out both files in hex if they don't match.
				t.Logf("Received file %v does not match expected file:\n received:\n%x\n expected: \n%x\n", file.Name(), receivedFile, diskFile)
				t.Fail()
			}
		}
	}
}
