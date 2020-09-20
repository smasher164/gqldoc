package gqldoc

// precompile templates

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
- [{{.}}]({{index $.Anchor "type" .}})
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
				<strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
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
				<strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
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
				<strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
			</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" (wrap .Arguments $.Anchor) 3 | minify}}
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

{{- if implementers .Definition}}
#### Implemented by
{{- range implementers .Definition}}
- [<code>{{.Name}}</code>]({{index $.Anchor "type" .Name}})
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
				<strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)
			</td>
			<td>{{.Description | desc}}
			{{- if .Directives}}
			{{indentTemplate "directives" . 3 | minify}}
			{{- end}}
			{{- if .Arguments}}
			{{indentTemplate "arguments" (wrap .Arguments $.Anchor) 3 | minify}}
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
- [<code>{{.}}</code>]({{index $.Anchor "type" .}})
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
			<td><strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
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
**Type:** [{{.Type}}]({{index $.Anchor "type" .Type.Name}})

{{.Description | desc}}

{{- if .Directives}}
{{indentTemplate "directives" . 0 | minify}}
{{end}}

{{- if .Arguments}}
#### Arguments
{{indentTemplate "tableQueries" (wrap . $.Anchor) 0 | minify}}
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
			<td><strong>{{.Name}}</strong> (<a href="{{index $.Anchor "type" .Type.Name}}"><strong>{{.Type}}</strong></a>)</td>
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
- <code>{{.Name}}</code>([<code>{{.Type}}</code>]({{index $.Anchor "type" .Type.Name}}))
{{- end}}
{{- end}}

{{- if fields .Type}}
#### Return fields
{{indentTemplate "tableMutations" (wrap . $.Anchor) 0 | minify}}
{{- end}}

---
{{end -}}
`

const mdSchema = `# Reference
{{- template "toc" .}}

{{if .Query -}}
## Queries
{{template "queries" (wrap .Query $.Anchor)}}{{end -}}
{{if .Mutation -}}
## Mutations
{{template "mutations" (wrap .Mutation $.Anchor)}}{{end -}}
{{if .Subscription -}}
## Subscriptions
{{template "queries" (wrap .Subscription $.Anchor)}}{{end -}}
{{if .Objects}}
## Objects
{{range .Objects}}{{template "object" (wrap . $.Anchor)}}
---
{{end -}}
{{end -}}
{{if .Interfaces}}
## Interfaces
{{range .Interfaces}}{{template "interface" (wrap . $.Anchor)}}
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
{{range .Unions}}{{template "union" (wrap . $.Anchor)}}
---
{{end -}}
{{end -}}
{{if .Inputs}}
## Input objects
{{range .Inputs}}{{template "input" (wrap . $.Anchor)}}
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
	Anchor map[string]map[string]string
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
	return res
}

func (md *md) updateAnchor(key, typ string, refs ...string) {
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

func (md *md) updateAnchors(name, ref string, defs interface{}) {
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

var mdTemplate *template.Template

func init() {
	// gm := goldmark.New(
	// 	goldmark.WithRendererOptions(
	// 		html.WithHardWraps(),
	// 	),
	// )
}

func FormatMarkdown(dst io.Writer, schema *ast.Schema) error {
	md := &md{
		count:  make(map[string]int),
		Anchor: make(map[string]map[string]string),
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
		"implementers": func(def *ast.Definition) []*ast.Definition {
			return schema.GetPossibleTypes(def)
		},
		"fields": func(T *ast.Type) ast.FieldList {
			return schema.Types[T.Name()].Fields
		},
		"minify": func(s string) (string, error) {
			m := minify.New()
			m.AddFunc("text/html", mhtml.Minify)
			return m.String("text/html", s)
		},
		"wrap": func(v interface{}, anchor map[string]map[string]string) interface{} {
			switch v := v.(type) {
			case *ast.Definition:
				r := struct {
					*ast.Definition
					Anchor map[string]map[string]string
				}{v, anchor}
				return r
			case *ast.FieldDefinition:
				r := struct {
					*ast.FieldDefinition
					Anchor map[string]map[string]string
				}{v, anchor}
				return r
			case ast.ArgumentDefinitionList:
				r := struct {
					ast.ArgumentDefinitionList
					Anchor map[string]map[string]string
				}{v, anchor}
				return r
			}
			return nil
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
	template.Must(t.New("toc").Parse(mdTOC))
	template.Must(t.Parse(mdSchema))
	return t.Execute(dst, md)
}
