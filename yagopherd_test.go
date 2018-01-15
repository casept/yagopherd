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

// Test obtaining gophermaps of various directories
func TestGophermap(t *testing.T) {
	tc := []struct {
		selector  string // Selector that should result in gophermap
		gophermap string // The expected gophermap (without ports, those are appended dynamically)
		gopherP   bool   // Whether it's a gopher+ gophermap
	}{
		{"\r\n", "1test\t/test\tlocalhost\r\n5test.exe\t/test.exe\tlocalhost\r\ngtest.gif\t/test.gif\tlocalhost\r\nItest.jpg\t/test.jpg\tlocalhost\r\n0test.txt\t/test.txt\tlocalhost\r\n9test.unknownfiletype\t/test.unknownfiletype\tlocalhost\r\n.", false},
		{"\r\n\t+", "1test\t/test\tlocalhost\r\n5test.exe\t/test.exe\tlocalhost\r\ngtest.gif\t/test.gif\tlocalhost\r\n:test.jpg\t/test.jpg\tlocalhost\r\n0test.txt\t/test.txt\tlocalhost\r\n9test.unknownfiletype\t/test.unknownfiletype\tlocalhost\r\n.", true},
	}

	for _, tt := range tc {
		// Setup connection
		conn, err := net.DialTCP("tcp", nil, Addr)
		if err != nil {
			t.Fatal(err)
		}

		// Request the gophermap
		_, err = conn.Write([]byte(tt.selector))
		if err != nil {
			t.Fatal(err)
		}

		receivedGophermap, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatalf("Error while trying to read server response: %v", err)
		}

		// We have to append the ephemeral port of the gopher server to each line of the test gophermap
		// Break down into lines
		var AppendStr string
		if tt.gopherP {
			AppendStr = "\t+\r\n"
		} else {
			AppendStr = "\r\n"
		}
		result := strings.Split(string(tt.gophermap), "\r\n")
		// Reconstruct the contents of the file, appending the port where needed
		var testGophermap string
		for i := range result {
			if result[i] == "." {
				testGophermap = testGophermap + "."
				break
			} else {
				testGophermap = testGophermap + result[i] + "\t" + strconv.Itoa(viper.GetInt("port")) + AppendStr
			}
		}

		if bytes.Equal(receivedGophermap, []byte(testGophermap)) == false {
			t.Errorf("Received gophermap does not match expected gophermap:\n received:\n%v\n expected:\n%v\n", string(receivedGophermap), string(testGophermap))
		}
		conn.Close()
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
