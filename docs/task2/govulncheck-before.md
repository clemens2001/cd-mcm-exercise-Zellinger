## Output of `govulncheck` before fixing the vulnerability:

```
=== Symbol Results ===

Vulnerability #1: GO-2026-4971
    Panic in Dial and LookupPort when handling NUL byte on Windows in net
  More info: https://pkg.go.dev/vuln/GO-2026-4971
  Standard library
    Found in: net@go1.26.2
    Fixed in: net@go1.26.3
    Example traces found:
      #1: internal/store/postgres.go:53:15: store.PostgresStore.GetAll calls sql.Rows.Next, which eventually calls net.Dialer.Dial
      #2: internal/store/postgres.go:53:15: store.PostgresStore.GetAll calls sql.Rows.Next, which eventually calls net.Dialer.DialContext
      #3: cmd/api/main.go:51:31: api.main calls http.ListenAndServe, which eventually calls net.Listen

Your code is affected by 1 vulnerability from the Go standard library.
This scan also found 2 vulnerabilities in packages you import and 5
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```

## Task 2: Dependency Scanning

**Vulnerability Found:**
* **CVE/ID:** GO-2026-4971
* **Affected Package:** `net` (Go Standard Library)
* **Found in Version:** go1.26.2
* **Fixed in Version:** go1.26.3

**Resolution:**

To resolve the vulnerability across development, containerization, and continuous integration, the toolchain configuration was upgraded globally to the patched patch-release version:
1. **`go.mod`**: Updated the module language target environment reference line to `go 1.26.3`.
2. **`Dockerfile`**: Changed the initial compiler layer to pull `FROM golang:1.26.3-alpine AS builder`.
3. **`ci.yml`**: Configured the GitHub Actions matrix strategy and `setup-go` action configurations to target `1.26.3` explicitly.

Now `govulncheck ./...` confirmed zero remaining vulnerabilities:

```
% ~/go/bin/govulncheck ./...
No vulnerabilities found.
```
