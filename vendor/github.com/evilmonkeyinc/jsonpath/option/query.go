package option

// QueryOptions represents optional functionality for the query functions that can be enabled or disabled.
//
// The default will be for all optional functionality to be disabled.
type QueryOptions struct {
	// AllowMapReferenceByIndex allow maps to be referenced by index in any token.
	AllowMapReferenceByIndex bool
	// AllowStringReferenceByIndex allow string characters to be referenced by index in any token.
	AllowStringReferenceByIndex bool

	// AllowMapReferenceByIndex allow maps to be referenced by index in range tokens.
	AllowMapReferenceByIndexInRange bool
	// AllowStringReferenceByIndexInRange allow string characters to be referenced by index in range tokens.
	AllowStringReferenceByIndexInRange bool

	// AllowMapReferenceByIndexInUnion allow maps to be referenced by index in union tokens.
	AllowMapReferenceByIndexInUnion bool
	// AllowStringReferenceByIndexInUnion allow string characters to be referenced by index in union tokens.
	AllowStringReferenceByIndexInUnion bool

	// AllowMapReferenceByIndexInSubscript allow maps to be referenced by index in subscript tokens.
	AllowMapReferenceByIndexInSubscript bool
	// AllowStringReferenceByIndexInSubscript allow string characters to be referenced by index in subscript tokens.
	AllowStringReferenceByIndexInSubscript bool

	// FailUnionOnInvalidIdentifier force union tokens to fail on missing or invalid keys or invalid index.
	FailUnionOnInvalidIdentifier bool
}
