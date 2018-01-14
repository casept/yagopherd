package main

import (
	"testing"
)

// Currently just used for benchmarks - actual functionality is tested as an integration test in a separate file.
var gophermapTC = []struct {
	req   req    // Request to return gophermaps for.
	bName string // How to name the sub-benchmark.
}{
	{req{selector: "", gopherP: false}, "GohperrootNonGopher+"},
	{req{selector: "", gopherP: true}, "GohperrootGopher+"},
	{req{selector: "", gopherP: false}, "SubdirNonGopher+"},
	{req{selector: "", gopherP: true}, "SubdirGopher+"},
}

// Prevent the compiler from optimizing out the benchmarked function.
var result gophermap

func BenchmarkReqToGophermap(b *testing.B) {
	var r gophermap
	var err error
	for i := 0; i <= len(gophermapTC)-1; i++ {
		tt := gophermapTC[i]
		b.Run(tt.bName, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				r, err = reqToGophermap(gophermapTC[i].req)
				if err != nil {
					b.Fatal(err)
				}
				// Prevent the compiler from optimizing out the benchmarked function.
				result = r
			}
		})
	}
}
