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
Running 10.025843334s test @ http://localhost:8000 [vessel-v0.0.1]
30 Connections

Summary:
  Requests:             21074 (2107 per second)
  Waiting:              [41.00%] Resolving DNS (4.10s), [0.00%] TLS Handshake (0.00s), [27.66%] Connecting (2.77s)
  Duration:             10.025843334s
  Latency:              max=55ms, avg=13.705419ms, p90=15ms, p95=16ms, p99=29ms
  Errors:               Total: 0: Timeout(0), Cancelled(0), Connection(0), Unknown(0)
  BytesReceived:        0.00MB/s
  BytesSent:            3.90MB/s

Response Codes Breakdown
        [200]: 21074
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
