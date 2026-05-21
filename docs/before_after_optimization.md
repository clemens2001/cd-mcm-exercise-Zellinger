
# Before Optimization

Before the optimization, the Dockerfile used the `alpine:3.19` base image for the runtime stage which still includes a number of packages that can introduce vulnerabilities. The `trivy` scan of the image revealed that there were no vulnerabilities in the application binary itself.

```shell
(base) clemenszellinger@Mac cd-mcm-exercise-Zellinger % trivy image product-catalog:latest
2026-05-21T16:41:12+02:00       INFO    [vuln] Vulnerability scanning is enabled
2026-05-21T16:41:12+02:00       INFO    [secret] Secret scanning is enabled
2026-05-21T16:41:12+02:00       INFO    [secret] If your scanning is slow, please try '--scanners vuln' to disable secret scanning
2026-05-21T16:41:12+02:00       INFO    [secret] Please see https://trivy.dev/docs/v0.70/guide/scanner/secret#recommendation for faster secret detection
2026-05-21T16:41:12+02:00       INFO    Detected OS     family="alpine" version="3.19.9"
2026-05-21T16:41:12+02:00       INFO    [alpine] Detecting vulnerabilities...   os_version="3.19" repository="3.19" pkg_num=16
2026-05-21T16:41:12+02:00       INFO    Number of language-specific files       num=1
2026-05-21T16:41:12+02:00       INFO    [gobinary] Detecting vulnerabilities...
2026-05-21T16:41:12+02:00       WARN    Using severities from other vendors for some vulnerabilities. Read https://trivy.dev/docs/v0.70/guide/scanner/vulnerability#severity-selection for details.
2026-05-21T16:41:12+02:00       WARN    This OS version is no longer supported by the distribution      family="alpine" version="3.19.9"
2026-05-21T16:41:12+02:00       WARN    The vulnerability detection may be insufficient because security updates are not provided

Report Summary

┌────────────────────────────────────────┬──────────┬─────────────────┬─────────┐
│                 Target                 │   Type   │ Vulnerabilities │ Secrets │
├────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ product-catalog:latest (alpine 3.19.9) │  alpine  │       10        │    -    │
├────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ app/api-server                         │ gobinary │        0        │    -    │
└────────────────────────────────────────┴──────────┴─────────────────┴─────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)


product-catalog:latest (alpine 3.19.9)

Total: 10 (UNKNOWN: 0, LOW: 3, MEDIUM: 5, HIGH: 2, CRITICAL: 0)

┌───────────────┬────────────────┬──────────┬────────┬──────────────────────┬──────────────────────┬──────────────────────────────────────────────────────────────┐
│    Library    │ Vulnerability  │ Severity │ Status │  Installed Version   │    Fixed Version     │                            Title                             │
├───────────────┼────────────────┼──────────┼────────┼──────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────┤
│ busybox       │ CVE-2024-58251 │ MEDIUM   │ fixed  │ 1.36.1-r20           │ 1.36.1-r21           │ In netstat in BusyBox through 1.37.0, local users can launch │
│               │                │          │        │                      │                      │ of networ...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2024-58251                   │
│               ├────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│               │ CVE-2025-46394 │ LOW      │        │                      │                      │ In tar in BusyBox through 1.37.0, a TAR archive can have     │
│               │                │          │        │                      │                      │ filenames...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2025-46394                   │
├───────────────┼────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│ busybox-binsh │ CVE-2024-58251 │ MEDIUM   │        │                      │                      │ In netstat in BusyBox through 1.37.0, local users can launch │
│               │                │          │        │                      │                      │ of networ...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2024-58251                   │
│               ├────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│               │ CVE-2025-46394 │ LOW      │        │                      │                      │ In tar in BusyBox through 1.37.0, a TAR archive can have     │
│               │                │          │        │                      │                      │ filenames...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2025-46394                   │
├───────────────┼────────────────┼──────────┤        ├──────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────┤
│ musl          │ CVE-2026-40200 │ HIGH     │        │ 1.2.4_git20230717-r5 │ 1.2.4_git20230717-r6 │ musl: musl libc: Arbitrary code execution and denial of      │
│               │                │          │        │                      │                      │ service via stack-based...                                   │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2026-40200                   │
│               ├────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│               │ CVE-2026-6042  │ MEDIUM   │        │                      │                      │ musl libc: GB18030 4-byte Decoder: musl libc: Denial of      │
│               │                │          │        │                      │                      │ Service via inefficient...                                   │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2026-6042                    │
├───────────────┼────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│ musl-utils    │ CVE-2026-40200 │ HIGH     │        │                      │                      │ musl: musl libc: Arbitrary code execution and denial of      │
│               │                │          │        │                      │                      │ service via stack-based...                                   │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2026-40200                   │
│               ├────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│               │ CVE-2026-6042  │ MEDIUM   │        │                      │                      │ musl libc: GB18030 4-byte Decoder: musl libc: Denial of      │
│               │                │          │        │                      │                      │ Service via inefficient...                                   │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2026-6042                    │
├───────────────┼────────────────┤          │        ├──────────────────────┼──────────────────────┼──────────────────────────────────────────────────────────────┤
│ ssl_client    │ CVE-2024-58251 │          │        │ 1.36.1-r20           │ 1.36.1-r21           │ In netstat in BusyBox through 1.37.0, local users can launch │
│               │                │          │        │                      │                      │ of networ...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2024-58251                   │
│               ├────────────────┼──────────┤        │                      │                      ├──────────────────────────────────────────────────────────────┤
│               │ CVE-2025-46394 │ LOW      │        │                      │                      │ In tar in BusyBox through 1.37.0, a TAR archive can have     │
│               │                │          │        │                      │                      │ filenames...                                                 │
│               │                │          │        │                      │                      │ https://avd.aquasec.com/nvd/cve-2025-46394                   │
└───────────────┴────────────────┴──────────┴────────┴──────────────────────┴──────────────────────┴──────────────────────────────────────────────────────────────┘
```

# After optimization: No vulnerabilities detected in the final image

After the optimization, the Dockerfile was changed to use `scratch` as the base image for the runtime stage, which is an empty image with no packages. This means that there are no OS-level vulnerabilities in the final image. The only file included in the final image is the application binary itself, which was already found to be free of vulnerabilities.

```shell
(base) clemenszellinger@Mac cd-mcm-exercise-Zellinger % trivy image product-catalog:latest      
2026-05-21T16:47:26+02:00       INFO    [vuln] Vulnerability scanning is enabled
2026-05-21T16:47:26+02:00       INFO    [secret] Secret scanning is enabled
2026-05-21T16:47:26+02:00       INFO    [secret] If your scanning is slow, please try '--scanners vuln' to disable secret scanning
2026-05-21T16:47:26+02:00       INFO    [secret] Please see https://trivy.dev/docs/v0.70/guide/scanner/secret#recommendation for faster secret detection
2026-05-21T16:47:26+02:00       INFO    Number of language-specific files       num=1
2026-05-21T16:47:26+02:00       INFO    [gobinary] Detecting vulnerabilities...

Report Summary

┌────────────────┬──────────┬─────────────────┬─────────┐
│     Target     │   Type   │ Vulnerabilities │ Secrets │
├────────────────┼──────────┼─────────────────┼─────────┤
│ app/api-server │ gobinary │        0        │    -    │
└────────────────┴──────────┴─────────────────┴─────────┘
Legend:
- '-': Not scanned
- '0': Clean (no security findings detected)
```