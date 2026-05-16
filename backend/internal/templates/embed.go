package templates

import (
	"embed"
	"html/template"
)

//go:embed report.html
var reportFS embed.FS

// ReportTemplate is the parsed HTML template for PDF reports.
// Uses Go's html/template for automatic HTML escaping (never template.HTML).
var ReportTemplate *template.Template

func init() {
	ReportTemplate = template.Must(template.ParseFS(reportFS, "report.html"))
}
