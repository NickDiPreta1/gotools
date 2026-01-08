# Go Tools — Development Roadmap

## Overview

Three CLI tools with shared infrastructure:

1. **blitz** — HTTP load tester
2. **streamrip** — Concurrent video/stream downloader
3. **netscan** — Network scanner and service discovery

Each tool builds on patterns from the previous, sharing common code in `gokit/`.

---

## Repository Strategy

**Development:** Monorepo with Go workspaces for easy iteration.

**Distribution:** Split into separate repos when stable:
- `github.com/NickDiPreta/blitz`
- `github.com/NickDiPreta/streamrip`
- `github.com/NickDiPreta/netscan`
- `github.com/NickDiPreta/gokit`

---

## Existing Foundation

From previous work:
- Worker pool pattern with graceful shutdown
- Channel-based result collection
- Context cancellation and timeouts
- Concurrent processing patterns

---

## Phase 0 — Project Setup & CLI Foundation

### Workspace Structure

```
gotools/
├── go.work                 # Go workspace file
├── gokit/                  # Shared library (its own module)
│   ├── go.mod              # module github.com/NickDiPreta/gokit
│   ├── cli/
│   │   ├── output.go       # Table, JSON, colored output
│   │   ├── progress.go     # Progress bars, spinners
│   │   └── config.go       # Config file loading
│   ├── stats/
│   │   ├── histogram.go    # Latency histograms
│   │   ├── counter.go      # Thread-safe counters
│   │   └── speed.go        # Download speed tracking
│   ├── pool/
│   │   └── pool.go         # Generic worker pool
│   └── netutil/
│       ├── resolver.go     # DNS resolution
│       └── timeout.go      # Timeout helpers
├── blitz/                  # Tool 1 (its own module)
│   ├── go.mod              # module github.com/NickDiPreta/blitz
│   └── main.go
├── streamrip/              # Tool 2 (its own module)
│   ├── go.mod              # module github.com/NickDiPreta/streamrip
│   ├── main.go
│   └── hls/                # HLS-specific code lives with the tool
│       ├── parser.go
│       ├── decrypt.go
│       └── segment.go
└── netscan/                # Tool 3 (its own module)
    ├── go.mod              # module github.com/NickDiPreta/netscan
    ├── main.go
    └── scanner/            # Scanner-specific code
        ├── tcp.go
        ├── banner.go
        └── fingerprint.go
```

### Initialize Workspace

```bash
go work init
go work use ./gokit ./blitz ./streamrip ./netscan
```

Each tool imports the shared library:
```go
import (
    "github.com/NickDiPreta/gokit/cli"
    "github.com/NickDiPreta/gokit/pool"
    "github.com/NickDiPreta/gokit/stats"
)
```

### Shared Infrastructure (gokit/)

| Package | Purpose |
|---------|---------|
| `cli/output.go` | Table formatter, JSON output, colored text |
| `cli/progress.go` | Progress bar, spinner, live updates |
| `stats/histogram.go` | Latency histograms, percentile calculation |
| `stats/counter.go` | Thread-safe atomic counters |
| `pool/pool.go` | Generic worker pool (migrate existing code) |

### Concepts

- Go workspaces (`go.work`)
- Multi-module repositories
- `flag` or cobra patterns
- ANSI escape codes for colors/cursor control
- `sync.Mutex` for thread-safe stats collection
- `io.Writer` interfaces for flexible output

---

## Phase 1 — blitz: HTTP Load Tester

### Stage 1.1: Basic Load Generation

**Goal:** `blitz -n 1000 -c 50 https://api.example.com/health`

#### Features
- Concurrent worker pool for HTTP requests
- Request counting (total, success, error by status code)
- Total duration and requests/second

#### Architecture

```
main.go
├── Parse flags (-n, -c, -url)
├── Create worker pool
├── Workers make HTTP requests
├── Collect results via channel
└── Print summary
```

#### Key Code Pattern

```go
type Result struct {
    StatusCode int
    Latency    time.Duration
    Error      error
    Timestamp  time.Time
}

func worker(ctx context.Context, client *http.Client, url string,
            jobs <-chan struct{}, results chan<- Result) {
    for range jobs {
        start := time.Now()
        resp, err := client.Do(req)
        results <- Result{
            Latency: time.Since(start),
            // ...
        }
    }
}
```

#### Concepts
- `http.Client` configuration (timeouts, keep-alive, connection pooling)
- Reusing connections for performance
- Basic load testing methodology

---

### Stage 1.2: Statistics and Latency Analysis

**Goal:** Percentile latencies (p50, p95, p99), not just averages.

#### Features
- Latency histogram with configurable buckets
- Percentile calculations
- Min/max/mean/stddev
- Status code distribution

#### Output Example

```
Summary:
  Total:        10.234s
  Requests:     10000
  Successful:   9847
  Failed:       153
  RPS:          977.12

Latency:
  Min:    2.31ms
  Mean:   45.67ms
  P50:    38.21ms
  P95:    112.34ms
  P99:    287.91ms
  Max:    1.23s

Status Codes:
  200: 9847 (98.47%)
  503: 142 (1.42%)
  Errors: 11 (0.11%)
```

#### Key Code: Histogram

```go
type Histogram struct {
    mu      sync.Mutex
    values  []time.Duration  // or use buckets for memory efficiency
}

func (h *Histogram) Add(d time.Duration)
func (h *Histogram) Percentile(p float64) time.Duration
func (h *Histogram) Mean() time.Duration
```

#### Concepts
- Lock-free vs mutex-based statistics
- Percentile algorithms (exact vs approximate)
- Memory-efficient data structures

---

### Stage 1.3: Live Progress and Rate Limiting

**Goal:** Real-time progress display and controlled request rate.

#### Features
- Live-updating terminal output (requests completed, current RPS, errors)
- Rate limiting: `-rate 100` = max 100 requests/second
- Duration-based runs: `-d 30s` instead of `-n`

#### Live Output Example

```
Running: 1547/10000 [=====>                    ] 15% | 487 req/s | Errors: 3
```

#### Rate Limiting Approaches

1. **Token bucket** — Classic algorithm, allows bursting
2. **Leaky bucket** — Smooth, constant rate
3. **`time.Ticker`** — Simple approach for steady rate

```go
func rateLimitedWorker(ctx context.Context, rate int, ...) {
    ticker := time.NewTicker(time.Second / time.Duration(rate))
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Send request
        case <-ctx.Done():
            return
        }
    }
}
```

#### Concepts
- Terminal manipulation (cursor movement, line clearing)
- Rate limiting algorithms
- `time.Ticker` patterns
- Combining multiple context cancellation sources

---

### Stage 1.4: Request Customization

**Goal:** Full control over HTTP requests.

#### Features
- Custom method: `-m POST`
- Custom headers: `-H "Authorization: Bearer xyz"`
- Request body: `-d '{"key": "value"}'` or `-d @file.json`
- Timeout per request: `-timeout 5s`

#### Config File Support

```yaml
# blitz.yaml
url: https://api.example.com/users
method: POST
headers:
  Authorization: Bearer ${TOKEN}
  Content-Type: application/json
body: |
  {"name": "test", "id": {{.Iteration}}}
concurrency: 50
requests: 10000
```

#### Concepts
- YAML parsing with `gopkg.in/yaml.v3`
- Environment variable interpolation
- Template execution for dynamic payloads
- File reading patterns (`@filename` convention)

---

### Stage 1.5: Advanced Features

**Goal:** Production-quality tool.

#### Features

1. **HTTP/2 Support**
   - Automatic with Go's `http.Client`
   - Compare HTTP/1.1 vs HTTP/2 performance

2. **Connection Tuning**
   ```go
   transport := &http.Transport{
       MaxIdleConns:        100,
       MaxIdleConnsPerHost: 100,
       IdleConnTimeout:     90 * time.Second,
   }
   ```

3. **Request Sequences** (optional)
   - Run multiple requests in sequence (login → get token → use token)
   - Variable capture from responses

4. **Output Formats**
   - JSON for programmatic consumption
   - CSV for spreadsheet analysis
   - Compare mode (run twice, diff results)

5. **Distributed Mode** (stretch goal)
   - Coordinator sends work to multiple agents
   - Aggregate results from multiple machines

#### Concepts
- HTTP/2 internals
- Connection pool tuning
- JSON streaming output
- Client/server coordination (for distributed mode)

---

### Stage 1.6: Polish and Comparison

#### Tasks
- Comprehensive `--help` output
- README with examples
- Compare results against `hey`, `wrk`, `vegeta`
- Fix any edge cases found during comparison
- Add benchmarks for the tool itself

---

## Phase 2 — streamrip: Concurrent Video Downloader

### How Streaming Video Works

Most modern video streaming uses **HLS (HTTP Live Streaming)** or **DASH**:

```
┌─────────────────────────────────────────────────────┐
│  manifest.m3u8 (playlist file)                      │
│  ─────────────────────────────────────              │
│  #EXTM3U                                            │
│  #EXT-X-TARGETDURATION:10                           │
│  #EXT-X-KEY:METHOD=AES-128,URI="key.bin"  (optional)│
│  segment_001.ts   ← 10 seconds of video             │
│  segment_002.ts   ← 10 seconds of video             │
│  segment_003.ts   ← ...                             │
│  ...                                                │
│  segment_150.ts   ← last segment                    │
└─────────────────────────────────────────────────────┘
```

**The speed trick:** Download segments in parallel instead of sequentially:

```
Sequential:     [seg1]──[seg2]──[seg3]──[seg4]──...  (slow, ~1x bandwidth)

Concurrent:     [seg1]
                [seg2]
                [seg3]     ← saturates available bandwidth
                [seg4]
                ...
```

Then merge all segments into a single file.

---

### Stage 2.1: M3U8 Parser and Sequential Download

**Goal:** `streamrip https://example.com/video/master.m3u8 -o video.ts`

#### Features
- Parse HLS manifest (.m3u8) files
- Handle master playlists (multiple quality levels) vs media playlists
- Download segments sequentially
- Concatenate into single .ts file

#### M3U8 Format Basics

```
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
https://cdn.example.com/seg0.ts
#EXTINF:10.0,
https://cdn.example.com/seg1.ts
#EXTINF:8.5,
https://cdn.example.com/seg2.ts
#EXT-X-ENDLIST
```

#### Key Code: Parser

```go
type Playlist struct {
    Version        int
    TargetDuration float64
    Segments       []Segment
    IsVOD          bool  // has #EXT-X-ENDLIST
    Key            *EncryptionKey  // optional
}

type Segment struct {
    Duration float64
    URI      string
    Sequence int
}

func ParseM3U8(r io.Reader, baseURL string) (*Playlist, error) {
    // Line-by-line parsing
    // Handle relative vs absolute URLs
}
```

#### Concepts
- Text parsing (line-by-line, state machine)
- URL resolution (relative paths)
- File concatenation
- Basic streaming download

---

### Stage 2.2: Concurrent Segment Downloading

**Goal:** Download segments in parallel with worker pool.

#### Features
- Worker pool for concurrent downloads
- Configurable concurrency: `-c 10`
- Download speed display
- Resume interrupted downloads

#### Architecture

```
                    ┌─────────────────┐
                    │  Parse manifest │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   Job Queue     │ (segment URLs)
                    └────────┬────────┘
           ┌─────────────────┼─────────────────┐
           ▼                 ▼                 ▼
      ┌─────────┐       ┌─────────┐       ┌─────────┐
      │Worker 1 │       │Worker 2 │       │Worker 3 │
      └────┬────┘       └────┬────┘       └────┬────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             ▼
                    ┌─────────────────┐
                    │  Results Chan   │
                    └────────┬────────┘
                             ▼
                    ┌─────────────────┐
                    │  Write to disk  │ (ordered by sequence)
                    └─────────────────┘
```

#### Key Challenge: Ordered Output

Segments must be written in order, but complete out of order:

```go
type SegmentResult struct {
    Sequence int
    Data     []byte
    Error    error
}

// Collector writes segments in order
func collector(results <-chan SegmentResult, output io.Writer) error {
    pending := make(map[int][]byte)
    nextSeq := 0

    for result := range results {
        pending[result.Sequence] = result.Data

        // Write any segments that are ready in order
        for {
            data, ok := pending[nextSeq]
            if !ok {
                break
            }
            output.Write(data)
            delete(pending, nextSeq)
            nextSeq++
        }
    }
    return nil
}
```

#### Concepts
- Reusing worker pool pattern from blitz
- Out-of-order completion with in-order output
- Streaming writes to disk
- Download resume patterns

---

### Stage 2.3: Progress and Speed Display

**Goal:** Rich progress display with download speed.

#### Features
- Overall progress bar
- Download speed (MB/s)
- ETA calculation
- Per-segment status (optional verbose mode)

#### Output Example

```
Downloading: video.ts
Quality: 1080p (1920x1080)
Segments: 47/156 [========>                ] 30%
Speed: 12.4 MB/s | ETA: 45s | Downloaded: 234.5 MB
```

#### Speed Calculation

```go
type SpeedTracker struct {
    mu           sync.Mutex
    samples      []speedSample
    windowSize   time.Duration
}

type speedSample struct {
    bytes     int64
    timestamp time.Time
}

func (s *SpeedTracker) Add(bytes int64) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.samples = append(s.samples, speedSample{bytes, time.Now()})
    s.pruneOld()
}

func (s *SpeedTracker) BytesPerSecond() float64 {
    // Calculate from samples in window
}
```

#### Concepts
- Rolling averages for speed calculation
- ETA estimation
- Terminal UI updates (reusing from blitz)

---

### Stage 2.4: Quality Selection and Master Playlists

**Goal:** Handle multi-quality streams, let user choose quality.

#### Features
- Parse master playlists with multiple qualities
- List available qualities: `streamrip --list-quality URL`
- Select quality: `-q 1080p` or `-q best` or `-q worst`
- Auto-select based on bandwidth (optional)

#### Master Playlist Format

```
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360
360p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1400000,RESOLUTION=842x480
480p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720
720p/playlist.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080
1080p/playlist.m3u8
```

#### Key Code

```go
type MasterPlaylist struct {
    Variants []Variant
}

type Variant struct {
    Bandwidth  int
    Resolution string
    URI        string
    Codecs     string
}

func (m *MasterPlaylist) SelectQuality(preference string) (*Variant, error) {
    switch preference {
    case "best":
        return m.highestBandwidth(), nil
    case "worst":
        return m.lowestBandwidth(), nil
    default:
        return m.byResolution(preference)
    }
}
```

#### Concepts
- Two-level manifest parsing
- User preference handling
- Bandwidth/quality trade-offs

---

### Stage 2.5: Encrypted Streams (AES-128)

**Goal:** Handle HLS streams encrypted with AES-128.

#### How HLS Encryption Works

```
#EXT-X-KEY:METHOD=AES-128,URI="https://example.com/key.bin",IV=0x1234...
segment_001.ts  ← encrypted with key
segment_002.ts  ← encrypted with key
```

- Key is fetched from the URI
- IV (initialization vector) may be explicit or derived from sequence number
- Each segment is AES-128-CBC encrypted

#### Features
- Detect encrypted playlists
- Fetch decryption key
- Decrypt segments during download
- Handle rotating keys (key changes mid-stream)

#### Key Code

```go
func decryptSegment(data []byte, key []byte, iv []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    if len(data)%aes.BlockSize != 0 {
        return nil, errors.New("ciphertext not multiple of block size")
    }

    mode := cipher.NewCBCDecrypter(block, iv)
    mode.CryptBlocks(data, data)

    // Remove PKCS7 padding
    return unpad(data), nil
}
```

#### Concepts
- `crypto/aes` and `crypto/cipher` packages
- AES-CBC mode
- PKCS7 padding
- IV derivation

---

### Stage 2.6: Auto-Detection and Browser Integration

**Goal:** Extract m3u8 URLs from webpage HTML.

#### Features
- Provide webpage URL, tool finds m3u8: `streamrip https://example.com/watch/123`
- Parse HTML for video sources
- Check common patterns (video tags, JavaScript, network requests)
- Cookie/header support for authenticated streams

#### Detection Strategies

1. **HTML Parsing**
   ```go
   // Look for <video> tags, <source> tags
   // Look for known player JavaScript patterns
   ```

2. **Common URL Patterns**
   ```go
   patterns := []string{
       `https?://[^"'\s]+\.m3u8[^"'\s]*`,
       `https?://[^"'\s]+/manifest[^"'\s]*`,
   }
   ```

3. **JavaScript Parsing** (basic)
   - Look for m3u8 URLs in inline scripts
   - Look for API calls that return manifest URLs

#### Features
- Cookie jar support: `--cookies cookies.txt`
- Custom headers: `-H "Referer: https://example.com"`
- User-agent spoofing

#### Concepts
- HTML parsing (`golang.org/x/net/html`)
- Regex for URL extraction
- Cookie handling in Go
- HTTP header manipulation

---

### Stage 2.7: Output Formats and FFmpeg Integration

**Goal:** Convert to common formats, not just .ts files.

#### Features
- Direct .ts concatenation (fast, no re-encoding)
- FFmpeg remux to .mp4: `-o video.mp4`
- Audio-only extraction: `--audio-only`
- Subtitle download (if available)

#### FFmpeg Integration

```go
func remuxToMP4(inputPath, outputPath string) error {
    cmd := exec.Command("ffmpeg",
        "-i", inputPath,
        "-c", "copy",  // no re-encoding
        "-movflags", "+faststart",
        outputPath,
    )
    return cmd.Run()
}
```

#### Concepts
- `os/exec` for running external commands
- Container formats (TS vs MP4)
- Streaming media concepts

---

### Stage 2.8: Polish and Edge Cases

#### Tasks
- Test with various streaming sites
- Handle redirects and CDN quirks
- Retry failed segments
- Timeout handling
- Comprehensive error messages
- README with examples

#### Edge Cases to Handle
- Segments on different domains
- Relative vs absolute URLs
- Live streams (no #EXT-X-ENDLIST)
- Discontinuities (#EXT-X-DISCONTINUITY)
- Byte-range requests

---

## Phase 3 — netscan: Network Scanner

### Stage 3.1: TCP Port Scanner

**Goal:** `netscan -p 1-1000 192.168.1.1`

#### Features
- TCP connect scanning
- Port range parsing (`-p 22,80,443` or `-p 1-1000`)
- Concurrent scanning with worker pool
- Timeout per connection attempt

#### Key Code

```go
type ScanResult struct {
    Port    int
    State   string  // "open", "closed", "filtered"
    Latency time.Duration
}

func scanPort(ctx context.Context, host string, port int, timeout time.Duration) ScanResult {
    address := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", address, timeout)
    if err != nil {
        // Distinguish between closed and filtered
        return ScanResult{Port: port, State: classifyError(err)}
    }
    conn.Close()
    return ScanResult{Port: port, State: "open", Latency: ...}
}
```

#### Concepts
- `net.Dial` and `net.DialTimeout`
- Error classification (connection refused vs timeout)
- Efficient port range iteration

---

### Stage 3.2: Host Discovery

**Goal:** `netscan discover 192.168.1.0/24`

#### Features
- CIDR range parsing
- Multiple discovery methods:
  - TCP ping (connect to common port)
  - ICMP ping (requires elevated privileges)
  - ARP scan (local network only)

#### CIDR Parsing

```go
func hostsInCIDR(cidr string) ([]net.IP, error) {
    ip, ipnet, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, err
    }
    var hosts []net.IP
    for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
        hosts = append(hosts, copyIP(ip))
    }
    return hosts, nil
}
```

#### Concepts
- IP address manipulation
- CIDR notation and subnetting
- `golang.org/x/net/icmp` for ICMP (optional)
- Raw sockets (optional, requires root)

---

### Stage 3.3: Service Detection

**Goal:** Identify what's running on open ports.

#### Features
- Banner grabbing (read initial response from service)
- Protocol detection (HTTP, SSH, FTP, SMTP, etc.)
- Version extraction where possible

#### Banner Grabbing

```go
func grabBanner(host string, port int, timeout time.Duration) (string, error) {
    conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
    if err != nil {
        return "", err
    }
    defer conn.Close()

    conn.SetReadDeadline(time.Now().Add(timeout))

    // Some services send banner immediately (SSH, FTP)
    // Others need a probe (HTTP)
    buf := make([]byte, 1024)
    n, _ := conn.Read(buf)
    if n > 0 {
        return string(buf[:n]), nil
    }

    // Send HTTP probe
    conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
    n, _ = conn.Read(buf)
    return string(buf[:n]), nil
}
```

#### Service Fingerprints

```go
var servicePatterns = map[string]*regexp.Regexp{
    "ssh":   regexp.MustCompile(`^SSH-`),
    "http":  regexp.MustCompile(`^HTTP/`),
    "ftp":   regexp.MustCompile(`^220[ -]`),
    "smtp":  regexp.MustCompile(`^220[ -].*SMTP`),
    "mysql": regexp.MustCompile(`^.\x00\x00\x00\x0a`),
}
```

#### Concepts
- Raw TCP reading/writing
- Protocol basics (how services identify themselves)
- Regex for pattern matching
- Handling binary protocols

---

### Stage 3.4: Output and Reporting

**Goal:** Professional-quality output.

#### Output Formats

**Table (default):**
```
HOST            PORT   STATE   SERVICE   VERSION
192.168.1.1     22     open    ssh       OpenSSH 8.9
192.168.1.1     80     open    http      nginx/1.18.0
192.168.1.1     443    open    https     nginx/1.18.0
192.168.1.1     3306   open    mysql     MySQL 8.0.32
```

**JSON:**
```json
{
  "host": "192.168.1.1",
  "ports": [
    {"port": 22, "state": "open", "service": "ssh", "version": "OpenSSH 8.9"}
  ],
  "scan_time": "2.34s"
}
```

**Grepable:**
```
192.168.1.1:22:open:ssh:OpenSSH 8.9
192.168.1.1:80:open:http:nginx/1.18.0
```

#### Features
- Save/load scan results
- Diff two scans (find new/removed ports)
- Summary statistics

---

### Stage 3.5: Advanced Scanning

#### Features

1. **SYN Scanning** (half-open)
   - Faster than connect scanning
   - Requires raw sockets (root/admin)
   - Uses `golang.org/x/net/ipv4`

2. **UDP Scanning**
   - Send UDP probes
   - Detect ICMP port unreachable vs no response

3. **Timing Profiles**
   - `-T0` (paranoid) to `-T5` (insane)
   - Control concurrency, timeouts, retry behavior

4. **OS Detection** (stretch goal)
   - TCP/IP fingerprinting
   - TTL analysis
   - Window size analysis

#### Concepts
- Raw socket programming
- TCP/IP internals (SYN, SYN-ACK, RST)
- UDP characteristics
- OS fingerprinting techniques

---

### Stage 3.6: Polish and Comparison

#### Tasks
- Test against nmap results for accuracy
- Handle edge cases (firewalls, rate limiting)
- Comprehensive documentation
- Add common port presets (`--top-ports 100`)

---

## Timeline

| Phase | Duration | Cumulative |
|-------|----------|------------|
| Phase 0: CLI Foundation | 1 week | 1 week |
| Phase 1.1-1.2: Basic blitz | 2 weeks | 3 weeks |
| Phase 1.3-1.4: Advanced blitz | 2 weeks | 5 weeks |
| Phase 1.5-1.6: Polish blitz | 2 weeks | 7 weeks |
| Phase 2.1-2.2: Basic streamrip | 2 weeks | 9 weeks |
| Phase 2.3-2.4: Quality & progress | 2 weeks | 11 weeks |
| Phase 2.5-2.6: Encryption & detection | 2 weeks | 13 weeks |
| Phase 2.7-2.8: FFmpeg & polish | 1.5 weeks | 14.5 weeks |
| Phase 3.1-3.2: Basic netscan | 2 weeks | 16.5 weeks |
| Phase 3.3-3.4: Service detection | 2 weeks | 18.5 weeks |
| Phase 3.5-3.6: Advanced & polish | 2 weeks | 20.5 weeks |

**Total: ~20 weeks**

---

## Go Concepts by Tool

| Concept | Tool |
|---------|------|
| Worker pools | All |
| Channels | All |
| Context | All |
| Interfaces | All |
| Generics | gokit/pool |
| sync.Mutex/atomic | stats |
| net package | netscan |
| http package | blitz, streamrip |
| crypto/aes | streamrip |
| Text parsing | streamrip |
| File I/O | streamrip |
| os/exec | streamrip (FFmpeg) |

---

## Tool Progression

```
blitz
├── HTTP client tuning
├── Worker pool
├── Progress bar
├── Statistics
└── Rate limiting
        │
        ▼
streamrip
├── Reuses HTTP client patterns
├── Reuses worker pool
├── Reuses progress bar
├── Adds: parsing, file I/O, crypto
└── Adds: ordered output from unordered completion
        │
        ▼
netscan
├── Reuses worker pool
├── Reuses progress bar
├── Reuses output formatting
├── Adds: net package (raw TCP)
└── Adds: protocol detection
```

---

## Splitting Into Separate Repos

When tools are stable:

1. Publish gokit first, tag v0.1.0
2. Update each tool's go.mod to require published gokit
3. Create separate repos for each tool
4. Users install via: `go install github.com/NickDiPreta/blitz@latest`
