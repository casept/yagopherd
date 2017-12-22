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

func TestGopherPGophermap(t *testing.T) {
	conn, err := net.DialTCP("tcp", nil, Addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err = conn.Write([]byte("\r\n\t+"))
	if err != nil {
		t.Fatal(err)
	}
	receivedGophermap, err := ioutil.ReadAll(conn)
	if err != nil {
		t.Fatalf("Error while trying to read server response: %v", err)
	}
	// The ./testdata/gopherpgophermap file contains a known-good gophermap of the test data directory.
	var rawDiskFile []byte
	rawDiskFile, err = ioutil.ReadFile("./testdata/gopherpgophermap")
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
