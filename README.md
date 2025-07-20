<img src="https://github.com/symonk/vessel/blob/main/.github/images/vessel.png" border="1" width="275" height="275"/>

[![GoDoc](https://pkg.go.dev/badge/github.com/symonk/vessel)](https://pkg.go.dev/github.com/symonk/vessel)
[![Build Status](https://github.com/symonk/vessel/actions/workflows/go_test.yml/badge.svg)](https://github.com/symonk/vessel/actions/workflows/go_test.yml)
[![codecov](https://codecov.io/gh/symonk/vessel/branch/main/graph/badge.svg)](https://codecov.io/gh/symonk/vessel)
[![Go Report Card](https://goreportcard.com/badge/github.com/symonk/vessel)](https://goreportcard.com/report/github.com/symonk/vessel)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/symonk/vessel/blob/master/LICENSE)

# ⚡ Vessel

**Vessel** is a blazing-fast, HTTP benchmarking tool for testing rest apis.

> ⚠️ **Note**: Vessel is currently in early-phase development and not yet production-ready. Contributions and feedback are welcome!

## 🏁 Features

- Support for HTTP1, HTTP1/1 & HTTP/2
- Real time CLI metrics with status breakdowns, grouped errors, latency and throughput etc
- Store output data in JSON or CSV output formats
- Supports custom auth mechanism, headers and payloads
- Full HTTP method support
- Concurrency and rate limiting controls
- Tunable configuration
- Templating and HTTP Sequences (coming soon)

---

## 📦 Installation

### Precompiled Binaries

Download the latest version from the [Releases](https://github.com/symonk/vessel/releases) page.

### From Source

```bash
go install github.com/symonk/vessel@latest
```

---

## 🚀 Quick Start

```bash
vessel run https://yourwebsite.com -c 50 -d 10s
```

- `-c 50` — 50 concurrent connections
- `-d 10s` — 10-second test duration

### Example with Headers and JSON Payload

```bash
vessel run https://api.yourwebsite.com/data \
  -X POST \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -b '{"name": "vessel"}'
```

---

## 📊 Output Sample

```text
Running 106.30525ms test @ http://localhost:8000 [vessel-v0.0.1]
10 Connections

Summary:
  Requests:             7 (7 per second)
  Duration:             106.30525ms
  Latency:              max=106ms, avg=23.857143ms, p90=11ms, p95=106ms, p99=106ms
  Errors:               Total: 3: Timeout(0), Cancelled(0), Connection(0), Unknown(3)
  BytesReceived:        0.00MB/s
  BytesSent:            0.01MB/s

Breakdown
        [200]: 7


Raw Errors:
Get &#34;http://localhost:8000&#34;: read tcp [::1]:59917-&gt;[::1]:8000: read: connection reset by peer
Get &#34;http://localhost:8000&#34;: write tcp [::1]:59916-&gt;[::1]:8000: write: broken pipe
Get &#34;http://localhost:8000&#34;: write tcp [::1]:59918-&gt;[::1]:8000: write: broken pipe
```

---

## ⚙️ Options

| Flag         | Description                            |
|--------------|----------------------------------------|
| `-c`         | Number of concurrent workers           |
| `-d`         | Duration of the test (e.g. `10s`, `1m`)|
| `-X`         | HTTP method (GET, POST, etc.)          |
| `-H`         | Custom headers                         |
| `-b`         | Request body payload                   |
| `--output`   | Output format (`json`, `csv`)          |
| `--rate`     | Requests per second                    |
| `--timeout`  | Request timeout                        |

---

## ⚠️ Disclaimer

Vessel is intended solely for **ethical performance testing** of web services you own or have explicit permission to test.  
Any use of this tool for denial-of-service (DoS) attacks, stress-testing unauthorized systems, or illegal activity is **strictly prohibited**.

The developers of Vessel are not responsible for any misuse or damages caused by this software.
