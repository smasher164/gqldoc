package gqldoc

// TODO subscriptions
// TODO table of contents

import (
	"io"
	"strings"
	"text/template"

	"github.com/mitchellh/go-wordwrap"
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

const mdScalar = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}
`

const mdEnum = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}

#### Values
{{- range .EnumValues}}
**{{.Name}}**

{{.Description | desc}}
{{end -}}
`

const mdUnion = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}

#### Possible types
{{- range .Types}}
- [{{.}}]({{anchor .}})
{{- end}}
`

const mdInput = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}

#### Input fields
<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
`

const mdArguments = `<table>
	<thead><tr><th>Arguments</th></tr></thead>
	<tbody>
	{{- range .}}
		<tr>
			<td>
				<strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)
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

const mdInterface = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if implementers .}}
#### Implemented by
{{- range implementers .}}
- [<code>{{.Name}}</code>]({{anchor .Name}})
{{- end}}
{{- end}}

#### Fields
<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" .Arguments 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
`

const mdObject = `### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if .Interfaces}}
#### Implements
{{- range .Interfaces}}
- [<code>{{.}}</code>]({{anchor .}})
{{- end}}
{{- end}}
{{- if .Fields}}
#### Fields
<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" .Arguments 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
{{- end}}
`

const mdQuery = `## Queries
{{range .Fields}}
### [{{.Name}}]({{anchor .Name}})
**Type:** [{{.Type}}]({{anchor .Type.Name}})

{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if .Arguments}}
#### Arguments
<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range .Arguments}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
{{- end}}

---

{{end}}
`

const mdMutation = `## Mutations
{{range .Fields}}
### [{{.Name}}]({{anchor .Name}})
{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if .Arguments}}
#### Input fields
{{- range .Arguments}}
- <code>{{.Name}}</code>([<code>{{.Type}}</code>]({{anchor .Type.Name}}))
{{- end}}
{{- end}}

{{if fields .Type}}
#### Return fields
<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
	{{- range fields .Type}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{anchor .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
{{end}}

---

{{end}}
`

func FormatMarkdown(dst io.Writer, schema *ast.Schema) error {
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
		"anchor": func(s string) (string, error) {
			// TODO validation
			return "#" + strings.ToLower(s), nil
		},
		"implementers": func(def *ast.Definition) []*ast.Definition { return schema.GetPossibleTypes(def) },
		"fields": func(T *ast.Type) ast.FieldList {
			return schema.Types[T.Name()].Fields
		},
	})
	t.New("arguments").Parse(mdArguments)
	t.New("directives").Parse(mdDirectives)
	t.New("enum").Parse(mdEnum)
	t.New("scalar").Parse(mdScalar)
	t.New("union").Parse(mdUnion)
	t.New("input").Parse(mdInput)
	t.New("object").Parse(mdObject)
	t.New("interface").Parse(mdInterface)
	t.New("query").Parse(mdQuery)
	t.New("mutation").Parse(mdMutation)
	return t.Lookup("mutation").Execute(dst, schema.Mutation)
}
