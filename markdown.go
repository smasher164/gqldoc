package gqldoc

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
- [{{.}}]({{index $.MD.Anchor "type" .}})
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
			<td>
				<strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
			</td>
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
	{{- range .ArgumentDefinitionList}}
		<tr>
			<td>
				<strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
				<br>
				{{- wordwrap .Description 69 | desc}}
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
			<td>
				<strong>{{.Name}}</strong>: {{len .Name | add 1 | sub 69 | wordwrap .Value.String | desc}}
			</td>
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
			<td>
				<strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
			</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" (wrapMD .Arguments $.MD) 3 | minify}}
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

{{- if index $.MD.Implementers .Definition.Name}}
#### Implemented by
{{- range index $.MD.Implementers .Definition.Name}}
- [<code>{{.Name}}</code>]({{index $.MD.Anchor "type" .Name}})
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
			<td>
				<strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
			</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" (wrapMD .Arguments $.MD) 3 | minify}}
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
- [<code>{{.}}</code>]({{index $.MD.Anchor "type" .}})
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
			<td><strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
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
**Type:** [{{.Type}}]({{index $.MD.Anchor "type" .Type.Name}})

{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if .Arguments}}
#### Arguments
{{indentTemplate "tableQueries" (wrapMD . $.MD) 0 | minify}}
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
	{{- range (index $.MD.Types .Type.Name).Fields}}
		<tr>
			<td><strong>{{.Name}}</strong> (<a href="{{index $.MD.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
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
- <code>{{.Name}}</code>([<code>{{.Type}}</code>]({{index $.MD.Anchor "type" .Type.Name}}))
{{- end}}
{{- end}}

{{- if (index $.MD.Types .Type.Name).Fields}}
#### Return fields
{{indentTemplate "tableMutations" (wrapMD . $.MD) 0 | minify}}
{{- end}}

---
{{end -}}
`

const mdSchema = `# Reference
{{- template "toc" .}}

{{if .Query -}}
## Queries
{{template "queries" (wrapMD .Query $)}}{{end -}}
{{if .Mutation -}}
## Mutations
{{template "mutations" (wrapMD .Mutation $)}}{{end -}}
{{if .Subscription -}}
## Subscriptions
{{template "queries" (wrapMD .Subscription $)}}{{end -}}
{{if .Objects}}
## Objects
{{range .Objects}}{{template "object" (wrapMD . $)}}
---
{{end -}}
{{end -}}
{{if .Interfaces}}
## Interfaces
{{range .Interfaces}}{{template "interface" (wrapMD . $)}}
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
{{range .Unions}}{{template "union" (wrapMD . $)}}
---
{{end -}}
{{end -}}
{{if .Inputs}}
## Input objects
{{range .Inputs}}{{template "input" (wrapMD . $)}}
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

const mdTOC = `{{with .Query}}
- [Queries]({{index $.Anchor "type" "Query"}})
{{- range .Fields}}
	- [{{.Name}}]({{index $.Anchor "field" .Name}})
{{- end}}
{{- end}}
{{- with .Mutation}}
- [Mutations]({{index $.Anchor "type" "Mutation"}})
{{- range .Fields}}
	- [{{.Name}}]({{index $.Anchor "field" .Name}})
{{- end}}
{{- end}}
{{- with .Subscription}}
- [Subscriptions]({{index $.Anchor "type" "Subscription"}})
{{- range .Fields}}
	- [{{.Name}}]({{index $.Anchor "field" .Name}})
{{- end}}
{{- end}}
{{- with .Objects}}
- [Objects]({{index $.Anchor "type" "Objects"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
{{- with .Interfaces}}
- [Interfaces]({{index $.Anchor "type" "Interfaces"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
{{- with .Enums}}
- [Enums]({{index $.Anchor "type" "Enums"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
{{- with .Unions}}
- [Unions]({{index $.Anchor "type" "Unions"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
{{- with .Inputs}}
- [Input objects]({{index $.Anchor "type" "Input objects"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
{{- with .Scalars}}
- [Scalars]({{index $.Anchor "type" "Scalars"}})
{{- range .}}
	- [{{.Name}}]({{index $.Anchor "type" .Name}})
{{- end}}
{{- end}}
`

var mdTemplate *template.Template

func init() {
	gm := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)
	mdTemplate = template.New("schema")
	mdTemplate = mdTemplate.Funcs(template.FuncMap{
		"indentTemplate": func(name string, v interface{}, tabs int) (string, error) {
			var buf strings.Builder
			if err := mdTemplate.ExecuteTemplate(&buf, name, v); err != nil {
				return "", err
			}
			prefix := strings.Repeat("\t", tabs)
			split := strings.Split(buf.String(), "\n")
			for i := range split {
				if i != 0 {
					split[i] = prefix + split[i]
				}
			}
			return strings.Join(split, "\n"), nil
		},
		"desc": func(input string) (string, error) {
			var buf strings.Builder
			err := gm.Convert([]byte(input), &buf)
			s := strings.TrimSuffix(buf.String(), "\n")
			s = strings.TrimPrefix(s, "<p>")
			s = strings.TrimSuffix(s, "</p>")
			return s, err
		},
		"wordwrap": func(s string, i int) string {
			return wordwrap.WrapString(s, uint(i))
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"minify": func(s string) (string, error) {
			m := minify.New()
			m.AddFunc("text/html", mhtml.Minify)
			return m.String("text/html", s)
		},
		"wrapMD": func(v interface{}, md *markdown) interface{} {
			switch v := v.(type) {
			case *ast.Definition:
				r := struct {
					*ast.Definition
					MD *markdown
				}{v, md}
				return r
			case *ast.FieldDefinition:
				r := struct {
					*ast.FieldDefinition
					MD *markdown
				}{v, md}
				return r
			case ast.ArgumentDefinitionList:
				r := struct {
					ast.ArgumentDefinitionList
					MD *markdown
				}{v, md}
				return r
			}
			return nil
		},
	})
	template.Must(mdTemplate.New("arguments").Parse(mdArguments))
	template.Must(mdTemplate.New("directives").Parse(mdDirectives))
	template.Must(mdTemplate.New("scalar").Parse(mdScalar))
	template.Must(mdTemplate.New("tableObject").Parse(mdTableObject))
	template.Must(mdTemplate.New("tableQueries").Parse(mdTableQueries))
	template.Must(mdTemplate.New("tableMutations").Parse(mdTableMutations))
	template.Must(mdTemplate.New("tableInput").Parse(mdTableInput))
	template.Must(mdTemplate.New("tableInterface").Parse(mdTableInterface))
	template.Must(mdTemplate.New("object").Parse(mdObject))
	template.Must(mdTemplate.New("interface").Parse(mdInterface))
	template.Must(mdTemplate.New("union").Parse(mdUnion))
	template.Must(mdTemplate.New("enum").Parse(mdEnum))
	template.Must(mdTemplate.New("input").Parse(mdInput))
	template.Must(mdTemplate.New("queries").Parse(mdQueries))
	template.Must(mdTemplate.New("mutations").Parse(mdMutations))
	template.Must(mdTemplate.New("toc").Parse(mdTOC))
	template.Must(mdTemplate.Parse(mdSchema))
}

type markdown struct {
	Query        *ast.Definition
	Mutation     *ast.Definition
	Subscription *ast.Definition
	Objects      []*ast.Definition
	Interfaces   []*ast.Definition
	Enums        []*ast.Definition
	Unions       []*ast.Definition
	Inputs       []*ast.Definition
	Scalars      []*ast.Definition

	count        map[string]int
	Anchor       map[string]map[string]string
	Implementers map[string][]*ast.Definition
	Types        map[string]*ast.Definition
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

func (md *markdown) filterFields(def *ast.Definition) *ast.Definition {
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
	def.Fields = res
	return def
}

func (md *markdown) filterKind(fields map[string]*ast.Definition, kind ast.DefinitionKind) []*ast.Definition {
	res := make([]*ast.Definition, 0, len(fields))
	for _, field := range fields {
		if field.Kind == kind && valid(field) {
			res = append(res, field)
		}
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res
}

func (md *markdown) updateAnchor(key, typ string, refs ...string) {
	k := strings.ReplaceAll(strings.ToLower(key), " ", "-")
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
	if _, ok := md.Anchor[typ]; !ok {
		md.Anchor[typ] = make(map[string]string)
	}
	md.Anchor[typ][key] = ref
	md.count[k]++
}

func (md *markdown) updateAnchors(name, ref string, defs interface{}) {
	switch defs := defs.(type) {
	case *ast.Definition:
		if defs == nil || len(defs.Fields) == 0 {
			return
		}
		md.updateAnchor(name, "type", ref)
		for _, field := range defs.Fields {
			md.updateAnchor(field.Name, "field")
		}
	case []*ast.Definition:
		if len(defs) == 0 {
			return
		}
		md.updateAnchor(name, "type", ref)
		for _, field := range defs {
			md.updateAnchor(field.Name, "type")
		}
	}
}

func FormatMarkdown(dst io.Writer, schema *ast.Schema) error {
	md := &markdown{
		count:        make(map[string]int),
		Anchor:       make(map[string]map[string]string),
		Implementers: schema.PossibleTypes,
		Types:        schema.Types,
	}
	md.Query = md.filterFields(schema.Query)
	md.updateAnchors("Query", "#queries", md.Query)
	md.Mutation = md.filterFields(schema.Mutation)
	md.updateAnchors("Mutation", "#mutations", md.Mutation)
	md.Subscription = md.filterFields(schema.Subscription)
	md.updateAnchors("Subscription", "#subscriptions", md.Subscription)
	md.Objects = md.filterKind(schema.Types, ast.Object)
	md.updateAnchors("Objects", "#objects", md.Objects)
	md.Interfaces = md.filterKind(schema.Types, ast.Interface)
	md.updateAnchors("Interfaces", "#interfaces", md.Interfaces)
	md.Enums = md.filterKind(schema.Types, ast.Enum)
	md.updateAnchors("Enums", "#enums", md.Enums)
	md.Unions = md.filterKind(schema.Types, ast.Union)
	md.updateAnchors("Unions", "#unions", md.Unions)
	md.Inputs = md.filterKind(schema.Types, ast.InputObject)
	md.updateAnchors("Input objects", "#input-objects", md.Inputs)
	md.Scalars = md.filterKind(schema.Types, ast.Scalar)
	md.updateAnchors("Scalars", "#scalars", md.Scalars)
	return mdTemplate.Execute(dst, md)
}
