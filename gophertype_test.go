package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

// Prevent the compiler from optimizing out the benchmarked function.
var reqResult string

func BenchmarkGophertype(b *testing.B) {
	var r string
	var err error
	// Benchmark several types of files to average out performance of varying filetypes.
	rawTestFiles, err := ioutil.ReadDir("./testdata/gopherroot")
	if err != nil {
		b.Fatal(err)
	}
	// Extract the names of files ASAP so it doesn't affect the actual benchmarking loop.
	var testFiles []string
	for _, file := range rawTestFiles {
		testFiles = append(testFiles, filepath.Join("./testdata/gopherroot", file.Name()))
	}

	for n := 0; n < b.N; n++ {
		for _, file := range testFiles {
			r, err = gophertype(file, true)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	reqResult = r
}
