package reporting

import (
	"bytes"
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
	"linux-hardener/internal/distro"
	"linux-hardener/internal/engine"
	"linux-hardener/internal/iface"
)

type FileDiff struct {
	Path string `json:"path"`
	Old  string `json:"old"`
	New  string `json:"new"`
}

type Report struct {
	Timestamp string            `json:"timestamp"`
	Host      string            `json:"host"`
	Distro    distro.InfoData   `json:"distro"`
	Policy    map[string]any    `json:"policy"`
	Diffs     []FileDiff        `json:"diffs"`
	Warnings  []string          `json:"warnings"`
	Meta      map[string]string `json:"meta"`
}

func Build(host string, ctx engine.Context, res iface.DryRunResult) Report {
	var diffs []FileDiff
	for _, d := range res.Diffs {
		diffs = append(diffs, FileDiff{Path: d.Path, Old: d.Old, New: d.New})
	}
	return Report{
		Timestamp: time.Now().Format(time.RFC3339),
		Host:      host,
		Distro:    ctx.Distro,
		Policy:    map[string]any{"name": ctx.Policy.Name, "profile": ctx.Policy.Profile},
		Diffs:     diffs,
		Warnings:  res.Warnings,
	}
}

func SaveJSON(r Report, out string) error {
	_ = os.MkdirAll(filepath.Dir(out), 0o755)
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil { return err }
	return os.WriteFile(out, b, 0o644)
}

func SaveHTML(r Report, out string) error {
	_ = os.MkdirAll(filepath.Dir(out), 0o755)
	tmpl := `<!doctype html><html><head><meta charset="utf-8"><title>Linux Hardener Report</title></head><body>
	<h1>Linux Hardener Report</h1>
	<p><b>Host:</b> {{.Host}}<br/>
	<b>Distro:</b> {{.Distro.Name}} {{.Distro.Version}}<br/>
	<b>Time:</b> {{.Timestamp}}</p>
	<h2>Diffs</h2>
	{{range .Diffs}}
	<h3>{{.Path}}</h3>
	<pre>--- OLD ---
{{.Old}}
--- NEW ---
{{.New}}</pre>
	{{end}}
	</body></html>`
	t, err := template.New("r").Parse(tmpl)
	if err != nil { return err }
	f, err := os.Create(out)
	if err != nil { return err }
	defer f.Close()
	return t.Execute(f, r)
}

func SavePDF(r Report, out string) error {
	_ = os.MkdirAll(filepath.Dir(out), 0o755)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Linux Hardener Report")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Host: "+r.Host)
	pdf.Ln(6)
	pdf.Cell(40, 10, "Distro: "+r.Distro.Name+" "+r.Distro.Version)
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Diffs:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	for _, d := range r.Diffs {
		pdf.MultiCell(0, 5, d.Path, "", "L", false)
		var buf bytes.Buffer
		buf.WriteString("--- OLD ---\n")
		buf.WriteString(d.Old)
		buf.WriteString("\n--- NEW ---\n")
		buf.WriteString(d.New)
		pdf.MultiCell(0, 4, buf.String(), "", "L", false)
		pdf.Ln(3)
	}
	return pdf.OutputFileAndClose(out)
}
