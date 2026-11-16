// Copyright 2026 Morteza Rajabi. Licensed under Apache-2.0.
// Standalone companion for Goss JSON reports; no third-party Go dependencies.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
)

type check struct {
	Resource string `json:"resource-type"`
	ID string `json:"resource-id"`
	Property string `json:"property"`
	Title string `json:"title"`
	Successful *bool `json:"successful"`
	Skipped bool `json:"skipped"`
	Result *int `json:"result"`
	Status string
}
type report struct {
	Results []check `json:"results"`
	Summary *struct {
		Count int `json:"test-count"`
		Failed int `json:"failed-count"`
		Skipped int `json:"skipped-count"`
	} `json:"summary"`
	Passed, Failed, Skipped int
	Title, Lang, Dir, Heading, PassLabel, FailLabel, SkipLabel, Note string
}

// Parse fails closed on empty, truncated or inconsistent reports. Unknown fields
// are allowed so upstream can add metadata without breaking this companion.
func parse(r io.Reader) (report, error) {
	var out report
	data, err := io.ReadAll(io.LimitReader(r, 16*1024*1024+1))
	if err != nil { return out, err }
	if len(data) > 16*1024*1024 { return out, fmt.Errorf("input exceeds 16 MiB") }
	if err = json.Unmarshal(data, &out); err != nil { return out, fmt.Errorf("invalid Goss JSON: %w", err) }
	out.Passed, out.Failed, out.Skipped = 0, 0, 0
	if out.Summary == nil || len(out.Results) == 0 { return out, fmt.Errorf("report must contain summary and at least one check") }
	for i := range out.Results {
		c := &out.Results[i]
		if c.Result == nil || c.Successful == nil || c.Resource == "" || c.Property == "" { return out, fmt.Errorf("check %d lacks required Goss fields", i+1) }
		switch *c.Result {
		case 0:
			if !*c.Successful || c.Skipped { return out, fmt.Errorf("inconsistent passing check %d", i+1) }
			c.Status = "pass"; out.Passed++
		case 1:
			if *c.Successful || c.Skipped { return out, fmt.Errorf("inconsistent failing check %d", i+1) }
			c.Status = "fail"; out.Failed++
		case 2:
			if !c.Skipped || !*c.Successful { return out, fmt.Errorf("inconsistent skipped check %d", i+1) }
			c.Status = "skip"; out.Skipped++
		default:
			return out, fmt.Errorf("unknown result code in check %d", i+1)
		}
	}
	if out.Summary.Count != len(out.Results) || out.Summary.Failed != out.Failed || out.Summary.Skipped != out.Skipped { return out, fmt.Errorf("summary disagrees with individual checks") }
	return out, nil
}

func render(w io.Writer, r report, lang, title string) error {
	r.Lang, r.Dir, r.Title = lang, "ltr", title
	r.Heading, r.PassLabel, r.FailLabel, r.SkipLabel = "Server validation report", "Passed", "Failed", "Skipped"
	r.Note = "Generated from Goss results. Skipped checks provide no assurance. This report is not a compliance certification. Raw actual/expected values, errors and command output are omitted. Resource names and titles may still contain sensitive information; review before sharing."
	if lang == "fa" {
		r.Dir = "rtl"
		r.Heading, r.PassLabel, r.FailLabel, r.SkipLabel = "گزارش بررسی سرور", "موفق", "ناموفق", "بررسی‌نشده"
		r.Note = "بر پایهٔ نتایج Goss؛ تست ردشده از اجرا اطمینانی ایجاد نمی‌کند. این گزارش گواهی انطباق نیست. مقادیر خام، خطاها و خروجی دستورها نمایش داده نمی‌شوند. نام منابع و عنوان‌ها ممکن است حساس باشند؛ پیش از اشتراک بررسی کنید."
	} else if lang != "en" { return fmt.Errorf("lang must be en or fa") }
	t, err := template.New("report").Parse(page)
	if err != nil { return err }
	return t.Execute(w, r)
}

func run(args []string, stderr io.Writer) int {
	f := flag.NewFlagSet("goss-audit-report", flag.ContinueOnError); f.SetOutput(stderr)
	in := f.String("input", "", "Goss JSON input file")
	out := f.String("output", "", "new HTML output file (never overwritten)")
	lang := f.String("lang", "en", "en or fa")
	title := f.String("title", "Infrastructure audit", "report title")
	if err := f.Parse(args); err != nil { if err == flag.ErrHelp { return 0 }; return 2 }
	if *in == "" || *out == "" || f.NArg() != 0 { fmt.Fprintln(stderr, "provide -input and -output; see -help"); return 2 }
	file, err := os.Open(*in)
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	r, err := parse(file); file.Close()
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	var buf bytes.Buffer
	if err = render(&buf, r, *lang, *title); err != nil { fmt.Fprintln(stderr, err); return 2 }
	dest, err := os.OpenFile(*out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil { fmt.Fprintln(stderr, err); return 2 }
	_, err = dest.Write(buf.Bytes()); closeErr := dest.Close()
	if err != nil || closeErr != nil { fmt.Fprintf(stderr, "write failed: %v; close: %v\n", err, closeErr); return 2 }
	fmt.Fprintf(stderr, "Report saved: %s (passed=%d failed=%d skipped=%d)\n", *out, r.Passed, r.Failed, r.Skipped)
	if r.Failed > 0 { return 1 }
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stderr)) }

const page = `<!doctype html>
<html lang="{{.Lang}}" dir="{{.Dir}}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'">
<title>{{.Title}}</title><style>
:root{color-scheme:dark}*{box-sizing:border-box}body{margin:0;background:#101820;color:#e9f0f2;font:16px/1.6 system-ui,sans-serif}main{max-width:1000px;margin:auto;padding:36px 20px}header{border-bottom:1px solid #35444d;padding-bottom:20px}small{color:#82b6c1}h1{font-size:clamp(24px,4vw,36px);margin:8px 0}h2{font-size:18px;font-weight:400;color:#aec0ca}.stats{display:flex;gap:12px;flex-wrap:wrap;margin:24px 0}.stat{border:1px solid #35444d;padding:14px 24px;min-width:140px;border-radius:10px}.stat strong{display:block;font-size:30px}.pass{color:#8de4b6}.fail{color:#ff9a9a}.skip{color:#ffd887}.table{overflow:auto}table{width:100%;border-collapse:collapse}td,th{text-align:start;padding:12px;border-bottom:1px solid #35444d;vertical-align:top;overflow-wrap:anywhere}code{direction:ltr;unicode-bidi:isolate}footer{margin-top:28px;color:#a5bac5;font-size:13px} @media print{body{background:white;color:black}main{padding:10px}td,th{border-color:#aaa}.pass,.fail,.skip,small,footer,h2{color:black}}
</style></head><body><main><header><small>GOSS / SERVER AUDIT · Morteza Rajabi</small><h1>{{.Title}}</h1><h2>{{.Heading}}</h2></header>
<section class="stats" aria-label="Summary"><div class="stat pass">{{.PassLabel}}<strong>{{.Passed}}</strong></div><div class="stat fail">{{.FailLabel}}<strong>{{.Failed}}</strong></div><div class="stat skip">{{.SkipLabel}}<strong>{{.Skipped}}</strong></div></section>
<div class="table"><table><thead><tr><th>{{if eq .Lang "fa"}}وضعیت{{else}}Status{{end}}</th><th>{{if eq .Lang "fa"}}منبع / ویژگی{{else}}Resource / property{{end}}</th><th>{{if eq .Lang "fa"}}عنوان{{else}}Title{{end}}</th></tr></thead><tbody>{{range .Results}}<tr><td class="{{.Status}}">{{if eq .Status "pass"}}{{$.PassLabel}}{{else if eq .Status "fail"}}{{$.FailLabel}}{{else}}{{$.SkipLabel}}{{end}}</td><td><code>{{.Resource}} · {{.ID}} · {{.Property}}</code></td><td>{{.Title}}</td></tr>{{end}}</tbody></table></div><footer>{{.Note}}<p>Goss: goss-org/goss · Apache-2.0 · Report companion: Morteza Rajabi</p></footer></main></body></html>`
