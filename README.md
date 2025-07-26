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
vessel https://yourwebsite.com -c 50 -d 10s
```

- `-c 50` — 50 concurrent connections
- `-d 10s` — 10-second test duration

### Example with Headers and JSON Payload

```bash
vessel https://api.yourwebsite.com/data \
  -X POST \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -b '{"name": "vessel"}'
```

---

## 📊 Output Sample

```text
 _   _                    _
| | | |			 | |
| | | | ___  ___ ___  ___| |
| | | |/ ⚡\/ __/ __|/ ⚡\ |
\ \_/ /  __/\__ \__ \  __/ |
 \___/ \___||___/___/\___|_| https://github.com/symonk/vessel

Running test @ http://localhost:8000 [vessel-v0.0.1]
Workers: 20
Cores: 10

complete [⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡] 6.051627167s

Requests:	340949 (56825/second)
bytes:		Received(0.91MB) | Sent(0.00MB) | Total(0.91MB)
Latency:	max=4ms, avg=0.010004ms, p50=0ms, p90=0ms, p95=0ms, p99=0ms
Errored:	Total: 0: Timeout(0), Cancelled(0), Connection(0), Unknown(0)
Conns:		581
Waiting:	[12.13%] Resolving DNS (0.04s), [0.00%] TLS Handshake (0.00s), [7.27%] Connecting (0.02s) [25.22%] Getting Connections (0.08s)

Response Codes Breakdown
	[200]: 340951
```

---

## ⚙️ Options
| Flag            | Short | Type      | Default | Description                                                                                       |
| --------------- | ----- | --------- | ------- | ------------------------------------------------------------------------------------------------- |
| `--quiet`       | `-q`  | bool      | `false` | Suppresses all output                                                                             |
| `--max-rps`     | `-r`  | int       | `0`     | Rate limit requests per second (0 means no limit)                                                 |
| `--concurrency` | `-c`  | int       | `10`    | Number of concurrent requests                                                                     |
| `--duration`    | `-d`  | duration  | `0`     | Duration to send requests for (must be parsable by `time.ParseDuration`)                          |
| `--method`      | `-m`  | string    | `GET`   | HTTP method to perform (e.g., GET, POST)                                                          |
| `--timeout`     | `-t`  | duration  | `0`     | Per request timeout before terminating the request (must be parsable by `time.ParseDuration`)     |
| `--http2`       |       | bool      | `false` | Enable HTTP/2 support                                                                             |
| `--host`        |       | string    | `""`    | Set a custom Host header                                                                          |
| `--user-agent`  | `-u`  | string    | `""`    | Set a custom User-Agent header (always suffixed with the tool's user agent)                       |
| `--basic-auth`  | `-b`  | string    | `""`    | Colon-separated `user:pass` for Basic Auth header                                                 |
| `--headers`     | `-H`  | \[]string | `[]`    | Colon-separated `header:value` pairs for arbitrary HTTP headers (can be specified multiple times) |
| `--number`      | `-n`  | int64     | `50`    | Total number of requests to send (cannot be used together with `--duration`)                      |
| `--follow`      | `-f`  | bool      | `true`  | Automatically follow redirects                                                                    |
| `--show-cfg`    | `-s`  | bool      | `false` | Print the current configuration to stdout on startup                                              |
| `--insecure`    | `-i`  | bool      | `false` | Skip TLS server certificate and hostname verification (insecure, disables certificate validation) |
| `--max-conns`   |       | int       | 1024    | Maximum number of connections (per host) that should be used                                      |


---

## ⚠️ Disclaimer

Vessel is intended solely for **ethical performance testing** of web services you own or have explicit permission to test.  
Any use of this tool for denial-of-service (DoS) attacks, stress-testing unauthorized systems, or illegal activity is **strictly prohibited**.

The developers of Vessel are not responsible for any misuse or damages caused by this software.
