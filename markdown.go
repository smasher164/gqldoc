package gqldoc

// TODO directives
// TODO urls
// TODO queries
// TODO mutations
// TODO subscriptions
// TODO Table of contents
// TODO max width

import (
	"io"
	"strings"
	"text/template"

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

const mdScalar = `### [{{.Name}}]()
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}
`

const mdEnum = `### [{{.Name}}]()
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

const mdUnion = `### [{{.Name}}]()
{{.Description | desc}}
{{- if .Directives}}
{{template "directives" .}}
{{- end}}

#### Possible types
{{- range .Types}}
- [{{.}}]()
{{- end}}
`

const mdInput = `### [{{.Name}}]()
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
			<td><strong>{{.Name}}</strong> (<a href=""><strong>{{.Type}}</strong></a>)</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3}}
			{{- end}}</td>
		</tr>
	{{- end}}
	</tbody>
</table>
`

const mdArguments = `{{define "arguments" -}}
<table>
	<thead><tr><th>Arguments</th></tr></thead>
	<tbody>
	{{- range .}}
		<tr>
			<td>
				<strong>{{.Name}}</strong> (<a href=""><strong>{{.Type}}</strong></a>)
				<br>
				{{- .Description | desc}}
			</td>
		</tr>
	{{- end}}
	</tbody>
</table>
{{- end}}
`

const mdInterface = `### [{{.Name}}]()
{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if .Types}}
#### Implemented by
{{- range .Types}}
- [<code>{{.}}</code>]()
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
			<td><strong>{{.Name}}</strong> (<a href=""><strong>{{.Type}}</strong></a>)</td>
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

const mdObject = `### [{{.Name}}]()
{{.Description | desc}}

{{- if .Directives}}
{{template "directives" .}}
{{end}}

{{- if .Interfaces}}
#### Implements
{{- range .Interfaces}}
- [<code>{{.}}</code>]()
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
			<td><strong>{{.Name}}</strong> (<a href=""><strong>{{.Type}}</strong></a>)</td>
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

const mdDirectives = `{{define "directives" -}}
{{range .Directives -}}
<table>
	<thead>
		<tr><th>{{.Name}}</th></tr>
	</thead>
	<tbody>
		{{- range .Arguments}}
		<tr>
			<td><strong>{{.Name}}</strong>: {{.Value.String | desc}}</td>
		</tr>
		{{- end}}
	</tbody>
</table>
{{- end}}
{{- end}}
`

func FormatMarkdown(dst io.Writer, schema *ast.Schema) error {
	gm := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	t := template.New("")
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
	})
	t = template.Must(t.Parse(mdArguments))
	t = template.Must(t.Parse(mdDirectives))

	// t = template.Must(t.Parse(mdEnum))
	// t = template.Must(t.Parse(mdScalar))
	// t = template.Must(t.Parse(mdUnion))
	// t = template.Must(t.Parse(mdInput))

	// s := make([]string, 0, len(schema.PossibleTypes["Contribution"]))
	// for _, def := range schema.PossibleTypes["Contribution"] {
	// 	s = append(s, def.Name)
	// }
	// schema.Types["Contribution"].Types = s
	// t = template.Must(t.Parse(mdInterface))

	// t = template.Must(t.Parse(mdObject))

	return t.Execute(dst, schema.Types["Assignee"])
}
