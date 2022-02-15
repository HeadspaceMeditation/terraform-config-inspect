package tfconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/template"
)

func RenderMarkdown(w io.Writer, module *Module) error {
	tmpl := template.New("md")
	tmpl.Funcs(template.FuncMap{
		"tt": func(i interface{}) string {
			var s string
			switch i.(type) {
			case float64, float32:
				s = fmt.Sprintf("%.0f", i)
			case nil:
				return "``"
			default:
				s = fmt.Sprintf("%v", i)
			}
			return "`" + s + "`"
		},
		"req": func(i interface{}) string {
			switch i.(type) {
			case nil:
				return "yes"
			default:
				return "no"
			}
		},
		"commas": func(s []string) string {
			return strings.Join(s, ", ")
		},
		"json": func(v interface{}) (string, error) {
			j, err := json.Marshal(v)
			return string(j), err
		},
		"skip": func(p tfconfig.SourcePos) bool {
			blacklist := []string{"environment.tf.json", "global-variables.tf.json", "account-variables.tf.json"}

			for _, b := range blacklist {
				if strings.HasSuffix(p.Filename, b) {
					return false
				}
			}
			return true
		},
		"severity": func(s tfconfig.DiagSeverity) string {
			switch s {
			case tfconfig.DiagError:
				return "Error: "
			case tfconfig.DiagWarning:
				return "Warning: "
			default:
				return ""
			}
		},
		"strip_newlines": func(input string) string {
			re := regexp.MustCompile(`\r?\n`)
			return re.ReplaceAllString(input, "<br />")
		},
	})
	template.Must(tmpl.Parse(markdownTemplate))
	return tmpl.Execute(w, module)
}

const markdownTemplate = `
## Inputs
| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
{{- range .Variables }}{{if skip .Pos }}
| {{ tt .Name }} | {{- if .Description}}{{ strip_newlines .Description }}{{ end }} | {{- if .Type}}{{ .Type }}{{ end }} | {{ tt .Default }} | {{req .Default }} |{{end}}{{end}}

{{- if .Outputs}}

## Outputs
| Name | Description |
|------|-------------|
{{- range .Outputs }}
| {{ tt .Name }} | {{ if .Description}}{{ strip_newlines .Description }}{{ end }} |
{{- end}}{{end}}

{{- if .ManagedResources}}

Managed Resources
-----------------
{{- range .ManagedResources }}
* {{ printf "%s.%s" .Type .Name | tt }}
{{- end}}{{end}}

{{- if .DataResources}}

Data Resources
--------------
{{- range .DataResources }}
* {{ printf "data.%s.%s" .Type .Name | tt }}
{{- end}}{{end}}

{{- if .ModuleCalls}}

Child Modules
-------------
{{- range .ModuleCalls }}
* {{ tt .Name }} from {{ tt .Source }}{{ if .Version }} ({{ tt .Version }}){{ end }}
{{- end}}{{end}}

{{- if .Diagnostics}}

Problems
-------------
{{- range .Diagnostics }}

{{ severity .Severity }}{{ .Summary }}{{ if .Pos }}
-------------

(at {{ tt .Pos.Filename }} line {{ .Pos.Line }}{{ end }})
{{ if .Detail }}
{{ .Detail }}
{{- end }}

{{- end}}{{end}}
`
