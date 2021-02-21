package replace

import "github.com/go-openapi/spec"

// RewriteSchemaToRef replaces a schema with a Ref
func RewriteSchemaToRef(sp *spec.Swagger, key string, ref spec.Ref) error {
	return rewriteSchemaToRef(sp, key, ref)
}

// UpdateRef replaces a ref by another one
func UpdateRef(sp interface{}, key string, ref spec.Ref) error {
	return updateRef(sp, key, ref)
}

// UpdateRefWithSchema replaces a ref with a schema (i.e. re-inline schema)
func UpdateRefWithSchema(sp *spec.Swagger, key string, sch *spec.Schema) error {
	return updateRefWithSchema(sp, key, sch)
}

// DeepestRefResult holds the results from DeepestRef analysis
type DeepestRefResult struct {
	Ref      spec.Ref
	Schema   *spec.Schema
	Warnings []string
}

// DeepestRef finds the first definition ref, from a cascade of nested refs which are not definitions.
//  - if no definition is found, returns the deepest ref.
//  - pointers to external files are expanded
//
// NOTE: all external $ref's are assumed to be already expanded at this stage.
func DeepestRef(sp *spec.Swagger, opts *spec.ExpandOptions, ref spec.Ref) (*DeepestRefResult, error) {
	return deepestRef(sp, opts, ref)
}
