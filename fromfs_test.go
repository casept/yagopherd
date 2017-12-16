package main

import (
	"testing"
)

// Prevent the compiler from optimizing out the benchmarked function.
var result gophermap

func BenchmarkDirToGophermap(b *testing.B) {
	// Prevent the compiler from optimizing out the benchmarked function.
	var r gophermap
	var err error
	for n := 0; n < b.N; n++ {
		r, err = dirToGophermap("./testdata/gopherroot/")
	}

	if err != nil {
		b.Fatal(err)
	}
	result = r
}
