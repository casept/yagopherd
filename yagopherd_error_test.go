package main

import (
	"io/ioutil"
	"net"
	"testing"
)

// Try requesting nonexistent files
// This is technically an integration test, but I think it's more logical to put it in a separate file rather than cluttering yagopherd_test.go.

func TestFileNotFoundError(t *testing.T) {
	tc := []struct {
		selector  string
		errorResp string
	}{
		// gopher+ client
		{"thisfiledoesnotexist.nonexistent\r\n\t+", "--1\r\n" + "1" + " [fakeadmin@fakegopherhole.example.com]" + "\r\n" + "The item thisfiledoesnotexist.nonexistent could not be found on this server." + "\r\n.\r\n"},
		// gopher client
		{"thisfiledoesnotexist.nonexistent\r\n", "3The item thisfiledoesnotexist.nonexistent could not be found on this server. Server admin: [fakeadmin@fakegopherhole.example.com].\r\n."},
	}
	for _, tt := range tc {
		conn, err := net.DialTCP("tcp", nil, Addr)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.Write([]byte(tt.selector))
		rcv, err := ioutil.ReadAll(conn)
		if err != nil {
			t.Fatal(err)
		}
		rcvStr := string(rcv[:])
		if rcvStr != tt.errorResp {
			t.Fatalf("Expected error message does not match received error message for selector %q:\nreceived:\n%v\nexpected:\n%v\n", tt.selector, rcvStr, tt.errorResp)
		}
		conn.Close()
	}
}
