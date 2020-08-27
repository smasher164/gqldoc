package gqldoc

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"unsafe"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// ParseFiles parses GraphQL schema definitions from the named files.
// The returned AST holds the combined schema of all the files. There
// must be at least one file, and the file's extension must be either
// .gql or .graphql. If an error occurs, parsing stops and the
// returned AST is nil.
func ParseFiles(filenames []string) (*ast.Schema, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("unable to parse schema: no files provided")
	}
	var src []*ast.Source
	for _, file := range filenames {
		if ext := filepath.Ext(file); ext != ".gql" && ext != ".graphql" {
			return nil, fmt.Errorf("unable to parse %s: must have extension .gql or .graphql", file)
		}
		b, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s: %w", file, err)
		}
		input := *(*string)(unsafe.Pointer(&b))
		src = append(src, &ast.Source{Name: file, Input: input})
	}
	schema, err := gqlparser.LoadSchema(src...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse schema: %w", err)
	}
	return schema, nil
}
