package sanitize

import (
	"path/filepath"
	"testing"
)

var tc = []struct {
	path string
	sane bool
}{
	{"../../../", false},
	{"/etc/passwd", false},
	{"../etc/passwd/", false},
	{"testdir/cat/catpicture", true},
	{"./testdir/cat/catpicture", true},
}

func TestIsSane(t *testing.T) {
	for _, tt := range tc {
		rootPath, err := filepath.Abs("testdir/")
		if err != nil {
			t.Fatal(err)
		}
		sane, err := IsSane(rootPath, tt.path)
		if err != nil {
			t.Fatal(err)
		}
		if sane != tt.sane {
			t.Errorf("Sane is %v, should be %v for path %v", sane, tt.sane, tt.path)
		}
	}
}

func TestIsSaneAbsOnly(t *testing.T) {
	rootPath := "testdata/"
	_, err := IsSane(rootPath, "./")
	if err == nil {
		t.Error("Did not get error for supplying a relative root path")
	}
}

func TesErrorIfNotSane(t *testing.T) {
	for _, tt := range tc {
		rootPath, err := filepath.Abs("testdir/")
		if err != nil {
			t.Fatal(err)
		}
		err = ErrorIfNotSane(rootPath, tt.path)
		switch err.(type) {
		case nil:
			if tt.sane == false {
				t.Errorf("expected error to be returned for unsanitary path %v, did not get one", tt.path)
			}
		case PathNotSaneError:
			if tt.sane == true {
				t.Errorf("did not expect PathNotSaneError, got one for path %v", tt.path)
			}
		}

	}
}
