# Developer Guide: Creating API Specs and Generating Go Beans

## 1. Write the OpenAPI Spec

- Use OpenAPI 3.0+ YAML format.
- Define:
    - `info`, `servers`, `tags`
    - `paths` for each endpoint, with HTTP methods, parameters, request bodies, and responses.
    - `components.schemas` for all request/response objects.
- Use `x-oapi-codegen-extra-tags` for Go struct tags (e.g., validation).
- The default behavior is to generate the required fields as non-pointer types and optional fields as pointer types.
  - Use `x-go-type-skip-optional-pointer: false` to ensure fields are generated as pointer type. 
- Use `x-go-json-ignore: true` to add `json:"-"` tag to a field, preventing it from being serialized in JSON.
- Use `$ref` to reuse schema definitions.

**Reference:**  
See `specs/bulkEdit/v1beta2/bulk_edit.yaml` for a complete example.
---

## 2. Create oapi-codegen Config

- Create a YAML config file (e.g., `oapi-models-config.yaml`) at the same level as your OpenAPI spec file.
- This file configures the code generation process.
- Set:
    - `package`: Go package name for generated code.
    - `generate.models`: `true` to generate Go structs.
    - `output`: Relative path for the generated file (e.g., `./bean/bean.go`).
    - `output-options`: Customize struct naming, skip pruning, etc.
    - `exclude-schemas`: Exclude error types if needed.

**Reference:**  
```yaml
# yaml-language-server: ...
package: api
generate:
  models: true
output-options:
  # to make sure that all types are generated
  skip-prune: true
  name-normalizer: ToCamelCaseWithDigits
  prefer-skip-optional-pointer: true
  prefer-skip-optional-pointer-on-container-types: true
  exclude-schemas:
    - "ErrorResponse"
    - "ApiError"
output: {{ // relative path to the generator.go file }}
```
---

## 3. Add go:generate Directive
- In your Go file (e.g., `generator.go`), add a `//go:generate` comment:
    ```go
        //go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen "--config=../../../../specs/bulkEdit/v1beta2/oapi-models-config.yaml" "../../../../specs/bulkEdit/v1beta2/bulk_edit.yaml"
    ```
- Here `--config` is the relative path (from the `generator.go` file) to your config file, i.e. `oapi-models-config.yaml` file.
- And the last argument is the path to your OpenAPI spec file, i.e. `bulk_edit.yaml`.
---

## 4. Generate the Beans
Run: (From the root dir)
```bash
    go generate ./pkg/...
```
> This will generate Go structs at the path specified in your config (e.g., bean/bean.go).
---

## 5. Best Practices
Keep OpenAPI specs DRY: use $ref and shared schemas.
Document every field and endpoint.
Use validation tags for all struct fields.
Exclude error response types from bean generation if not needed.
Keep config and spec files versioned and close to your Go code.
---

## Summary:
Write your OpenAPI YAML, create a codegen config, add a go:generate directive, and run go generate to produce Go beans at the desired path. 
Follow the structure in bulk_edit.yaml and oapi-models-config.yaml for consistency.
