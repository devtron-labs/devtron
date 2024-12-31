package registry

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type LocalRegistry struct {
	pathTemplate string
	strict       bool
	debug        bool
}

// NewLocalSchemas creates a new "registry", that will serve schemas from files, given a list of schema filenames
func newLocalRegistry(pathTemplate string, strict bool, debug bool) (*LocalRegistry, error) {
	return &LocalRegistry{
		pathTemplate,
		strict,
		debug,
	}, nil
}

// DownloadSchema retrieves the schema from a file for the resource
func (r LocalRegistry) DownloadSchema(resourceKind, resourceAPIVersion, k8sVersion string) ([]byte, error) {
	schemaFile, err := schemaPath(r.pathTemplate, resourceKind, resourceAPIVersion, k8sVersion, r.strict)
	if err != nil {
		return []byte{}, nil
	}
	f, err := os.Open(schemaFile)
	if err != nil {
		if os.IsNotExist(err) {
			msg := fmt.Sprintf("could not open file %s", schemaFile)
			if r.debug {
				log.Print(msg)
			}
			return nil, newNotFoundError(errors.New(msg))
		}

		msg := fmt.Sprintf("failed to open schema at %s: %s", schemaFile, err)
		if r.debug {
			log.Print(msg)
		}
		return nil, errors.New(msg)
	}

	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		msg := fmt.Sprintf("failed to read schema at %s: %s", schemaFile, err)
		if r.debug {
			log.Print(msg)
		}
		return nil, err
	}

	if r.debug {
		log.Printf("using schema found at %s", schemaFile)
	}
	return content, nil
}
