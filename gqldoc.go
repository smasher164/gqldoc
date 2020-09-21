package gqldoc

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"unsafe"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
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

type panicWriter struct{ io.Writer }

func (w panicWriter) Write(p []byte) (n int, err error) {
	if n, err = w.Writer.Write(p); err == nil {
		return
	}
	panic(err)
}

// Formatter formats and writes the schema to dst, returning any
// error encountered.
type Formatter func(dst io.Writer, schema *ast.Schema) error

// FormatGraphQL reformats the GraphQL schema and writes it to dst,
// returning any error encountered when calling Write.
func FormatGraphQL(dst io.Writer, schema *ast.Schema) (err error) {
	defer func() {
		if e, ok := recover().(error); ok {
			err = e
		}
	}()
	f := formatter.NewFormatter(panicWriter{dst})
	f.FormatSchema(schema)
	return nil
}
