package util

import (
	"net/url"
	"path/filepath"
	"strings"
)

func URIToPath(uri string) string {
	switch {
	case strings.HasPrefix(uri, "file:///"):
		uri = uri[len("file://"):]
	case strings.HasPrefix(uri, "file://"):
		uri = uri[len("file:/"):]
	}

	if path, err := url.PathUnescape(uri); err == nil {
		uri = path
	}

	return filepath.FromSlash(uri)
}
