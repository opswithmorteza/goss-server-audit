# Goss Server Audit — Morteza Rajabi

A small, dependency-free Go companion added to the Apache-2.0 licensed
[Goss](https://github.com/goss-org/goss) project. Original Goss authorship and
license remain intact. Morteza Rajabi's additions are the HTML report companion,
its tests, demo fixture and read-only Linux policy example; the Goss engine was
not authored by this fork's maintainer.

## What is new

- Offline, responsive HTML reports in English or Persian (RTL), also printable.
- Separate passed, failed and skipped counts; skipped tests are never counted as passed.
- Escaped untrusted text and no external scripts, fonts, images or requests.
- Omits raw expected/actual values, errors and command output by design.
- Rejects malformed, oversized, empty and internally inconsistent reports.
- No report overwrite; files created with private permissions on Unix.
- Exit status 0: report generated without failed checks; 1: report generated with failures;
  2: input/usage/output error. Skips remain visible but do not themselves fail the job.

Resource identifiers and titles are included and may contain sensitive information.
Review before sharing. This is a point-in-time validation report, not a security
assessment, vulnerability scan or compliance certification.

## Build and try the demo

Use a Go version compatible with the repository's `go.mod` (currently 1.26.8+).
The companion uses only Go's standard library. It runs on Linux, macOS and Windows;
Goss host validation itself should be run on its supported Linux target.

```sh
go test extras/audit-report.go extras/audit-report_test.go
go build -o goss-audit-report extras/audit-report.go
./goss-audit-report -input extras/audit-demo.json -output demo-en.html -title "DEMO — Server audit"
# Exit 1 is expected: the fictional demo includes one failure.
./goss-audit-report -input extras/audit-demo.json -output demo-fa.html -lang fa -title "نمونهٔ گزارش بررسی سرور"
```

Open the generated HTML in a browser. The demo contains fabricated checks, not
results from a real customer or production host. Windows: use `.exe` for the binary.

Pre-generated examples: [English HTML](extras/audit-demo-en.html) ·
[Persian HTML](extras/audit-demo-fa.html). Download and open locally.

Validation on 2026-11-16: all three companion test groups passed on Windows
with Go 1.27.4; `go vet` and standalone build passed (81.2% statement coverage).
[GitHub Actions tests and builds also passed on Linux and Windows](https://github.com/opswithmorteza/goss-server-audit/actions/runs/43871646926).
Tests cover malformed and
oversized inputs, inconsistent counts, unknown result codes, HTML escaping,
raw-value omission, Persian RTL, failure exit status and overwrite protection.
The Linux security policy is an example; it has not been validated on a real
target host in this work. Full upstream Goss integration tests are separate.

## Collect real results on Linux

Install a reviewed, pinned Goss release using upstream instructions. Inspect
`extras/audit-linux.yaml` before running: it assumes systemd, auditd and
`/usr/sbin/sshd`; it reads SSH configuration and may require root privileges.
It makes no configuration changes and performs no remediation. Goss can execute
commands from policy files, so only run policies you trust. `sshd -T` checks the
default effective policy; host/user-specific Match rules need additional tests.

```sh
# Run in a private directory; JSON may contain raw command output or secrets.
umask 077
rc=0
sudo goss -g extras/audit-linux.yaml validate --format json > results.json || rc=$?
if [ "$rc" -gt 1 ]; then echo "Goss execution error" >&2; exit "$rc"; fi
./goss-audit-report -input results.json -output audit.html -title "Linux baseline review"
```

Do not hide the final command's exit status in CI. Upload only reviewed reports;
do not publish `results.json`. Use unique output filenames per run. Upstream
Goss JSON format is the integration boundary (`outputs/json.go` and
`resource/validate.go`); this companion does not alter Goss validation behavior.

## فارسی

این افزونهٔ سبک با Go، نتایج Goss را به گزارش HTML فارسی یا انگلیسی تبدیل
می‌کند. شمارش موفق، ناموفق و بررسی‌نشده جداست. خروجی آفلاین است و مقدارهای
خام و خروجی دستورها را نمایش نمی‌دهد؛ نام منابع و عنوان‌ها همچنان باید پیش
از اشتراک بررسی شوند. فایل نمونهٔ لینوکس فقط بررسی می‌کند و تنظیمات سرور را
تغییر نمی‌دهد. نمونهٔ نمایشی ساختگی است و نشان‌دهندهٔ تست سرور واقعی نیست.

هستهٔ Goss متعلق به سازندگان اصلی است. توسعهٔ گزارش‌ساز، تست‌ها و نمونه‌های
این بخش توسط مرتضی رجبی و تحت همان مجوز Apache-2.0 ارائه می‌شود.
