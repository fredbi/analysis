// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package analysis

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

<<<<<<< HEAD
	"github.com/go-openapi/analysis/internal/flatten/normalize"
	"github.com/go-openapi/analysis/internal/flatten/operations"
	"github.com/go-openapi/analysis/internal/flatten/replace"
	"github.com/go-openapi/analysis/internal/flatten/schutils"
	"github.com/go-openapi/analysis/internal/flatten/sortref"
=======
	"strconv"

	"github.com/davecgh/go-spew/spew"
>>>>>>> a5ee2e3 (temp work)
	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/spec"
)

<<<<<<< HEAD
const definitionsPath = "#/definitions"
=======
// FlattenOpts configuration for flattening a swagger specification.
type FlattenOpts struct {
	// If Expand is true, we skip flattening the spec and expand it instead
	Expand   bool
	Spec     *Spec
	BasePath string
	// Enable propagation of schema extensions specified in the slice
	PropagateExt []string
>>>>>>> a5ee2e3 (temp work)

// newRef stores information about refs created during the flattening process
type newRef struct {
	key      string
	newName  string
	path     string
	isOAIGen bool
	resolved bool
	schema   *spec.Schema
	parents  []string
}

// context stores intermediary results from flatten
type context struct {
	newRefs  map[string]*newRef
	warnings []string
	resolved map[string]string
}

func newContext() *context {
	return &context{
		newRefs:  make(map[string]*newRef, 150),
		warnings: make([]string, 0),
		resolved: make(map[string]string, 50),
	}
}

// Flatten an analyzed spec and produce a self-contained spec bundle.
//
// There is a minimal and a full flattening mode.
//
// Minimally flattening a spec means:
//   - Expanding parameters, responses, path items, parameter items and header items (references to schemas are left
//     unscathed)
//   - Importing external (http, file) references so they become internal to the document
//   - Moving every JSON pointer to a $ref to a named definition (i.e. the reworked spec does not contain pointers
//     like "$ref": "#/definitions/myObject/allOfs/1")
//
// A minimally flattened spec thus guarantees the following properties:
//   - all $refs point to a local definition (i.e. '#/definitions/...')
//   - definitions are unique
//
// NOTE: arbitrary JSON pointers (other than $refs to top level definitions) are rewritten as definitions if they
// represent a complex schema or express commonality in the spec.
// Otherwise, they are simply expanded.
// Self-referencing JSON pointers cannot resolve to a type and trigger an error.
//
// Minimal flattening is necessary and sufficient for codegen rendering using go-swagger.
//
// Fully flattening a spec means:
//   - Moving every complex inline schema to be a definition with an auto-generated name in a depth-first fashion.
//
// By complex, we mean every JSON object with some properties.
// Arrays, when they do not define a tuple,
// or empty objects with or without additionalProperties, are not considered complex and remain inline.
//
// NOTE: rewritten schemas get a vendor extension x-go-gen-location so we know from which part of the spec definitions
// have been created.
//
// Available flattening options:
//   - Minimal: stops flattening after minimal $ref processing, leaving schema constructs untouched
//   - Expand: expand all $ref's in the document (inoperant if Minimal set to true)
//   - Verbose: croaks about name conflicts detected
//   - RemoveUnused: removes unused parameters, responses and definitions after expansion/flattening
//
// NOTE: expansion removes all $ref save circular $ref, which remain in place
//
// TODO: additional options
//   - ProgagateNameExtensions: ensure that created entries properly follow naming rules when their parent have set a
//     x-go-name extension
//   - LiftAllOfs:
//   - limit the flattening of allOf members when simple objects
//   - merge allOf with validation only
//   - merge allOf with extensions only
//   - ...
func Flatten(opts FlattenOpts) error {
	debugLog("FlattenOpts: %#v", opts)

	opts.flattenContext = newContext()

	// 1. Recursively expand responses, parameters, path items and items in simple schemas.
	//
	// This simplifies the spec and leaves only the $ref's in schema objects.
	if err := expand(&opts); err != nil {
		return err
	}

	// 2. Strip the current document from absolute $ref's that actually a in the root,
	// so we can recognize them as proper definitions
	//
	// In particular, this works around issue go-openapi/spec#76: leading absolute file in $ref is stripped
	if err := normalizeRef(&opts); err != nil {
		return err
	}

<<<<<<< HEAD
	// 3. Optionally remove shared parameters and responses already expanded (now unused).
	//
	// Operation parameters (i.e. under paths) remain.
	if opts.RemoveUnused {
		removeUnusedShared(&opts)
	}

	// 4. Import all remote references.
	if err := importReferences(&opts); err != nil {
=======
	// rewrite the inline schemas (schemas that aren't simple types or arrays of simple types)
	inlinedInternalKeys, err := nameInlinedSchemas(&opts)
	if err != nil {
>>>>>>> a5ee2e3 (temp work)
		return err
	}

<<<<<<< HEAD
	// 5. full flattening: rewrite inline schemas (schemas that aren't simple types or arrays or maps)
	if !opts.Minimal && !opts.Expand {
		if err := nameInlinedSchemas(&opts); err != nil {
=======
	// DEBUG
	opts.PropagateExt = []string{"x-go-name"}
	if len(opts.PropagateExt) > 0 && len(inlinedInternalKeys) > 0 {
		if err = propagatesExtensions(&opts, inlinedInternalKeys); err != nil {
>>>>>>> a5ee2e3 (temp work)
			return err
		}
	}

<<<<<<< HEAD
	// 6. Rewrite JSON pointers other than $ref to named definitions
	// and attempt to resolve conflicting names whenever possible.
	if err := stripPointersAndOAIGen(&opts); err != nil {
		return err
	}

	// 7. Strip the spec from unused definitions
	if opts.RemoveUnused {
		removeUnused(&opts)
	}

	// 8. Issue warning notifications, if any
	opts.croak()

	// TODO: simplify known schema patterns to flat objects with properties
	// examples:
	//  - lift simple allOf object,
	//  - empty allOf with validation only or extensions only
	//  - rework allOf arrays
	//  - rework allOf additionalProperties

	return nil
}

func expand(opts *FlattenOpts) error {
	if err := spec.ExpandSpec(opts.Swagger(), opts.ExpandOpts(!opts.Expand)); err != nil {
		return err
	}

	opts.Spec.reload() // re-analyze
=======
	// TODO: simplify known schema patterns to flat objects with properties?
	return nil
}

func propagatesExtensions(opts *FlattenOpts, inlinedInternalKeys []string) error {
	// propagatesExtensions looks in the referers of the inlinedInternalKeys list of definitions
	// and propagate the list of configured extensions from referers down to these keys.
	//
	// TL;DR: gives more control over the naming of xxxItems0, xxxAdditionalItemsItems0, etc...
	//
	// It propagates extensions created in a parent schema to the schemas self-generated by Flatten().
	// In practice, we want to propagate the x-go-name extension and be able to rename derived schemas types
	// which were not in the original spec (e.g. for definitions created out of tuples, additionalItems and additionalProperties).
	//
	sort.Strings(inlinedInternalKeys)
	topFirst := inlinedInternalKeys
	spew.Dump(topFirst)

	//
	// WRONG groupedRefs := reverseIndexForSchemaRefs(opts)
	//spew.Dump(groupedRefs)
	// scan top-down and propagates x-go-name extensions whenever needed
	b, _ := json.MarshalIndent(topFirst, "", "  ")
	log.Println(string(b))
	/*
		for i := len(depthFirst) - 1; i >= 0; i-- {
			key := depthFirst[i]
			sch := opts.Spec.allSchemas[key]
			if sch.Schema != nil && sch.Schema.Ref.String() != "" { // schemas previously inlined
				asch, err := Schema(SchemaOpts{Schema: sch.Schema, Root: opts.Swagger(), BasePath: opts.BasePath})
				if err != nil {
					return fmt.Errorf("schema analysis [%s]: %v", sch.Ref.String(), err)
				}
				log.Printf("Ref = %s", sch.Schema.Ref.String())
				if !asch.IsSimpleSchema && !asch.IsArray && (asch.IsTuple || asch.IsTupleWithExtra || asch.IsExtendedObject) {
					log.Printf("Key: %s, IsTuple: %t, IsTupleWithExtra: %t, IsExtendedObject: %t", key, asch.IsTuple, asch.IsTupleWithExtra, asch.IsExtendedObject)
					log.Printf("schema go name: %s", goName(sch.Schema, ""))
				}
			}
		}
	*/
>>>>>>> a5ee2e3 (temp work)

	return nil
}

<<<<<<< HEAD
// normalizeRef strips the current file from any absolute file $ref. This works around issue go-openapi/spec#76:
// leading absolute file in $ref is stripped
func normalizeRef(opts *FlattenOpts) error {
	debugLog("normalizeRef")

	altered := false
	for k, w := range opts.Spec.references.allRefs {
		if !strings.HasPrefix(w.String(), opts.BasePath+definitionsPath) { // may be a mix of / and \, depending on OS
			continue
		}

		altered = true
		debugLog("stripping absolute path for: %s", w.String())

		// strip the base path from definition
		if err := replace.UpdateRef(opts.Swagger(), k,
			spec.MustCreateRef(path.Join(definitionsPath, path.Base(w.String())))); err != nil {
			return err
		}
	}

	if altered {
		opts.Spec.reload() // re-analyze
	}

	return nil
}

func removeUnusedShared(opts *FlattenOpts) {
	opts.Swagger().Parameters = nil
	opts.Swagger().Responses = nil

	opts.Spec.reload() // re-analyze
}

func importReferences(opts *FlattenOpts) error {
	var (
		imported bool
		err      error
	)

	for !imported && err == nil {
		// iteratively import remote references until none left.
		// This inlining deals with name conflicts by introducing auto-generated names ("OAIGen")
		imported, err = importExternalReferences(opts)

		opts.Spec.reload() // re-analyze
	}

	return err
}

// nameInlinedSchemas replaces every complex inline construct by a named definition.
func nameInlinedSchemas(opts *FlattenOpts) error {
	debugLog("nameInlinedSchemas")

	namer := &InlineSchemaNamer{
		Spec:           opts.Swagger(),
		Operations:     operations.AllOpRefsByRef(opts.Spec, nil),
		flattenContext: opts.flattenContext,
		opts:           opts,
	}

	depthFirst := sortref.DepthFirst(opts.Spec.allSchemas)
	for _, key := range depthFirst {
		sch := opts.Spec.allSchemas[key]
		if sch.Schema == nil || sch.Schema.Ref.String() != "" || sch.TopLevel {
			continue
		}
=======
func goName(schema *swspec.Schema, origName string) string {
	if goName, _ := schema.Extensions.GetString("x-go-name"); goName != "" {
		return goName
	}
	return origName
}

func nameInlinedSchemas(opts *FlattenOpts) ([]string, error) {
	namer := &inlineSchemaNamer{Spec: opts.Swagger(), Operations: opRefsByRef(gatherOperations(opts.Spec, nil))}
	depthFirst := sortDepthFirst(opts.Spec.allSchemas)
	inlinedInternalKeys := make([]string, 0, len(depthFirst))

	for _, key := range depthFirst {
		sch := opts.Spec.allSchemas[key]
		if sch.Schema != nil && sch.Schema.Ref.String() == "" && !sch.TopLevel { // inline schema
			asch, err := Schema(SchemaOpts{Schema: sch.Schema, Root: opts.Swagger(), BasePath: opts.BasePath})
			if err != nil {
				return nil, fmt.Errorf("schema analysis [%s]: %v", sch.Ref.String(), err)
			}

			if !asch.IsSimpleSchema && !asch.IsArray { // complex schemas get moved
				var err error
				if inlinedInternalKeys, err = namer.Name(key, sch.Schema, asch, inlinedInternalKeys); err != nil {
					return nil, err
				}
			}
		}
	}
	return inlinedInternalKeys, nil
}
>>>>>>> a5ee2e3 (temp work)

		asch, err := Schema(SchemaOpts{Schema: sch.Schema, Root: opts.Swagger(), BasePath: opts.BasePath})
		if err != nil {
			return fmt.Errorf("schema analysis [%s]: %w", key, err)
		}

<<<<<<< HEAD
		if asch.isAnalyzedAsComplex() { // move complex schemas to definitions
			if err := namer.Name(key, sch.Schema, asch); err != nil {
				return err
			}
		}
	}

	opts.Spec.reload() // re-analyze

	return nil
=======
	for _, pk := range depthGroupOrder {
		res := grouped[pk]
		sort.Sort(res)
		for _, v := range res {
			sorted = append(sorted, v.Key)
		}
	}

	return
}

type key struct {
	Segments int
	Key      string
}
type keys []key

func (k keys) Len() int      { return len(k) }
func (k keys) Swap(i, j int) { k[i], k[j] = k[j], k[i] }
func (k keys) Less(i, j int) bool {
	return k[i].Segments > k[j].Segments || (k[i].Segments == k[j].Segments && k[i].Key < k[j].Key)
}

type inlineSchemaNamer struct {
	Spec       *swspec.Swagger
	Operations map[string]opRef
}

func opRefsByRef(oprefs map[string]opRef) map[string]opRef {
	result := make(map[string]opRef, len(oprefs))
	for _, v := range oprefs {
		result[v.Ref.String()] = v
	}
	return result
}

func (isn *inlineSchemaNamer) Name(key string, schema *swspec.Schema, aschema *AnalyzedSchema, inputInternalKeys []string) ([]string, error) {
	if swspec.Debug {
		log.Printf("naming inlined schema at %s", key)
	}

	parts := keyParts(key)
	for _, name := range namesFromKey(parts, aschema, isn.Operations) {
		if name != "" {
			// create unique name
			newName := uniqifyName(isn.Spec.Definitions, swag.ToJSONName(name))

			// clone schema
			sch, err := cloneSchema(schema)
			if err != nil {
				return nil, err
			}

			// replace values on schema
			if err := rewriteSchemaToRef(isn.Spec, key, swspec.MustCreateRef("#/definitions/"+newName)); err != nil {
				return nil, fmt.Errorf("name inlined schema: %v", err)
			}

			sch.AddExtension("x-go-gen-location", genLocation(parts))
			// fmt.Printf("{\n  %q,\n  \"\",\n  spec.MustCreateRef(%q),\n  \"\",\n},\n", key, "#/definitions/"+newName)

			// save cloned schema to definitions
			saveSchema(isn.Spec, newName, sch)

			if aschema.IsTuple || aschema.IsTupleWithExtra || aschema.IsExtendedObject {
				// save self-created schema definitions: this applies only to complex inline schemas from:
				// - tuples
				// - additionalItems in tuples
				// - additionalProperties schemas
				inputInternalKeys = append(inputInternalKeys, key)
			}
		}
	}
	return inputInternalKeys, nil
>>>>>>> a5ee2e3 (temp work)
}

func removeUnused(opts *FlattenOpts) {
	expected := make(map[string]struct{})
	for k := range opts.Swagger().Definitions {
		expected[path.Join(definitionsPath, jsonpointer.Escape(k))] = struct{}{}
	}

	for _, k := range opts.Spec.AllDefinitionReferences() {
		delete(expected, k)
	}

	for k := range expected {
		debugLog("removing unused definition %s", path.Base(k))
		if opts.Verbose {
			log.Printf("info: removing unused definition: %s", path.Base(k))
		}
		delete(opts.Swagger().Definitions, path.Base(k))
	}

	opts.Spec.reload() // re-analyze
}

func importKnownRef(entry sortref.RefRevIdx, refStr, newName string, opts *FlattenOpts) error {
	// rewrite ref with already resolved external ref (useful for cyclical refs):
	// rewrite external refs to local ones
	debugLog("resolving known ref [%s] to %s", refStr, newName)

	for _, key := range entry.Keys {
		if err := replace.UpdateRef(opts.Swagger(), key, spec.MustCreateRef(path.Join(definitionsPath, newName))); err != nil {
			return err
		}
<<<<<<< HEAD
=======
		// responses
		if parts.IsOperationResponse() {
			piref := parts.PathItemRef()
			if piref.String() != "" {
				if op, ok := operations[piref.String()]; ok {
					startIndex = 6
					baseNames = append(baseNames, []string{op.ID, parts.ResponseName(), "body"})
				}
			}
		}
	}

	// definitions
	if parts.IsDefinition() {
		nm := parts.DefinitionName()
		if nm != "" {
			startIndex = 2
			baseNames = append(baseNames, []string{parts.DefinitionName()})
		}
	}

	var result []string
	for _, segments := range baseNames {
		nm := parts.BuildName(segments, startIndex, aschema)
		if nm != "" {
			result = append(result, nm)
		}
	}
	sort.Strings(result)
	return result
}

const (
	pths        = "paths"
	responses   = "responses"
	parameters  = "parameters"
	definitions = "definitions"
)

var ignoredKeys map[string]struct{}

func init() {
	ignoredKeys = map[string]struct{}{
		"schema":     {},
		"properties": {},
		"not":        {},
		"anyOf":      {},
		"oneOf":      {},
	}
}

type splitKey []string

func (s splitKey) IsDefinition() bool {
	return len(s) > 1 && s[0] == definitions
}

func (s splitKey) DefinitionName() string {
	if !s.IsDefinition() {
		return ""
	}
	return s[1]
}

func (s splitKey) isKeyName(i int) bool {
	if i <= 0 {
		return false
	}
	count := 0
	for idx := i - 1; idx > 0; idx-- {
		if s[idx] != "properties" {
			break
		}
		count++
	}

	if count%2 != 0 {
		return true
	}
	return false
}

func (s splitKey) BuildName(segments []string, startIndex int, aschema *AnalyzedSchema) string {
	for i, part := range s[startIndex:] {
		if _, ignored := ignoredKeys[part]; !ignored || s.isKeyName(startIndex+i) {
			if part == "items" || part == "additionalItems" {
				if aschema.IsTuple || aschema.IsTupleWithExtra {
					segments = append(segments, "tuple")
				} else {
					segments = append(segments, "items")
				}
				if part == "additionalItems" {
					segments = append(segments, part)
				}
				continue
			}
			segments = append(segments, part)
		}
	}
	log.Printf("DEBUG FRED: BuildName = %s", strings.Join(segments, " "))
	return strings.Join(segments, " ")
}

func (s splitKey) IsOperation() bool {
	return len(s) > 1 && s[0] == pths
}

func (s splitKey) IsSharedOperationParam() bool {
	return len(s) > 2 && s[0] == pths && s[2] == parameters
}

func (s splitKey) IsOperationParam() bool {
	return len(s) > 3 && s[0] == pths && s[3] == parameters
}

func (s splitKey) IsOperationResponse() bool {
	return len(s) > 3 && s[0] == pths && s[3] == responses
}

func (s splitKey) IsDefaultResponse() bool {
	return len(s) > 4 && s[0] == pths && s[3] == responses && s[4] == "default"
}

func (s splitKey) IsStatusCodeResponse() bool {
	isInt := func() bool {
		_, err := strconv.Atoi(s[4])
		return err == nil
	}
	return len(s) > 4 && s[0] == pths && s[3] == responses && isInt()
}

func (s splitKey) ResponseName() string {
	if s.IsStatusCodeResponse() {
		code, _ := strconv.Atoi(s[4])
		return http.StatusText(code)
	}
	if s.IsDefaultResponse() {
		return "Default"
	}
	return ""
}

var validMethods map[string]struct{}

func init() {
	validMethods = map[string]struct{}{
		"GET":     {},
		"HEAD":    {},
		"OPTIONS": {},
		"PATCH":   {},
		"POST":    {},
		"PUT":     {},
		"DELETE":  {},
	}
}

func (s splitKey) PathItemRef() swspec.Ref {
	if len(s) < 3 {
		return swspec.Ref{}
	}
	pth, method := s[1], s[2]
	if _, validMethod := validMethods[strings.ToUpper(method)]; !validMethod && !strings.HasPrefix(method, "x-") {
		return swspec.Ref{}
	}
	return swspec.MustCreateRef("#" + path.Join("/", pths, jsonpointer.Escape(pth), strings.ToUpper(method)))
}

func (s splitKey) PathRef() swspec.Ref {
	if !s.IsOperation() {
		return swspec.Ref{}
	}
	return swspec.MustCreateRef("#" + path.Join("/", pths, jsonpointer.Escape(s[1])))
}

func keyParts(key string) splitKey {
	var res []string
	for _, part := range strings.Split(key[1:], "/") {
		if part != "" {
			res = append(res, jsonpointer.Unescape(part))
		}
	}
	return res
}

func rewriteSchemaToRef(spec *swspec.Swagger, key string, ref swspec.Ref) error {
	if swspec.Debug {
		log.Printf("rewriting schema to ref for %s with %s", key, ref.String())
	}
	pth := key[1:]
	ptr, err := jsonpointer.New(pth)
	if err != nil {
		return err
	}

	value, _, err := ptr.Get(spec)
	if err != nil {
		return err
	}

	switch refable := value.(type) {
	case *swspec.Schema:
		return rewriteParentRef(spec, key, ref)
	case *swspec.SchemaOrBool:
		if refable.Schema != nil {
			refable.Schema = &swspec.Schema{SchemaProps: swspec.SchemaProps{Ref: ref}}
		}
	case *swspec.SchemaOrArray:
		if refable.Schema != nil {
			refable.Schema = &swspec.Schema{SchemaProps: swspec.SchemaProps{Ref: ref}}
		}
	case swspec.Schema:
		return rewriteParentRef(spec, key, ref)
	default:
		return fmt.Errorf("no schema with ref found at %s for %T", key, value)
>>>>>>> a5ee2e3 (temp work)
	}

	return nil
}

func importNewRef(entry sortref.RefRevIdx, refStr string, opts *FlattenOpts) error {
	var (
		isOAIGen bool
		newName  string
	)

	debugLog("resolving schema from remote $ref [%s]", refStr)

	sch, err := spec.ResolveRefWithBase(opts.Swagger(), &entry.Ref, opts.ExpandOpts(false))
	if err != nil {
		return fmt.Errorf("could not resolve schema: %w", err)
	}

	// at this stage only $ref analysis matters
	partialAnalyzer := &Spec{
		references: referenceAnalysis{},
		patterns:   patternAnalysis{},
		enums:      enumAnalysis{},
	}
	partialAnalyzer.reset()
	partialAnalyzer.analyzeSchema("", sch, "/")

	// now rewrite those refs with rebase
	for key, ref := range partialAnalyzer.references.allRefs {
		if err := replace.UpdateRef(sch, key, spec.MustCreateRef(normalize.RebaseRef(entry.Ref.String(), ref.String()))); err != nil {
			return fmt.Errorf("failed to rewrite ref for key %q at %s: %w", key, entry.Ref.String(), err)
		}
	}

	// generate a unique name - isOAIGen means that a naming conflict was resolved by changing the name
	newName, isOAIGen = uniqifyName(opts.Swagger().Definitions, nameFromRef(entry.Ref))
	debugLog("new name for [%s]: %s - with name conflict:%t", strings.Join(entry.Keys, ", "), newName, isOAIGen)

	opts.flattenContext.resolved[refStr] = newName

	// rewrite the external refs to local ones
	for _, key := range entry.Keys {
		if err := replace.UpdateRef(opts.Swagger(), key,
			spec.MustCreateRef(path.Join(definitionsPath, newName))); err != nil {
			return err
		}

		// keep track of created refs
		resolved := false
		if _, ok := opts.flattenContext.newRefs[key]; ok {
			resolved = opts.flattenContext.newRefs[key].resolved
		}

		debugLog("keeping track of ref: %s (%s), resolved: %t", key, newName, resolved)
		opts.flattenContext.newRefs[key] = &newRef{
			key:      key,
			newName:  newName,
			path:     path.Join(definitionsPath, newName),
			isOAIGen: isOAIGen,
			resolved: resolved,
			schema:   sch,
		}
	}

	// add the resolved schema to the definitions
	schutils.Save(opts.Swagger(), newName, sch)

	return nil
}

// importExternalReferences iteratively digs remote references and imports them into the main schema.
//
// At every iteration, new remotes may be found when digging deeper: they are rebased to the current schema before being imported.
//
// This returns true when no more remote references can be found.
func importExternalReferences(opts *FlattenOpts) (bool, error) {
	debugLog("importExternalReferences")

	groupedRefs := sortref.ReverseIndex(opts.Spec.references.schemas, opts.BasePath)
	sortedRefStr := make([]string, 0, len(groupedRefs))
	if opts.flattenContext == nil {
		opts.flattenContext = newContext()
	}

	// sort $ref resolution to ensure deterministic name conflict resolution
	for refStr := range groupedRefs {
		sortedRefStr = append(sortedRefStr, refStr)
	}
	sort.Strings(sortedRefStr)

	complete := true

	for _, refStr := range sortedRefStr {
		entry := groupedRefs[refStr]
		if entry.Ref.HasFragmentOnly {
			continue
		}

		complete = false

		newName := opts.flattenContext.resolved[refStr]
		if newName != "" {
			if err := importKnownRef(entry, refStr, newName, opts); err != nil {
				return false, err
			}

			continue
		}

		// resolve schemas
		if err := importNewRef(entry, refStr, opts); err != nil {
			return false, err
		}
	}

	// maintains ref index entries
	for k := range opts.flattenContext.newRefs {
		r := opts.flattenContext.newRefs[k]

		// update tracking with resolved schemas
		if r.schema.Ref.String() != "" {
			ref := spec.MustCreateRef(r.path)
			sch, err := spec.ResolveRefWithBase(opts.Swagger(), &ref, opts.ExpandOpts(false))
			if err != nil {
				return false, fmt.Errorf("could not resolve schema: %w", err)
			}

			r.schema = sch
		}

		if r.path == k {
			continue
		}

		// update tracking with renamed keys: got a cascade of refs
		renamed := *r
		renamed.key = r.path
		opts.flattenContext.newRefs[renamed.path] = &renamed

		// indirect ref
		r.newName = path.Base(k)
		r.schema = spec.RefSchema(r.path)
		r.path = k
		r.isOAIGen = strings.Contains(k, "OAIGen")
	}

	return complete, nil
}

// stripPointersAndOAIGen removes anonymous JSON pointers from spec and chain with name conflicts handler.
// This loops until the spec has no such pointer and all name conflicts have been reduced as much as possible.
func stripPointersAndOAIGen(opts *FlattenOpts) error {
	// name all JSON pointers to anonymous documents
	if err := namePointers(opts); err != nil {
		return err
	}

	// remove unnecessary OAIGen ref (created when flattening external refs creates name conflicts)
	hasIntroducedPointerOrInline, ers := stripOAIGen(opts)
	if ers != nil {
		return ers
	}

	// iterate as pointer or OAIGen resolution may introduce inline schemas or pointers
	for hasIntroducedPointerOrInline {
		if !opts.Minimal {
			opts.Spec.reload() // re-analyze
			if err := nameInlinedSchemas(opts); err != nil {
				return err
			}
		}

		if err := namePointers(opts); err != nil {
			return err
		}

		// restrip and re-analyze
		var err error
		if hasIntroducedPointerOrInline, err = stripOAIGen(opts); err != nil {
			return err
		}
	}

	return nil
}

// stripOAIGen strips the spec from unnecessary OAIGen constructs, initially created to dedupe flattened definitions.
//
// A dedupe is deemed unnecessary whenever:
//   - the only conflict is with its (single) parent: OAIGen is merged into its parent (reinlining)
//   - there is a conflict with multiple parents: merge OAIGen in first parent, the rewrite other parents to point to
//     the first parent.
//
// This function returns true whenever it re-inlined a complex schema, so the caller may chose to iterate
// pointer and name resolution again.
func stripOAIGen(opts *FlattenOpts) (bool, error) {
	debugLog("stripOAIGen")
	replacedWithComplex := false

	// figure out referers of OAIGen definitions (doing it before the ref start mutating)
	for _, r := range opts.flattenContext.newRefs {
		updateRefParents(opts.Spec.references.allRefs, r)
	}

	for k := range opts.flattenContext.newRefs {
		r := opts.flattenContext.newRefs[k]
		debugLog("newRefs[%s]: isOAIGen: %t, resolved: %t, name: %s, path:%s, #parents: %d, parents: %v,  ref: %s",
			k, r.isOAIGen, r.resolved, r.newName, r.path, len(r.parents), r.parents, r.schema.Ref.String())

		if !r.isOAIGen || len(r.parents) == 0 {
			continue
		}

		hasReplacedWithComplex, err := stripOAIGenForRef(opts, k, r)
		if err != nil {
			return replacedWithComplex, err
		}

		replacedWithComplex = replacedWithComplex || hasReplacedWithComplex
	}

	debugLog("replacedWithComplex: %t", replacedWithComplex)
	opts.Spec.reload() // re-analyze

	return replacedWithComplex, nil
}

// updateRefParents updates all parents of an updated $ref
func updateRefParents(allRefs map[string]spec.Ref, r *newRef) {
	if !r.isOAIGen || r.resolved { // bail on already resolved entries (avoid looping)
		return
	}
	for k, v := range allRefs {
		if r.path != v.String() {
			continue
		}

		found := false
		for _, p := range r.parents {
			if p == k {
				found = true

				break
			}
		}
		if !found {
			r.parents = append(r.parents, k)
		}
	}
}

func stripOAIGenForRef(opts *FlattenOpts, k string, r *newRef) (bool, error) {
	replacedWithComplex := false

	pr := sortref.TopmostFirst(r.parents)

	// rewrite first parent schema in hierarchical then lexicographical order
	debugLog("rewrite first parent %s with schema", pr[0])
	if err := replace.UpdateRefWithSchema(opts.Swagger(), pr[0], r.schema); err != nil {
		return false, err
	}

	if pa, ok := opts.flattenContext.newRefs[pr[0]]; ok && pa.isOAIGen {
		// update parent in ref index entry
		debugLog("update parent entry: %s", pr[0])
		pa.schema = r.schema
		pa.resolved = false
		replacedWithComplex = true
	}

	// rewrite other parents to point to first parent
	if len(pr) > 1 {
		for _, p := range pr[1:] {
			replacingRef := spec.MustCreateRef(pr[0])

			// set complex when replacing ref is an anonymous jsonpointer: further processing may be required
			replacedWithComplex = replacedWithComplex || path.Dir(replacingRef.String()) != definitionsPath
			debugLog("rewrite parent with ref: %s", replacingRef.String())

			// NOTE: it is possible at this stage to introduce json pointers (to non-definitions places).
			// Those are stripped later on.
			if err := replace.UpdateRef(opts.Swagger(), p, replacingRef); err != nil {
				return false, err
			}

			if pa, ok := opts.flattenContext.newRefs[p]; ok && pa.isOAIGen {
				// update parent in ref index
				debugLog("update parent entry: %s", p)
				pa.schema = r.schema
				pa.resolved = false
				replacedWithComplex = true
			}
		}
	}

	// remove OAIGen definition
	debugLog("removing definition %s", path.Base(r.path))
	delete(opts.Swagger().Definitions, path.Base(r.path))

	// propagate changes in ref index for keys which have this one as a parent
	for kk, value := range opts.flattenContext.newRefs {
		if kk == k || !value.isOAIGen || value.resolved {
			continue
		}

		found := false
		newParents := make([]string, 0, len(value.parents))
		for _, parent := range value.parents {
			switch {
			case parent == r.path:
				found = true
				parent = pr[0]
			case strings.HasPrefix(parent, r.path+"/"):
				found = true
				parent = path.Join(pr[0], strings.TrimPrefix(parent, r.path))
			}

			newParents = append(newParents, parent)
		}

		if found {
			value.parents = newParents
		}
	}

	// mark naming conflict as resolved
	debugLog("marking naming conflict resolved for key: %s", r.key)
	opts.flattenContext.newRefs[r.key].isOAIGen = false
	opts.flattenContext.newRefs[r.key].resolved = true

	// determine if the previous substitution did inline a complex schema
	if r.schema != nil && r.schema.Ref.String() == "" { // inline schema
		asch, err := Schema(SchemaOpts{Schema: r.schema, Root: opts.Swagger(), BasePath: opts.BasePath})
		if err != nil {
			return false, err
		}

		debugLog("re-inlined schema: parent: %s, %t", pr[0], asch.isAnalyzedAsComplex())
		replacedWithComplex = replacedWithComplex || !(path.Dir(pr[0]) == definitionsPath) && asch.isAnalyzedAsComplex()
	}

	return replacedWithComplex, nil
}

// namePointers replaces all JSON pointers to anonymous documents by a $ref to a new named definitions.
//
// This is carried on depth-first. Pointers to $refs which are top level definitions are replaced by the $ref itself.
// Pointers to simple types are expanded, unless they express commonality (i.e. several such $ref are used).
func namePointers(opts *FlattenOpts) error {
	debugLog("name pointers")

	refsToReplace := make(map[string]SchemaRef, len(opts.Spec.references.schemas))
	for k, ref := range opts.Spec.references.allRefs {
		if path.Dir(ref.String()) == definitionsPath {
			// this a ref to a top-level definition: ok
			continue
		}

		result, err := replace.DeepestRef(opts.Swagger(), opts.ExpandOpts(false), ref)
		if err != nil {
			return fmt.Errorf("at %s, %w", k, err)
		}

		replacingRef := result.Ref
		sch := result.Schema
		if opts.flattenContext != nil {
			opts.flattenContext.warnings = append(opts.flattenContext.warnings, result.Warnings...)
		}

		debugLog("planning pointer to replace at %s: %s, resolved to: %s", k, ref.String(), replacingRef.String())
		refsToReplace[k] = SchemaRef{
			Name:     k,            // caller
			Ref:      replacingRef, // called
			Schema:   sch,
			TopLevel: path.Dir(replacingRef.String()) == definitionsPath,
		}
	}

	depthFirst := sortref.DepthFirst(refsToReplace)
	namer := &InlineSchemaNamer{
		Spec:           opts.Swagger(),
		Operations:     operations.AllOpRefsByRef(opts.Spec, nil),
		flattenContext: opts.flattenContext,
		opts:           opts,
	}

	for _, key := range depthFirst {
		v := refsToReplace[key]
		// update current replacement, which may have been updated by previous changes of deeper elements
		result, erd := replace.DeepestRef(opts.Swagger(), opts.ExpandOpts(false), v.Ref)
		if erd != nil {
			return fmt.Errorf("at %s, %w", key, erd)
		}

		if opts.flattenContext != nil {
			opts.flattenContext.warnings = append(opts.flattenContext.warnings, result.Warnings...)
		}

		v.Ref = result.Ref
		v.Schema = result.Schema
		v.TopLevel = path.Dir(result.Ref.String()) == definitionsPath
		debugLog("replacing pointer at %s: resolved to: %s", key, v.Ref.String())

		if v.TopLevel {
			debugLog("replace pointer %s by canonical definition: %s", key, v.Ref.String())

			// if the schema is a $ref to a top level definition, just rewrite the pointer to this $ref
			if err := replace.UpdateRef(opts.Swagger(), key, v.Ref); err != nil {
				return err
			}

			continue
		}

		if err := flattenAnonPointer(key, v, refsToReplace, namer, opts); err != nil {
			return err
		}
	}

	opts.Spec.reload() // re-analyze

	return nil
}

func flattenAnonPointer(key string, v SchemaRef, refsToReplace map[string]SchemaRef, namer *InlineSchemaNamer, opts *FlattenOpts) error {
	// this is a JSON pointer to an anonymous document (internal or external):
	// create a definition for this schema when:
	// - it is a complex schema
	// - or it is pointed by more than one $ref (i.e. expresses commonality)
	// otherwise, expand the pointer (single reference to a simple type)
	//
	// The named definition for this follows the target's key, not the caller's
	debugLog("namePointers at %s for %s", key, v.Ref.String())

	// qualify the expanded schema
	asch, ers := Schema(SchemaOpts{Schema: v.Schema, Root: opts.Swagger(), BasePath: opts.BasePath})
	if ers != nil {
		return fmt.Errorf("schema analysis [%s]: %w", key, ers)
	}
	callers := make([]string, 0, 64)

	debugLog("looking for callers")

	an := New(opts.Swagger())
	for k, w := range an.references.allRefs {
		r, err := replace.DeepestRef(opts.Swagger(), opts.ExpandOpts(false), w)
		if err != nil {
			return fmt.Errorf("at %s, %w", key, err)
		}

		if opts.flattenContext != nil {
			opts.flattenContext.warnings = append(opts.flattenContext.warnings, r.Warnings...)
		}

		if r.Ref.String() == v.Ref.String() {
			callers = append(callers, k)
		}
	}

	debugLog("callers for %s: %d", v.Ref.String(), len(callers))
	if len(callers) == 0 {
		// has already been updated and resolved
		return nil
	}

	parts := sortref.KeyParts(v.Ref.String())
	debugLog("number of callers for %s: %d", v.Ref.String(), len(callers))

	// identifying edge case when the namer did nothing because we point to a non-schema object
	// no definition is created and we expand the $ref for all callers
	if (!asch.IsSimpleSchema || len(callers) > 1) && !parts.IsSharedParam() && !parts.IsSharedResponse() {
		debugLog("replace JSON pointer at [%s] by definition: %s", key, v.Ref.String())
		if err := namer.Name(v.Ref.String(), v.Schema, asch); err != nil {
			return err
		}

		// regular case: we named the $ref as a definition, and we move all callers to this new $ref
		for _, caller := range callers {
			if caller == key {
				continue
			}

			// move $ref for next to resolve
			debugLog("identified caller of %s at [%s]", v.Ref.String(), caller)
			c := refsToReplace[caller]
			c.Ref = v.Ref
			refsToReplace[caller] = c
		}

		return nil
	}

	debugLog("expand JSON pointer for key=%s", key)

	if err := replace.UpdateRefWithSchema(opts.Swagger(), key, v.Schema); err != nil {
		return err
	}
	// NOTE: there is no other caller to update

	return nil
}
