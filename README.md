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

- Blazing fast HTTP/1.1 and HTTP/2 benchmarking
- Real-time CLI metrics with latencies, throughput, and errors
- JSON and CSV output formats
- Custom headers, payloads, and methods
- Concurrency and rate controls

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
Running 10s test @ https://yourwebsite.com
50 connections

Summary:
  Requests:     12000
  Duration:     10.01s
  Latency:      avg=8.3ms max=240ms p95=15ms
  Errors:       2 timeouts, 3 connection resets
  Throughput:   1.1MB/s
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
