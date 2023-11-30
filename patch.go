package analysis

import (
	"regexp"
	"sync"
	"text/template"

	"github.com/go-openapi/spec"
)

type Matcher struct {
	Regexp  string `json:"regexp"`
	matcher *matcher
}

func NewMatcher() *Matcher {
	return &Matcher{matcher: &matcher{compileOnce: &sync.Once{}}}
}

func (m Matcher) Matched(pointer string) bool {
	return m.matcher.Match(m.Regexp, pointer)
}

type matcher struct {
	compileOnce *sync.Once
	rex         *regexp.Regexp
}

func (m matcher) Match(rex string, pointer string) bool {
	m.compileOnce.Do(func() {
		m.rex = regexp.MustCompile(rex)
	})

	return m.rex.MatchString(pointer)
}

type PatchAction string

const (
	PatchUpdate PatchAction = "update"
	PatchAdd    PatchAction = "add"
	PathRemove  PatchAction = "remove"
)

type ExprOrTemplate struct {
	Expr     string
	Template *template.Template
}

type PatchRule struct {
	Match       Matcher        `json:"match"`  // a regexp on the pointers scanner under this entry
	Action      PatchAction    `json:"action"` // action to perform on match
	ContentExpr ExprOrTemplate `json:"content"`
}

type PatchExtension struct {
	Rules []PatchRule `json:"rules"`
}

// Patch a spec with the x-go-patch extension.
func Patch(patchSpec *spec.Swagger, patched ...*spec.Swagger) error {
	// sp := New(patchSpec)

	return nil
}
