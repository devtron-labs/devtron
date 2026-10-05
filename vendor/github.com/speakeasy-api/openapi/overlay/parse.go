package overlay

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ParseReader parses an overlay from the given reader.
func ParseReader(r io.Reader) (*Overlay, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read overlay data: %w", err)
	}

	var overlay Overlay
	if err := yaml.Unmarshal(data, &overlay); err != nil {
		return nil, err
	}

	return &overlay, nil
}

// Parse will parse the given reader as an overlay file.
func Parse(path string) (*Overlay, error) {
	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %q: %w", path, err)
	}

	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read overlay file at path %q: %w", path, err)
	}

	var overlay Overlay
	err = yaml.Unmarshal(data, &overlay)
	if err != nil {
		return nil, err
	}

	return &overlay, nil
}

// Format will validate reformat the given file
func Format(path string) error {
	overlay, err := Parse(path)
	if err != nil {
		return err
	}
	filePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to open overlay file at path %q: %w", path, err)
	}
	formatted, err := overlay.ToString()
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(formatted), 0600)
}

// Format writes the file back out as YAML.
func (o *Overlay) Format(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(o)
}
