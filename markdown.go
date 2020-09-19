package gqldoc

// TODO table of contents

import (
	"io"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/mitchellh/go-wordwrap"
	"github.com/tdewolff/minify/v2"
	mhtml "github.com/tdewolff/minify/v2/html"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func indent(tabs int, input string) string {
	prefix := strings.Repeat("\t", tabs)
	split := strings.Split(input, "\n")
	for i := range split {
		if i != 0 {
			split[i] = prefix + split[i]
		}
	}
	return strings.Join(split, "\n")
}

const mdScalar = `### {{.Name}}
{{.Description | desc}}
{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{- end}}
`

const mdEnum = `### {{.Name}}
{{.Description | desc}}
{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{- end}}

#### Values
{{- range .EnumValues}}
**{{.Name}}**

{{.Description | desc}}
{{end -}}
`

const mdUnion = `### {{.Name}}
{{.Description | desc}}
{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{- end}}

#### Possible types
{{- range .Types}}
- [{{.}}]({{anchor . "type"}})
{{- end}}
`

const mdTableInput = `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>`

const mdInput = `### {{.Name}}
{{.Description | desc}}
{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{- end}}

#### Input fields
{{indentTemplate "tableInput" . 0 | minify}}
`

const mdArguments = `<table>
	<thead><tr><th>Arguments</th></tr></thead>
	<tbody>
	{{- range .}}
		<tr>
			<td>
				<strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)
				<br>
				{{- wrap .Description 69 | desc}}
			</td>
		</tr>
	{{- end}}
	</tbody>
</table>
`

const mdDirectives = `{{range .Directives -}}
<table>
	<thead>
		<tr><th>{{.Name}}</th></tr>
	</thead>
	<tbody>
		{{- range .Arguments}}
		<tr>
			<td><strong>{{.Name}}</strong>: {{len .Name | add 1 | sub 69 | wrap .Value.String | desc}}</td>
		</tr>
		{{- end}}
	</tbody>
</table>
{{- end -}}
`

const mdTableInterface = `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" .Arguments 3 | minify}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>`

const mdInterface = `### {{.Name}}
{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if implementers .}}
#### Implemented by
{{- range implementers .}}
- [<code>{{.Name}}</code>]({{anchor .Name "type"}})
{{- end}}
{{- end}}

#### Fields
{{indentTemplate "tableInterface" . 0 | minify}}
`

const mdTableObject = `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" .Arguments 3 | minify}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>`

const mdObject = `### {{.Name}}
{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if .Interfaces}}
#### Implements
{{- range .Interfaces}}
- [<code>{{.}}</code>]({{anchor . "type"}})
{{- end}}
{{- end}}
{{- if .Fields}}
#### Fields
{{indentTemplate "tableObject" . 0 | minify}}
{{- end}}
`

const mdTableQueries = `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Arguments}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>`

const mdQueries = `{{range .Fields -}}
### {{.Name}}
**Type:** [{{.Type}}]({{anchor .Type.Name "type"}})

{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if .Arguments}}
#### Arguments
{{indentTemplate "tableQueries" . 0 | minify}}
{{- end}}

---
{{end -}}
`

const mdTableMutations = `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range fields .Type}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name "type"}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>`

const mdMutations = `{{range .Fields -}}
### {{.Name}}
{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if .Arguments}}
#### Input fields
{{- range .Arguments}}
- <code>{{.Name}}</code>([<code>{{.Type}}</code>]({{anchor .Type.Name "type"}}))
{{- end}}
{{- end}}

{{- if fields .Type}}
#### Return fields
{{indentTemplate "tableMutations" . 0 | minify}}
{{- end}}

---
{{end -}}
`

const mdSchema = `# Reference
{{if .Query -}}
## Queries
{{template "queries" .Query}}{{end -}}
{{if .Mutation -}}
## Mutations
{{template "mutations" .Mutation}}{{end -}}
{{if .Subscription -}}
## Subscriptions
{{template "queries" .Subscription}}{{end -}}
{{if .Objects}}
## Objects
{{range .Objects}}{{template "object" .}}
---
{{end -}}
{{end -}}
{{if .Interfaces}}
## Interfaces
{{range .Interfaces}}{{template "interface" .}}
---
{{end -}}
{{end -}}
{{if .Enums}}
## Enums
{{range .Enums}}{{template "enum" .}}
---
{{end -}}
{{end -}}
{{if .Unions}}
## Unions
{{range .Unions}}{{template "union" .}}
---
{{end -}}
{{end -}}
{{if .Inputs}}
## Input objects
{{range .Inputs}}{{template "input" .}}
---
{{end -}}
{{end -}}
{{if .Scalars}}
## Scalars
{{range .Scalars}}{{template "scalar" .}}
---
{{end -}}
{{end -}}
`

type md struct {
	Query        *ast.Definition
	Mutation     *ast.Definition
	Subscription *ast.Definition
	Objects      []*ast.Definition
	Interfaces   []*ast.Definition
	Enums        []*ast.Definition
	Unions       []*ast.Definition
	Inputs       []*ast.Definition
	Scalars      []*ast.Definition

	count  map[string]int
	anchor map[string]map[string]string
}

func valid(f interface{}) bool {
	var name string
	switch f := f.(type) {
	case *ast.FieldDefinition:
		name = f.Name
	case *ast.Definition:
		name = f.Name
	}
	switch name {
	case "", "Query", "Mutation", "Subscription":
		return false
	}
	return !strings.HasPrefix(name, "_")
}

func (md *md) filterFields(def *ast.Definition) *ast.Definition {
	if def == nil {
		return def
	}
	res := make(ast.FieldList, 0, len(def.Fields))
	for _, field := range def.Fields {
		if valid(field) {
			res = append(res, field)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	for _, field := range res {
		md.updateAnchor(field.Name, "field")
	}
	def.Fields = res
	return def
}

func (md *md) filterKind(fields map[string]*ast.Definition, kind ast.DefinitionKind) []*ast.Definition {
	res := make([]*ast.Definition, 0, len(fields))
	for _, field := range fields {
		if field.Kind == kind && valid(field) {
			res = append(res, field)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	for _, field := range res {
		md.updateAnchor(field.Name, "type")
	}
	return res
}

func (md *md) updateAnchor(key, typ string, refs ...string) {
	k := strings.ToLower(key)
	c := md.count[k]
	var ref string
	if len(refs) == 1 {
		ref = refs[0]
	} else {
		ref = "#" + k
	}
	if c > 0 {
		ref += "-" + strconv.Itoa(c)
	}
	if _, ok := md.anchor[typ]; !ok {
		md.anchor[typ] = make(map[string]string)
	}
	md.anchor[typ][key] = ref
	md.count[k]++
}

func FormatMarkdown(dst io.Writer, schema *ast.Schema) error {
	md := &md{
		count:  make(map[string]int),
		anchor: make(map[string]map[string]string),
	}
	md.updateAnchor("Query", "type", "#queries")
	md.Query = md.filterFields(schema.Query)
	md.updateAnchor("Mutation", "type", "#mutations")
	md.Mutation = md.filterFields(schema.Mutation)
	md.updateAnchor("Subscription", "type", "#subscriptions")
	md.Subscription = md.filterFields(schema.Subscription)
	md.Objects = md.filterKind(schema.Types, ast.Object)
	md.Interfaces = md.filterKind(schema.Types, ast.Interface)
	md.Enums = md.filterKind(schema.Types, ast.Enum)
	md.Unions = md.filterKind(schema.Types, ast.Union)
	md.Inputs = md.filterKind(schema.Types, ast.InputObject)
	md.Scalars = md.filterKind(schema.Types, ast.Scalar)
	gm := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	t := template.New("schema")
	t = t.Funcs(template.FuncMap{
		"indentTemplate": func(name string, v interface{}, tabs int) (string, error) {
			var buf strings.Builder
			if err := t.ExecuteTemplate(&buf, name, v); err != nil {
				return "", err
			}
			return indent(tabs, buf.String()), nil
		},
		"desc": func(input string) (string, error) {
			var buf strings.Builder
			err := gm.Convert([]byte(input), &buf)
			s := strings.TrimSuffix(buf.String(), "\n")
			s = strings.TrimPrefix(s, "<p>")
			s = strings.TrimSuffix(s, "</p>")
			return s, err
		},
		"wrap": func(s string, i int) string {
			return wordwrap.WrapString(s, uint(i))
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"anchor": func(s, T string) string {
			return md.anchor[T][s]
		},
		"implementers": func(def *ast.Definition) []*ast.Definition { return schema.GetPossibleTypes(def) },
		"fields": func(T *ast.Type) ast.FieldList {
			return schema.Types[T.Name()].Fields
		},
		"minify": func(s string) (string, error) {
			m := minify.New()
			m.AddFunc("text/html", mhtml.Minify)
			return m.String("text/html", s)
		},
	})
	template.Must(t.New("arguments").Parse(mdArguments))
	template.Must(t.New("directives").Parse(mdDirectives))
	template.Must(t.New("scalar").Parse(mdScalar))
	template.Must(t.New("tableObject").Parse(mdTableObject))
	template.Must(t.New("tableQueries").Parse(mdTableQueries))
	template.Must(t.New("tableMutations").Parse(mdTableMutations))
	template.Must(t.New("tableInput").Parse(mdTableInput))
	template.Must(t.New("tableInterface").Parse(mdTableInterface))
	template.Must(t.New("object").Parse(mdObject))
	template.Must(t.New("interface").Parse(mdInterface))
	template.Must(t.New("union").Parse(mdUnion))
	template.Must(t.New("enum").Parse(mdEnum))
	template.Must(t.New("input").Parse(mdInput))
	template.Must(t.New("queries").Parse(mdQueries))
	template.Must(t.New("mutations").Parse(mdMutations))
	template.Must(t.Parse(mdSchema))
	return t.Execute(dst, md)
}
