<p align="center">
  <p align="center">
    <img  width="200px" alt="OpenAPI" src="https://github.com/user-attachments/assets/b9fa9c14-1c6f-4d8b-910f-15e5f962bab6">
  </p>
  <h1 align="center"><b>OpenAPI Overlay</b></h1>
  <p align="center">An implementation of the <a href="https://github.com/OAI/Overlay-Specification/blob/3f398c6/versions/1.0.0.md">OpenAPI Overlay Specification</a> for applying modifications to OpenAPI documents
</p>
  <p align="center">
    <!-- Overlay Reference badge -->
    <a href="https://speakeasy.com/openapi/overlays"><img alt="Overlay reference" src="https://www.speakeasy.com/assets/badges/overlay-reference.svg" /></a>
    <!-- Built By Speakeasy Badge -->
    <a href="https://speakeasy.com/"><img alt="Built by Speakeasy" src="https://www.speakeasy.com/assets/badges/built-by-speakeasy.svg" /></a>
    <a href="https://github.com/speakeasy-api/openapi/releases/latest"><img alt="Release" src="https://img.shields.io/github/release/speakeasy-api/openapi.svg?style=for-the-badge"></a>
    <a href="https://pkg.go.dev/github.com/speakeasy-api/openapi-overlay?tab=doc"><img alt="Go Doc" src="https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge"></a>
   <br />
    <a href="https://github.com/speakeasy-api/openapi/actions/workflows/test.yaml"><img alt="GitHub Action: Test" src="https://img.shields.io/github/actions/workflow/status/speakeasy-api/openapi/test.yaml?style=for-the-badge"></a>
    <a href="https://goreportcard.com/report/github.com/speakeasy-api/openapi-overlay"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/speakeasy-api/openapi-overlay?style=for-the-badge"></a>
    <a href="/LICENSE"><img alt="Software License" src="https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge"></a>
   <br />
    <a href="https://join.slack.com/t/speakeasy-dev/shared_invite/zt-1cwb3flxz-lS5SyZxAsF_3NOq5xc8Cjw"><img src="https://img.shields.io/static/v1?label=Slack&message=Join&color=7289da&style=for-the-badge" /></a>
  </p>
</p>

> ⚠️ This an alpha implementation. If you'd like to discuss a production use case please join the Speakeasy [slack](https://join.slack.com/t/speakeasy-dev/shared_invite/zt-1df0lalk5-HCAlpcQiqPw8vGukQWhexw).

## Features

- **OpenAPI Overlay Specification Compliance**: Full implementation of the [OpenAPI Overlay Specification](https://github.com/OAI/Overlay-Specification/blob/3f398c6/versions/1.0.0.md) (2023-10-12) and [version 1.1.0](https://github.com/OAI/Overlay-Specification/blob/e2c3cec/versions/1.1.0-dev.md)
- **JSONPath Target Selection**: Uses JSONPath expressions to select nodes for modification
- **Remove, Update, and Copy Actions**: Support for remove actions (pruning nodes), update actions (merging values), and copy actions (duplicating or moving nodes)
- **Flexible Input/Output**: Works with both YAML and JSON formats
- **Batch Operations**: Apply multiple modifications to large numbers of nodes in a single operation
- **YAML v1.2 Support**: Uses [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) for YAML v1.2 parsing (superset of JSON)

## About OpenAPI Overlays

This specification defines a means of editing an OpenAPI Specification file by applying a list of actions. Each action is either a remove action that prunes nodes or an update that merges a value into nodes. The nodes impacted are selected by a target expression which uses JSONPath. This implementation also supports [version 1.1.0](https://github.com/OAI/Overlay-Specification/blob/e2c3cec/versions/1.1.0-dev.md) which adds a `copy` action for duplicating or moving nodes within the document.

The specification itself says very little about the input file to be modified or the output file. The presumed intention is that the input and output be an OpenAPI Specification, but that is not required.

In many ways, this is similar to [JSONPatch](https://jsonpatch.com/), but without the requirement to use a single explicit path for each operation. This allows the creator of an overlay file to apply a single modification to a large number of nodes in the file within a single operation.

<!-- START USAGE EXAMPLES -->

## Apply an overlay to an OpenAPI document

Shows loading an overlay specification and applying it to transform an OpenAPI document.

```go
overlayContent := `overlay: 1.0.0
info:
  title: Pet Store Enhancement Overlay
  version: 1.0.0
actions:
  - target: $.info.description
    update: Enhanced pet store API with additional features`

openAPIContent := `openapi: 3.1.0
info:
  title: Pet Store API
  version: 1.0.0
  description: A simple pet store API
paths:
  /pets:
    get:
      summary: List pets
      responses:
        '200':
          description: A list of pets`

overlayFile := "temp_overlay.yaml"
openAPIFile := "temp_openapi.yaml"
if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
	panic(err)
}
if err := os.WriteFile(openAPIFile, []byte(openAPIContent), 0644); err != nil {
	panic(err)
}
defer os.Remove(overlayFile)
defer os.Remove(openAPIFile)

overlayDoc, err := overlay.Parse(overlayFile)
if err != nil {
	panic(err)
}

openAPINode, err := loader.LoadSpecification(openAPIFile)
if err != nil {
	panic(err)
}

err = overlayDoc.ApplyTo(openAPINode)
if err != nil {
	panic(err)
}

// Convert back to YAML string
var buf strings.Builder
encoder := yaml.NewEncoder(&buf)
encoder.SetIndent(2)
err = encoder.Encode(openAPINode)
if err != nil {
	panic(err)
}

fmt.Printf("Transformed document:\n%s", buf.String())
```

## Create an overlay specification programmatically

Shows building an overlay specification with update and remove actions.

```go
// Create update value as yaml.Node
var updateNode yaml.Node
updateNode.SetString("Enhanced API with additional features")

overlayDoc := &overlay.Overlay{
	Version: "1.0.0",
	Info: overlay.Info{
		Title:   "API Enhancement Overlay",
		Version: "1.0.0",
	},
	Actions: []overlay.Action{
		{
			Target: "$.info.description",
			Update: updateNode,
		},
		{
			Target: "$.paths['/deprecated-endpoint']",
			Remove: true,
		},
	},
}

result, err := overlayDoc.ToString()
if err != nil {
	panic(err)
}

fmt.Printf("Overlay specification:\n%s", result)
```

## Parse an overlay specification from a file

Shows loading an overlay file and accessing its properties.

```go
overlayContent := `overlay: 1.0.0
info:
  title: API Modification Overlay
  version: 1.0.0
actions:
  - target: $.info.title
    update: Enhanced Pet Store API
  - target: $.info.version
    update: 2.0.0`

overlayFile := "temp_overlay.yaml"
if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
	panic(err)
}
defer func() { _ = os.Remove(overlayFile) }()

overlayDoc, err := overlay.Parse(overlayFile)
if err != nil {
	panic(err)
}

fmt.Printf("Overlay Version: %s\n", overlayDoc.Version)
fmt.Printf("Title: %s\n", overlayDoc.Info.Title)
fmt.Printf("Number of Actions: %d\n", len(overlayDoc.Actions))

for i, action := range overlayDoc.Actions {
	fmt.Printf("Action %d Target: %s\n", i+1, action.Target)
}
```

## Validate an overlay specification

Shows loading and validating an overlay specification for correctness.

```go
invalidOverlay := `overlay: 1.0.0
info:
  title: Invalid Overlay
actions:
  - target: $.info.title
    description: Missing update or remove`

overlayFile := "temp_invalid_overlay.yaml"
if err := os.WriteFile(overlayFile, []byte(invalidOverlay), 0644); err != nil {
	panic(err)
}
defer func() { _ = os.Remove(overlayFile) }()

overlayDoc, err := overlay.Parse(overlayFile)
if err != nil {
	fmt.Printf("Parse error: %s\n", err.Error())
	return
}

validationErr := overlayDoc.Validate()
if validationErr != nil {
	fmt.Println("Validation errors:")
	fmt.Printf("  %s\n", validationErr.Error())
} else {
	fmt.Println("Overlay specification is valid!")
}
```

## Use remove actions in overlays

Shows removing specific paths and properties from an OpenAPI document.

```go
openAPIContent := `openapi: 3.1.0
info:
  title: API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
  /users/{id}:
    get:
      summary: Get user
  /admin:
    get:
      summary: Admin endpoint
      deprecated: true`

overlayContent := `overlay: 1.0.0
info:
  title: Cleanup Overlay
  version: 1.0.0
actions:
  - target: $.paths['/admin']
    remove: true`

openAPIFile := "temp_openapi.yaml"
overlayFile := "temp_overlay.yaml"
if err := os.WriteFile(openAPIFile, []byte(openAPIContent), 0644); err != nil {
	panic(err)
}
if err := os.WriteFile(overlayFile, []byte(overlayContent), 0644); err != nil {
	panic(err)
}
defer func() { _ = os.Remove(openAPIFile) }()
defer func() { _ = os.Remove(overlayFile) }()

overlayDoc, err := overlay.Parse(overlayFile)
if err != nil {
	panic(err)
}

openAPINode, err := loader.LoadSpecification(openAPIFile)
if err != nil {
	panic(err)
}

err = overlayDoc.ApplyTo(openAPINode)
if err != nil {
	panic(err)
}

var buf strings.Builder
encoder := yaml.NewEncoder(&buf)
encoder.SetIndent(2)
err = encoder.Encode(openAPINode)
if err != nil {
	panic(err)
}

fmt.Printf("Document after removing deprecated endpoint:\n%s", buf.String())
```

<!-- END USAGE EXAMPLES -->

## Contributing

This repository is maintained by Speakeasy, but we welcome and encourage contributions from the community to help improve its capabilities and stability.

### How to Contribute

1. **Open Issues**: Found a bug or have a feature suggestion? Open an issue to describe what you'd like to see changed.

2. **Pull Requests**: We welcome pull requests! If you'd like to contribute code:
   - Fork the repository
   - Create a new branch for your feature/fix
   - Submit a PR with a clear description of the changes and any related issues

3. **Feedback**: Share your experience using the packages or suggest improvements.

All contributions, whether they're bug reports, feature requests, or code changes, help make this project better for everyone.

Please ensure your contributions adhere to our coding standards and include appropriate tests where applicable.
