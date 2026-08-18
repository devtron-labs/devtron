package loader

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/speakeasy-api/openapi/overlay"
	"gopkg.in/yaml.v3"
)

// GetOverlayExtendsPath returns the path to file if the extends URL is a file
// URL. Otherwise, returns an empty string and an error. The error may occur if
// no extends URL is present or if the URL is not a file URL or if the URL is
// malformed.
func GetOverlayExtendsPath(o *overlay.Overlay) (string, error) {
	if o.Extends == "" {
		return "", errors.New("overlay does not specify an extends URL")
	}

	// Handle Windows file paths that might be formatted as file URLs
	// file:///C:/path or file://C:/path on Windows need special handling
	if runtime.GOOS == "windows" && strings.HasPrefix(o.Extends, "file://") {
		// Remove the file:// or file:/// prefix
		path := strings.TrimPrefix(o.Extends, "file:///")
		if path == o.Extends {
			path = strings.TrimPrefix(o.Extends, "file://")
		}

		// If it looks like a Windows path (e.g., C:/... or C:\...), use it directly
		if len(path) >= 2 && path[1] == ':' {
			// Convert forward slashes to backslashes for Windows
			path = filepath.FromSlash(path)
			return path, nil
		}
	}

	specUrl, err := url.Parse(o.Extends)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %q: %w", o.Extends, err)
	}

	if specUrl.Scheme != "file" {
		return "", fmt.Errorf("only file:// extends URLs are supported, not %q", o.Extends)
	}

	// On Windows, url.Parse().Path for file:///C:/path returns /C:/path
	// We need to strip the leading slash for Windows absolute paths
	path := specUrl.Path
	if runtime.GOOS == "windows" && len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = path[1:] // Remove leading slash
	}

	return path, nil
}

// LoadExtendsSpecification will load and parse a YAML or JSON file as specified
// in the extends parameter of the overlay. Currently, this only supports file
// URLs.
func LoadExtendsSpecification(o *overlay.Overlay) (*yaml.Node, error) {
	path, err := GetOverlayExtendsPath(o)
	if err != nil {
		return nil, err
	}

	return LoadSpecification(path)
}

// LoadSpecificationFromReader parses a YAML or JSON specification from the given reader.
func LoadSpecificationFromReader(r io.Reader) (*yaml.Node, error) {
	var ys yaml.Node
	if err := yaml.NewDecoder(r).Decode(&ys); err != nil {
		return nil, fmt.Errorf("failed to parse specification from reader: %w", err)
	}

	return &ys, nil
}

// LoadSpecification will load and parse a YAML or JSON file from the given path.
func LoadSpecification(path string) (*yaml.Node, error) {
	rs, err := os.Open(path) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open schema from path %q: %w", path, err)
	}
	defer rs.Close()

	var ys yaml.Node
	err = yaml.NewDecoder(rs).Decode(&ys)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema at path %q: %w", path, err)
	}

	return &ys, nil
}

// LoadEitherSpecification is a convenience function that will load a
// specification from the given file path if it is non-empty. Otherwise, it will
// attempt to load the path from the overlay's extends URL. Also returns the name
// of the file loaded.
func LoadEitherSpecification(path string, o *overlay.Overlay) (*yaml.Node, string, error) {
	var (
		y   *yaml.Node
		err error
	)

	if path != "" {
		y, err = LoadSpecification(path)
	} else {
		path, _ = GetOverlayExtendsPath(o)
		y, err = LoadExtendsSpecification(o)
	}

	return y, path, err
}
