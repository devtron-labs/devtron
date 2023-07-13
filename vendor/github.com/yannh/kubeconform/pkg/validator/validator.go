// This is the main package to import to embed kubeconform in your software
package validator

import (
	"context"
	"fmt"
	"io"

	"github.com/yannh/kubeconform/pkg/cache"
	"github.com/yannh/kubeconform/pkg/registry"
	"github.com/yannh/kubeconform/pkg/resource"

	"github.com/xeipuuv/gojsonschema"
	"sigs.k8s.io/yaml"
)

// Different types of validation results
type Status int

const (
	_       Status = iota
	Error          // an error occurred processing the file / resource
	Skipped        // resource has been skipped, for example if its Kind was part of the kinds to skip
	Valid          // resource is valid
	Invalid        // resource is invalid
	Empty          // resource is empty. Note: is triggered for files starting with a --- separator.
)

// Result contains the details of the result of a resource validation
type Result struct {
	Resource resource.Resource
	Err      error
	Status   Status
}

// Validator exposes multiple methods to validate your Kubernetes resources.
type Validator interface {
	ValidateResource(res resource.Resource) Result
	Validate(filename string, r io.ReadCloser) []Result
	ValidateWithContext(ctx context.Context, filename string, r io.ReadCloser) []Result
}

// Opts contains a set of options for the validator.
type Opts struct {
	Cache                string              // Cache schemas downloaded via HTTP to this folder
	Debug                bool                // Debug infos will be print here
	SkipTLS              bool                // skip TLS validation when downloading from an HTTP Schema Registry
	SkipKinds            map[string]struct{} // List of resource Kinds to ignore
	RejectKinds          map[string]struct{} // List of resource Kinds to reject
	KubernetesVersion    string              // Kubernetes Version - has to match one in https://github.com/instrumenta/kubernetes-json-schema
	Strict               bool                // thros an error if resources contain undocumented fields
	IgnoreMissingSchemas bool                // skip a resource if no schema for that resource can be found
}

// New returns a new Validator
func New(schemaLocations []string, opts Opts) (Validator, error) {
	// Default to our kubernetes-json-schema fork
	// raw.githubusercontent.com is frontend by Fastly and very fast
	if len(schemaLocations) == 0 {
		schemaLocations = []string{"https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json"}
	}

	registries := []registry.Registry{}
	for _, schemaLocation := range schemaLocations {
		reg, err := registry.New(schemaLocation, opts.Cache, opts.Strict, opts.SkipTLS, opts.Debug)
		if err != nil {
			return nil, err
		}
		registries = append(registries, reg)
	}

	if opts.KubernetesVersion == "" {
		opts.KubernetesVersion = "master"
	}

	if opts.SkipKinds == nil {
		opts.SkipKinds = map[string]struct{}{}
	}
	if opts.RejectKinds == nil {
		opts.RejectKinds = map[string]struct{}{}
	}

	return &v{
		opts:           opts,
		schemaDownload: downloadSchema,
		schemaCache:    cache.NewInMemoryCache(),
		regs:           registries,
	}, nil
}

type v struct {
	opts           Opts
	schemaCache    cache.Cache
	schemaDownload func(registries []registry.Registry, kind, version, k8sVersion string) (*gojsonschema.Schema, error)
	regs           []registry.Registry
}

// ValidateResource validates a single resource. This allows to validate
// large resource streams using multiple Go Routines.
func (val *v) ValidateResource(res resource.Resource) Result {
	// For backward compatibility reasons when determining whether
	// a resource should be skipped or rejected we use both
	// the GVK encoding of the resource signatures (the recommended method
	// for skipping/rejecting resources) and the raw Kind.

	skip := func(signature resource.Signature) bool {
		if _, ok := val.opts.SkipKinds[signature.GroupVersionKind()]; ok {
			return ok
		}
		_, ok := val.opts.SkipKinds[signature.Kind]
		return ok
	}

	reject := func(signature resource.Signature) bool {
		if _, ok := val.opts.RejectKinds[signature.GroupVersionKind()]; ok {
			return ok
		}
		_, ok := val.opts.RejectKinds[signature.Kind]
		return ok
	}

	if len(res.Bytes) == 0 {
		return Result{Resource: res, Err: nil, Status: Empty}
	}

	var r map[string]interface{}
	unmarshaller := yaml.Unmarshal
	if val.opts.Strict {
		unmarshaller = yaml.UnmarshalStrict
	}

	if err := unmarshaller(res.Bytes, &r); err != nil {
		return Result{Resource: res, Status: Error, Err: fmt.Errorf("error unmarshalling resource: %s", err)}
	}

	if r == nil { // Resource is empty
		return Result{Resource: res, Err: nil, Status: Empty}
	}

	sig, err := res.SignatureFromMap(r)
	if err != nil {
		return Result{Resource: res, Err: fmt.Errorf("error while parsing: %s", err), Status: Error}
	}

	if skip(*sig) {
		return Result{Resource: res, Err: nil, Status: Skipped}
	}

	if reject(*sig) {
		return Result{Resource: res, Err: fmt.Errorf("prohibited resource kind %s", sig.Kind), Status: Error}
	}

	cached := false
	var schema *gojsonschema.Schema

	if val.schemaCache != nil {
		s, err := val.schemaCache.Get(sig.Kind, sig.Version, val.opts.KubernetesVersion)
		if err == nil {
			cached = true
			schema = s.(*gojsonschema.Schema)
		}
	}

	if !cached {
		if schema, err = val.schemaDownload(val.regs, sig.Kind, sig.Version, val.opts.KubernetesVersion); err != nil {
			return Result{Resource: res, Err: err, Status: Error}
		}

		if val.schemaCache != nil {
			val.schemaCache.Set(sig.Kind, sig.Version, val.opts.KubernetesVersion, schema)
		}
	}

	if schema == nil {
		if val.opts.IgnoreMissingSchemas {
			return Result{Resource: res, Err: nil, Status: Skipped}
		}

		return Result{Resource: res, Err: fmt.Errorf("could not find schema for %s", sig.Kind), Status: Error}
	}

	resourceLoader := gojsonschema.NewGoLoader(r)

	results, err := schema.Validate(resourceLoader)
	if err != nil {
		// This error can only happen if the Object to validate is poorly formed. There's no hope of saving this one
		return Result{Resource: res, Status: Error, Err: fmt.Errorf("problem validating schema. Check JSON formatting: %s", err)}
	}

	if results.Valid() {
		return Result{Resource: res, Status: Valid}
	}

	msg := ""
	for _, errMsg := range results.Errors() {
		if msg != "" {
			msg += " - "
		}
		details := errMsg.Details()
		msg += fmt.Sprintf("For field %s: %s", details["field"].(string), errMsg.Description())
	}

	return Result{Resource: res, Status: Invalid, Err: fmt.Errorf("%s", msg)}
}

// ValidateWithContext validates resources found in r
// filename should be a name for the stream, such as a filename or stdin
func (val *v) ValidateWithContext(ctx context.Context, filename string, r io.ReadCloser) []Result {
	validationResults := []Result{}
	resourcesChan, _ := resource.FromStream(ctx, filename, r)
	for {
		select {
		case res, ok := <-resourcesChan:
			validationResults = append(validationResults, val.ValidateResource(res))
			if !ok {
				resourcesChan = nil
			}

		case <-ctx.Done():
			break
		}

		if resourcesChan == nil {
			break
		}
	}

	r.Close()
	return validationResults
}

// Validate validates resources found in r
// filename should be a name for the stream, such as a filename or stdin
func (val *v) Validate(filename string, r io.ReadCloser) []Result {
	return val.ValidateWithContext(context.Background(), filename, r)
}

func downloadSchema(registries []registry.Registry, kind, version, k8sVersion string) (*gojsonschema.Schema, error) {
	var err error
	var schemaBytes []byte

	for _, reg := range registries {
		schemaBytes, err = reg.DownloadSchema(kind, version, k8sVersion)
		if err == nil {
			schema, err := gojsonschema.NewSchema(gojsonschema.NewBytesLoader(schemaBytes))

			// If we got a non-parseable response, we try the next registry
			if err != nil {
				continue
			}
			return schema, err
		}

		// If we get a 404, we try the next registry, but we exit if we get a real failure
		if _, notfound := err.(*registry.NotFoundError); notfound {
			continue
		}

		return nil, err
	}

	return nil, nil // No schema found - we don't consider it an error, resource will be skipped
}
