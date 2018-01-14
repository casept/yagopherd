package main

import (
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
	"github.com/spf13/viper"
)

func TestConstructGopherItem(t *testing.T) {
	testfile, err := filepath.Abs("testdata/gopherroot/test.txt")
	if err != nil {
		t.Fatal(err)
	}

	tc := []struct {
		selector string
		gopherP  bool
		out      gopherItem
	}{
		{"/test.txt", false, gopherItem{"0", "test.txt", "/test.txt", viper.GetString("host"), viper.GetInt("port"), true, false, testfile, "", false}},
		{"/test", false, gopherItem{"1", "test", "/test", viper.GetString("host"), viper.GetInt("port"), true, false, testfile, "", false}},
		{"/test.txt", true, gopherItem{"0", "test.txt", "/test.txt", viper.GetString("host"), viper.GetInt("port"), true, true, testfile, "", true}},
		{"/test", true, gopherItem{"1", "test", "/test", viper.GetString("host"), viper.GetInt("port"), true, true, testfile, "", true}},
	}

	for _, tt := range tc {
		item, err := constructGopherItem(tt.selector, tt.gopherP)
		if err != nil {
			t.Fatal(err)
		}
		if diff := deep.Equal(item, tt.out); diff != nil {
			t.Error(diff)
		}
	}
}
