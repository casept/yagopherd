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
		// Gopher gopherroot
		{"\r\n", "1test\t/test\tlocalhost\r\n5test.exe\t/test.exe\tlocalhost\r\ngtest.gif\t/test.gif\tlocalhost\r\nItest.jpg\t/test.jpg\tlocalhost\r\n0test.txt\t/test.txt\tlocalhost\r\n9test.unknownfiletype\t/test.unknownfiletype\tlocalhost\r\n.", false},
		// Gopher+ gopherroot
		{"\r\n\t+", "1test\t/test\tlocalhost\r\n5test.exe\t/test.exe\tlocalhost\r\ngtest.gif\t/test.gif\tlocalhost\r\n:test.jpg\t/test.jpg\tlocalhost\r\n0test.txt\t/test.txt\tlocalhost\r\n9test.unknownfiletype\t/test.unknownfiletype\tlocalhost\r\n.", true},
		// Gopher subdir
		{"/test", "1test2\t/test/test2\tlocalhost\r\n5test2.exe\t/test/test2.exe\tlocalhost\r\ngtest2.gif\t/test/test2.gif\tlocalhost\r\nItest2.jpg\t/test/test2.jpg\tlocalhost\r\n0test2.txt\t/test/test2.txt\tlocalhost\r\n9test2.unknownfiletype\t/test/test2.unknownfiletype\tlocalhost\r\n.", false},
		// Gopher+ subdir
		{"/test\t+", "1test2\t/test/test2\tlocalhost\r\n5test2.exe\t/test/test2.exe\tlocalhost\r\ngtest2.gif\t/test/test2.gif\tlocalhost\r\n:test2.jpg\t/test/test2.jpg\tlocalhost\r\n0test2.txt\t/test/test2.txt\tlocalhost\r\n9test2.unknownfiletype\t/test/test2.unknownfiletype\tlocalhost\r\n.", true},
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

		if !bytes.Equal(receivedGophermap, []byte(testGophermap)) {
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
		if !file.IsDir() {
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

			if !bytes.Equal(receivedFile, diskFile) {
				// Dump out both files in hex if they don't match.
				t.Errorf("Received file %v does not match expected file:\n received:\n%v\n expected: see %v\n", file.Name(), string(receivedFile), file.Name())
			}
		}
	}
}
