package loader

import (
	"fmt"
	"io"

	"github.com/speakeasy-api/openapi/overlay"
)

// LoadOverlay is a tool for loading and parsing an overlay file from the file
// system.
func LoadOverlay(path string) (*overlay.Overlay, error) {
	o, err := overlay.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay from path %q: %w", path, err)
	}

	return o, nil
}

// LoadOverlayFromReader parses an overlay from the given reader.
func LoadOverlayFromReader(r io.Reader) (*overlay.Overlay, error) {
	o, err := overlay.ParseReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay from reader: %w", err)
	}

	return o, nil
}
