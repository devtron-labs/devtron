package overlay

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// Version constants for the Overlay specification
const (
	// LatestVersion is the latest supported overlay version
	LatestVersion = "1.1.0"
	// Version100 is the Overlay 1.0.0 version
	Version100 = "1.0.0"
	// Version110 is the Overlay 1.1.0 version
	Version110 = "1.1.0"
)

// JSONPath implementation constants
const (
	// JSONPathRFC9535 enables RFC 9535 JSONPath implementation
	JSONPathRFC9535 = "rfc9535"
	// JSONPathLegacy enables legacy yamlpath implementation (for backward compatibility)
	JSONPathLegacy = "legacy"
)

// Extensible provides a place for extensions to be added to components of the
// Overlay configuration. These are  a map from x-* extension fields to their values.
type Extensions map[string]any

// Overlay is the top-level configuration for an OpenAPI overlay.
type Overlay struct {
	Extensions `yaml:"-,inline"`

	// Version is the version of the overlay configuration (1.0.0 or 1.1.0)
	Version string `yaml:"overlay"`

	// JSONPathVersion controls the JSONPath implementation used.
	// For version 1.0.0: default is legacy, use "rfc9535" to opt-in to RFC 9535
	// For version 1.1.0: default is RFC 9535, use "legacy" to opt-out
	JSONPathVersion string `yaml:"x-speakeasy-jsonpath,omitempty"`

	// Info describes the metadata for the overlay.
	Info Info `yaml:"info"`

	// Extends is a URL to the OpenAPI specification this overlay applies to.
	Extends string `yaml:"extends,omitempty"`

	// Actions is the list of actions to perform to apply the overlay.
	Actions []Action `yaml:"actions"`
}

func (o *Overlay) ToString() (string, error) {
	buf := bytes.NewBuffer([]byte{})
	decoder := yaml.NewEncoder(buf)
	decoder.SetIndent(2)
	err := decoder.Encode(o)
	return buf.String(), err
}

// Info describes the metadata for the overlay.
type Info struct {
	Extensions `yaml:"-,inline"`

	// Title is the title of the overlay.
	Title string `yaml:"title"`

	// Version is the version of the overlay.
	Version string `yaml:"version"`

	// Description is an optional description of the overlay (new in Overlay 1.1.0).
	Description string `yaml:"description,omitempty"`
}

type Action struct {
	Extensions `yaml:"-,inline"`

	// Target is the JSONPath to the target of the action.
	Target string `yaml:"target"`

	// Description is a description of the action.
	Description string `yaml:"description,omitempty"`

	// Update is the sub-document to use to merge or replace in the target. This is
	// ignored if Remove is set.
	Update yaml.Node `yaml:"update,omitempty"`

	// Remove marks the target node for removal rather than update.
	Remove bool `yaml:"remove,omitempty"`

	// Copy is a JSONPath to the source node to copy to the target. This is
	// mutually exclusive with Update and Remove.
	Copy string `yaml:"copy,omitempty"`
}
