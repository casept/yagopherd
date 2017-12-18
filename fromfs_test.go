package main

import (
	"testing"
)

// Prevent the compiler from optimizing out the benchmarked function.
var result gophermap

func BenchmarkDirToGophermap(b *testing.B) {
	var r gophermap
	var err error
	for n := 0; n < b.N; n++ {
		r, err = dirToGophermap("./testdata/gopherroot/", true)
	}

	if err != nil {
		b.Fatal(err)
	}
	// Prevent the compiler from optimizing out the benchmarked function.
	result = r
}
