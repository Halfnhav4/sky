package codegen

import (
	"bytes"
	"testing"

	"github.com/skydb/sky/query/ast"
	"github.com/stretchr/testify/assert"
)

// Ensure that variable definitions can be generated.
func TestVarDecl(t *testing.T) {
	assert.Equal(t, `sky_string_t _foo;`, vardecl(ast.NewVariable("foo", "string")))
	assert.Equal(t, `int32_t _foo;`, vardecl(ast.NewVariable("foo", "factor")))
	assert.Equal(t, `int32_t _foo;`, vardecl(ast.NewVariable("foo", "integer")))
	assert.Equal(t, `double _foo;`, vardecl(ast.NewVariable("foo", "float")))
	assert.Equal(t, `bool _foo;`, vardecl(ast.NewVariable("foo", "boolean")))
}

// Ensure that system variable declarations are not generated.
func TestVarDeclSystem(t *testing.T) {
	assert.Equal(t, "", vardecl(ast.NewVariable("@eof", "string")))
}

// Ensure that the timestamp variable declaration is not generated.
func TestVarDeclTimestamp(t *testing.T) {
	assert.Equal(t, "", vardecl(ast.NewVariable("timestamp", "integer")))
}

// Ensure that metamethod definitions can be generated.
func TestMetaDecl(t *testing.T) {
	assert.Equal(t, `foo = function(event) return ffi.string(event._foo.data, event._foo.length) end,`, metadecl(ast.NewVariable("foo", "string")))
	assert.Equal(t, `foo = function(event) return event._foo end,`, metadecl(ast.NewVariable("foo", "factor")))
}

// Ensure that system variables do not generate metamethods.
func TestMetaDeclSystem(t *testing.T) {
	assert.Equal(t, "", metadecl(ast.NewVariable("@eof", "string")))
}

// MustExecuteTemplate executes a named template and returns the result.
// Panic occurs on error.
func MustExecuteTemplate(name string, data interface{}) string {
	var b bytes.Buffer
	if err := tmpl.ExecuteTemplate(&b, name, data); err != nil {
		panic("Template '" + name + "' did not execute: " + err.Error())
	}
	return b.String()
}
