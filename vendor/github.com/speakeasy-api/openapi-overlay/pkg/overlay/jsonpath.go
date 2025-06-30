package overlay

import (
	"fmt"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/config"
	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

type Queryable interface {
	Query(root *yaml.Node) []*yaml.Node
}

type yamlPathQueryable struct {
	path *yamlpath.Path
}

func (y yamlPathQueryable) Query(root *yaml.Node) []*yaml.Node {
	if y.path == nil {
		return []*yaml.Node{}
	}
	// errors aren't actually possible from yamlpath.
	result, _ := y.path.Find(root)
	return result
}

func (o *Overlay) NewPath(target string, warnings *[]string) (Queryable, error) {
	rfcJSONPath, rfcJSONPathErr := jsonpath.NewPath(target, config.WithPropertyNameExtension())
	if o.UsesRFC9535() {
		return rfcJSONPath, rfcJSONPathErr
	}
	if rfcJSONPathErr != nil && warnings != nil {
		*warnings = append(*warnings, fmt.Sprintf("invalid rfc9535 jsonpath %s: %s\nThis will be treated as an error in the future. Please fix and opt into the new implementation with `\"x-speakeasy-jsonpath\": rfc9535` in the root of your overlay. See overlay.speakeasy.com for an implementation playground.", target, rfcJSONPathErr.Error()))
	}

	path, err := yamlpath.NewPath(target)
	return mustExecute(path), err
}

func (o *Overlay) UsesRFC9535() bool {
	return o.JSONPathVersion == "rfc9535"
}

func mustExecute(path *yamlpath.Path) yamlPathQueryable {
	return yamlPathQueryable{path}
}
