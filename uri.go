package main

import (
	"net/url"
	"path/filepath"
	"strings"
	"unicode"
)

func uriToPath(uri string) string {
	switch {
	case strings.HasPrefix(uri, "file:///"):
		uri = uri[len("file://"):]
	case strings.HasPrefix(uri, "file://"):
		uri = uri[len("file:/"):]
	}

	if path, err := url.PathUnescape(uri); err == nil {
		uri = path
	}

	if isWindowsDriveURIPath(uri) {
		uri = strings.ToUpper(string(uri[1])) + uri[2:]
	}

	return filepath.FromSlash(uri)
}

func isWindowsDriveURIPath(uri string) bool {
	//nolint:gomnd
	if len(uri) < 4 {
		return false
	}

	return uri[0] == '/' && unicode.IsLetter(rune(uri[1])) && uri[2] == ':'
}
