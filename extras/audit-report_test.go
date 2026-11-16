package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `{"results":[{"resource-type":"file","resource-id":"/etc/shadow","property":"mode","title":"<script>alert(1)</script>","successful":true,"skipped":false,"result":0,"matcher-result":{"actual":"SECRET"}},{"resource-type":"port","resource-id":"tcp:23","property":"listening","successful":false,"skipped":false,"result":1},{"resource-type":"service","resource-id":"demo","property":"running","successful":true,"skipped":true,"result":2}],"summary":{"test-count":3,"failed-count":1,"skipped-count":1}}`

func TestParseAndRender(t *testing.T) {
	r, err := parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if r.Passed != 1 || r.Failed != 1 || r.Skipped != 1 {
		t.Fatal("wrong counts")
	}
	for _, lang := range []string{"en", "fa"} {
		var out bytes.Buffer
		if err := render(&out, r, lang, "<img src=x onerror=alert(1)>"); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(out.String(), "<script>") || strings.Contains(out.String(), "<img src=x") || strings.Contains(out.String(), "SECRET") {
			t.Fatal("unsafe/raw values in report")
		}
		if !strings.Contains(out.String(), "&lt;script&gt;") {
			t.Fatal("title not escaped")
		}
		if lang == "fa" && !strings.Contains(out.String(), `dir="rtl"`) {
			t.Fatal("missing RTL")
		}
	}
}

func TestRejectMalformedOrInconsistentInput(t *testing.T) {
	cases := []string{"", "{}", sample + "{}", strings.Replace(sample, `"test-count":3`, `"test-count":4`, 1), strings.Replace(sample, `"result":1`, `"result":99`, 1), strings.Replace(sample, `"successful":false`, `"successful":true`, 1), strings.Replace(sample, `"result":0`, `"result":null`, 1), `{"results":[],"summary":{"test-count":0}}`}
	for _, input := range cases {
		if _, err := parse(strings.NewReader(input)); err == nil {
			t.Errorf("accepted invalid input: %.100s", input)
		}
	}
	if _, err := parse(strings.NewReader(strings.Repeat(" ", 16*1024*1024+1))); err == nil {
		t.Fatal("oversized input accepted")
	}
}

func TestCLIExitCodesAndNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "input.json")
	out := filepath.Join(dir, "report.html")
	if err := os.WriteFile(in, []byte(sample), 0600); err != nil {
		t.Fatal(err)
	}
	if got := run([]string{"-input", in, "-output", out}, io.Discard); got != 1 {
		t.Fatalf("failed-check exit=%d", got)
	}
	before, _ := os.ReadFile(out)
	if got := run([]string{"-input", in, "-output", out}, io.Discard); got != 2 {
		t.Fatal("existing report overwritten")
	}
	after, _ := os.ReadFile(out)
	if !bytes.Equal(before, after) {
		t.Fatal("existing content changed")
	}
	if got := run([]string{"-input", in, "-output", filepath.Join(dir, "bad.html"), "-lang", "xx"}, io.Discard); got != 2 {
		t.Fatal("invalid language accepted")
	}
	passing := `{"results":[{"resource-type":"file","property":"exists","result":0,"successful":true}],"summary":{"test-count":1,"failed-count":0,"skipped-count":0}}`
	os.WriteFile(in, []byte(passing), 0600)
	if got := run([]string{"-input", in, "-output", filepath.Join(dir, "pass.html")}, io.Discard); got != 0 {
		t.Fatalf("passing exit=%d", got)
	}
}
