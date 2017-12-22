package main

import (
	"bytes"
	"testing"
)

func TestExtractGopherreq(t *testing.T) {
	tc := []struct {
		inputBytes []byte
		gopherP    bool
		selector   string
	}{
		{[]byte("\r\n"), false, ""},
		{[]byte("test.unknownfiletype\r\n"), false, "test.unknownfiletype"},
		{[]byte("\r\n\t+"), true, ""},
		{[]byte("test.unknownfiletype\r\n\t+"), true, "test.unknownfiletype"},
	}
	for _, testcase := range tc {
		req, err := extractReq(bytes.NewReader(testcase.inputBytes))
		if err != nil {
			t.Fatal(err)
		}
		if req.gopherP != testcase.gopherP {
			t.Fatalf("Gopher+ is %t, should be %t for input %q", req.gopherP, testcase.gopherP, string(testcase.inputBytes))
		}
		if req.selector != testcase.selector {
			t.Fatalf("Selector is %q, should be %q for input %q", req.selector, testcase.selector, string(testcase.inputBytes))
		}
	}
}
