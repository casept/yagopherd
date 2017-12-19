package main

import (
	"os"
	"path/filepath"
	"strings"
)

// gophertype gets the gopher response code via file extension.
// Might get replaced with MIME type sniffing in the future, although it will be slower due to having to open the file to check magic numbers (and possibly require cgo).
// Some clients might also depend on file extensions.
func gophertype(path string, gopherP bool) (gophertype string, err error) {
	file, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if file.IsDir() == true {
		return "1", nil
	}
	fileExt := strings.ToLower(filepath.Ext(path))

	// TODO: Let user map types in server config
	gophertypes := make(map[string]string)
	//Image formats
	// Gopher and Gopher+ specify different image identifiers
	var picID string
	if gopherP {
		picID = "I"
	} else {
		picID = ":"
	}
	gophertypes[".jpg"] = picID
	gophertypes[".jpeg"] = picID
	gophertypes[".jp2"] = picID
	gophertypes[".jpx"] = picID
	gophertypes[".tiff"] = picID
	gophertypes[".tif"] = picID
	gophertypes[".bmp"] = picID
	gophertypes[".png"] = picID
	gophertypes[".webp"] = picID
	gophertypes[".pbm"] = picID
	gophertypes[".pgm"] = picID
	gophertypes[".ppm"] = picID
	gophertypes[".pnm"] = picID
	gophertypes["heif"] = picID
	gophertypes[".heic"] = picID
	gophertypes[".bpg"] = picID
	gophertypes[".ecw"] = picID
	gophertypes[".fits"] = picID
	gophertypes[".fit"] = picID
	gophertypes[".fts"] = picID
	gophertypes[".flif"] = picID
	gophertypes[".ico"] = picID
	gophertypes[".jxr"] = picID
	gophertypes[".hdp"] = picID
	gophertypes[".svg"] = picID

	// Spec treats GIF specially
	gophertypes[".gif"] = "g"

	// DOS/Windows executables
	gophertypes[".com"] = "5"
	gophertypes[".exe"] = "5"

	// Text files of various kinds
	// TODO: Expand this list
	gophertypes[".txt"] = "0"
	gophertypes[".md"] = "0"
	gophertypes[".rtf"] = "0"

	// Additional gopher+ types
	if gopherP {
		// Video
		gophertypes[".mp4"] = ";"
		gophertypes[".m4v"] = ";"
		gophertypes[".mkv"] = ";"
		gophertypes[".webm"] = ";"
		gophertypes[".mov"] = ";"
		gophertypes[".avi"] = ";"
		gophertypes[".wmv"] = ";"
		gophertypes[".mpg"] = ";"
		gophertypes[".flv"] = ";"

		// Audio
		gophertypes[".mp3"] = "<"
		gophertypes[".mid"] = "<"
		gophertypes[".m4a"] = "<"
		gophertypes[".ogg"] = "<"
		gophertypes[".flac"] = "<"
		gophertypes[".wav"] = "<"
		gophertypes[".amr"] = "<"
	}

	// Check if a value is in the map
	// See https://stackoverflow.com/questions/2050391/how-to-check-if-a-map-contains-a-key-in-go
	if gophertype, exists := gophertypes[fileExt]; exists {
		return gophertype, nil
	}

	// Other types of files
	// We assume unknown files are binaries so the user is prompted to download them instead of seeing a bunch of gibberish
	return "9", nil
}
