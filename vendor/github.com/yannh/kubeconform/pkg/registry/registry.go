package registry

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Manifest struct {
	Kind, Version string
}

// Registry is an interface that should be implemented by any source of Kubernetes schemas
type Registry interface {
	DownloadSchema(resourceKind, resourceAPIVersion, k8sVersion string) ([]byte, error)
}

// Retryable indicates whether an error is a temporary or a permanent failure
type Retryable interface {
	IsNotFound() bool
}

// NotFoundError is returned when the registry does not contain a schema for the resource
type NotFoundError struct {
	err error
}

func newNotFoundError(err error) *NotFoundError {
	return &NotFoundError{err}
}
func (e *NotFoundError) Error() string   { return e.err.Error() }
func (e *NotFoundError) Retryable() bool { return false }

func schemaPath(tpl, resourceKind, resourceAPIVersion, k8sVersion string, strict bool) (string, error) {
	normalisedVersion := k8sVersion
	if normalisedVersion != "master" {
		normalisedVersion = "v" + normalisedVersion
	}

	strictSuffix := ""
	if strict {
		strictSuffix = "-strict"
	}

	groupParts := strings.Split(resourceAPIVersion, "/")
	versionParts := strings.Split(groupParts[0], ".")

	kindSuffix := "-" + strings.ToLower(versionParts[0])
	if len(groupParts) > 1 {
		kindSuffix += "-" + strings.ToLower(groupParts[1])
	}

	tmpl, err := template.New("tpl").Parse(tpl)
	if err != nil {
		return "", err
	}

	tplData := struct {
		NormalizedKubernetesVersion string
		StrictSuffix                string
		ResourceKind                string
		ResourceAPIVersion          string
		Group                       string
		KindSuffix                  string
	}{
		normalisedVersion,
		strictSuffix,
		strings.ToLower(resourceKind),
		groupParts[len(groupParts)-1],
		groupParts[0],
		kindSuffix,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, tplData)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func New(schemaLocation string, cache string, strict bool, skipTLS bool, debug bool) (Registry, error) {
	if schemaLocation == "default" {
		schemaLocation = "https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json"
	} else if !strings.HasSuffix(schemaLocation, "json") { // If we dont specify a full templated path, we assume the paths of our fork of kubernetes-json-schema
		schemaLocation += "/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json"
	}

	// try to compile the schemaLocation template to ensure it is valid
	if _, err := schemaPath(schemaLocation, "Deployment", "v1", "master", true); err != nil {
		return nil, fmt.Errorf("failed initialising schema location registry: %s", err)
	}

	if strings.HasPrefix(schemaLocation, "http") {
		return newHTTPRegistry(schemaLocation, cache, strict, skipTLS, debug)
	}

	return newLocalRegistry(schemaLocation, strict, debug)
}
