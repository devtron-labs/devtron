package resource

import (
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"
)

// Resource represents a Kubernetes resource within a file
type Resource struct {
	Path   string
	Bytes  []byte
	sig    *Signature // Cache signature parsing
	sigErr error      // Cache potential signature parsing error
}

// Signature is a key representing a Kubernetes resource
type Signature struct {
	Kind, Version, Namespace, Name string
}

// GroupVersionKind returns a string with the GVK encoding of a resource signature.
// This encoding slightly differs from the Kubernetes upstream implementation
// in order to be suitable for being used in the kubeconform command-line arguments.
func (sig *Signature) GroupVersionKind() string {
	return fmt.Sprintf("%s/%s", sig.Version, sig.Kind)
}

// QualifiedName returns a string for a signature in the format version/kind/namespace/name
func (sig *Signature) QualifiedName() string {
	return fmt.Sprintf("%s/%s/%s/%s", sig.Version, sig.Kind, sig.Namespace, sig.Name)
}

// Signature computes a signature for a resource, based on its Kind, Version, Namespace & Name
func (res *Resource) Signature() (*Signature, error) {
	if res.sig != nil {
		return res.sig, res.sigErr
	}

	resource := struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name         string `yaml:"name"`
			Namespace    string `yaml:"namespace"`
			GenerateName string `yaml:"generateName"`
		} `yaml:"Metadata"`
	}{}
	err := yaml.Unmarshal(res.Bytes, &resource)

	name := resource.Metadata.Name
	if resource.Metadata.GenerateName != "" {
		name = resource.Metadata.GenerateName + "{{ generateName }}"
	}

	// We cache the result to not unmarshall every time we want to access the signature
	res.sig = &Signature{Kind: resource.Kind, Version: resource.APIVersion, Namespace: resource.Metadata.Namespace, Name: name}

	if err != nil { // Exit if there was an error unmarshalling
		res.sigErr = err
		return res.sig, res.sigErr
	}

	if res.sig.Kind == "" {
		res.sigErr = fmt.Errorf("missing 'kind' key")
		return res.sig, res.sigErr
	}

	if res.sig.Version == "" {
		res.sigErr = fmt.Errorf("missing 'apiVersion' key")
		return res.sig, res.sigErr
	}

	return res.sig, res.sigErr
}

func (res *Resource) SignatureFromMap(m map[string]interface{}) (*Signature, error) {
	if res.sig != nil {
		return res.sig, res.sigErr
	}

	Kind, ok := m["kind"].(string)
	if !ok {
		res.sigErr = fmt.Errorf("missing 'kind' key")
		return res.sig, res.sigErr
	}

	APIVersion, ok := m["apiVersion"].(string)
	if !ok {
		res.sigErr = fmt.Errorf("missing 'apiVersion' key")
		return res.sig, res.sigErr
	}

	var name, ns string
	Metadata, ok := m["metadata"].(map[string]interface{})
	if ok {
		name, _ = Metadata["name"].(string)
		ns, _ = Metadata["namespace"].(string)
		if _, ok := Metadata["generateName"].(string); ok {
			name = Metadata["generateName"].(string) + "{{ generateName }}"
		}
	}

	// We cache the result to not unmarshall every time we want to access the signature
	res.sig = &Signature{Kind: Kind, Version: APIVersion, Namespace: ns, Name: name}
	return res.sig, nil
}

// Resources returns a list of resources if the resource is of type List, a single resource otherwise
// See https://github.com/yannh/kubeconform/issues/53
func (res *Resource) Resources() []Resource {
	resources := []Resource{}
	if s, err := res.Signature(); err == nil && strings.ToLower(s.Kind) == "list" {
		// A single file of type List
		list := struct {
			Version string
			Kind    string
			Items   []interface{}
		}{}

		yaml.Unmarshal(res.Bytes, &list)

		for _, item := range list.Items {
			r := Resource{Path: res.Path}
			r.Bytes, _ = yaml.Marshal(item)
			resources = append(resources, r)
		}
		return resources
	}

	return []Resource{*res}
}
