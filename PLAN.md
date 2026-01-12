# Go Tools — Development Roadmap

## Overview

Four projects with shared infrastructure:

1. **blitz** — HTTP load tester
2. **streamrip** — Concurrent video/stream downloader
3. **netscan** — Network scanner and service discovery
4. **audiotrack** — Audiobook library and progress tracker

Each tool builds on patterns from the previous, sharing common code in `gokit/`.

---

## Repository Strategy

**Development:** Monorepo with Go workspaces for easy iteration.

**Distribution:** Split into separate repos when stable:
- `github.com/NickDiPreta/blitz`
- `github.com/NickDiPreta/streamrip`
- `github.com/NickDiPreta/netscan`
- `github.com/NickDiPreta/audiotrack`
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

### Overview

Phase 0 establishes the foundation for all four tools. This phase focuses on creating a well-organized workspace, implementing reusable shared libraries, and ensuring the CLI infrastructure is solid before building any tool-specific features.

---

### Stage 0.1: Go Workspace Setup

**Goal:** Initialize the monorepo structure with Go workspaces.

#### Tasks

1. **Create directory structure**
   ```bash
   mkdir -p gokit/{cli,stats,pool,netutil}
   mkdir -p blitz streamrip netscan audiotrack
   ```

2. **Initialize Go modules**
   ```bash
   # Shared library
   cd gokit && go mod init github.com/NickDiPreta/gokit

   # Each tool as separate module
   cd ../blitz && go mod init github.com/NickDiPreta/blitz
   cd ../streamrip && go mod init github.com/NickDiPreta/streamrip
   cd ../netscan && go mod init github.com/NickDiPreta/netscan
   cd ../audiotrack && go mod init github.com/NickDiPreta/audiotrack
   ```

3. **Create Go workspace file**
   ```bash
   cd .. # back to gotools root
   go work init
   go work use ./gokit ./blitz ./streamrip ./netscan ./audiotrack
   ```

4. **Verify workspace**
   ```bash
   go work sync
   cat go.work  # Should list all modules
   ```

#### Deliverables
- `go.work` file at repository root
- `go.mod` file in each module directory
- All modules recognized by `go work sync`

---

### Stage 0.2: CLI Output Package (gokit/cli/output.go)

**Goal:** Create reusable colored output and table formatting utilities.

#### Tasks

1. **Define color constants and helpers**
   ```go
   // ANSI color codes
   const (
       Reset   = "\033[0m"
       Red     = "\033[31m"
       Green   = "\033[32m"
       Yellow  = "\033[33m"
       Blue    = "\033[34m"
       Magenta = "\033[35m"
       Cyan    = "\033[36m"
       Bold    = "\033[1m"
       Dim     = "\033[2m"
   )

   func Colorize(color, text string) string
   func Success(text string) string  // Green
   func Error(text string) string    // Red
   func Warning(text string) string  // Yellow
   func Info(text string) string     // Cyan
   ```

2. **Implement table formatter**
   ```go
   type Table struct {
       headers []string
       rows    [][]string
       writer  io.Writer
   }

   func NewTable(headers ...string) *Table
   func (t *Table) AddRow(values ...string)
   func (t *Table) Render()
   func (t *Table) SetWriter(w io.Writer)
   ```
   - Auto-calculate column widths
   - Support alignment (left, right, center)
   - Handle long strings (truncate or wrap)

3. **Implement JSON output mode**
   ```go
   type OutputFormat int
   const (
       FormatTable OutputFormat = iota
       FormatJSON
       FormatCSV
   )

   func SetOutputFormat(format OutputFormat)
   func Print(v interface{}) error  // Auto-formats based on current mode
   ```

4. **Add terminal detection**
   ```go
   func IsTerminal() bool           // Check if stdout is a TTY
   func TerminalWidth() int         // Get terminal width for formatting
   func DisableColors()             // For non-TTY or --no-color flag
   ```

#### Tests to Write
- Color output with/without TTY
- Table rendering with various column widths
- JSON output formatting
- Terminal width detection

#### Deliverables
- `gokit/cli/output.go` with color, table, and JSON formatting
- `gokit/cli/output_test.go` with unit tests

---

### Stage 0.3: Progress Display Package (gokit/cli/progress.go)

**Goal:** Create live-updating progress bars and spinners.

#### Tasks

1. **Implement progress bar**
   ```go
   type ProgressBar struct {
       total     int64
       current   int64
       width     int
       startTime time.Time
       mu        sync.Mutex
   }

   func NewProgressBar(total int64) *ProgressBar
   func (p *ProgressBar) Increment(n int64)
   func (p *ProgressBar) Set(n int64)
   func (p *ProgressBar) Render() string
   func (p *ProgressBar) Done()
   ```
   - Display format: `[=====>                    ] 30% | 150/500 | ETA: 45s`
   - Calculate ETA from elapsed time and progress
   - Support custom width

2. **Implement spinner for indeterminate progress**
   ```go
   type Spinner struct {
       frames  []string
       current int
       message string
       done    chan struct{}
   }

   func NewSpinner(message string) *Spinner
   func (s *Spinner) Start()
   func (s *Spinner) Stop()
   func (s *Spinner) SetMessage(msg string)
   ```
   - Default frames: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` or `|/-\`
   - Auto-refresh at configurable interval

3. **Implement live stats display**
   ```go
   type LiveDisplay struct {
       lines    []string
       mu       sync.Mutex
       ticker   *time.Ticker
   }

   func NewLiveDisplay() *LiveDisplay
   func (l *LiveDisplay) SetLine(index int, content string)
   func (l *LiveDisplay) Start(refreshRate time.Duration)
   func (l *LiveDisplay) Stop()
   ```
   - Clear and redraw on each update
   - Handle terminal resize
   - Use ANSI escape codes for cursor movement

4. **Terminal control utilities**
   ```go
   func ClearLine()
   func MoveCursorUp(n int)
   func MoveCursorDown(n int)
   func HideCursor()
   func ShowCursor()
   ```

#### Tests to Write
- Progress bar percentage calculation
- ETA estimation accuracy
- Spinner frame cycling
- Live display line updates

#### Deliverables
- `gokit/cli/progress.go` with progress bar, spinner, and live display
- `gokit/cli/progress_test.go` with unit tests

---

### Stage 0.4: Configuration Package (gokit/cli/config.go)

**Goal:** Create config file loading and environment variable handling.

#### Tasks

1. **Implement config file loading**
   ```go
   type Config struct {
       path   string
       values map[string]interface{}
   }

   func LoadConfig(path string) (*Config, error)
   func (c *Config) Get(key string) interface{}
   func (c *Config) GetString(key string) string
   func (c *Config) GetInt(key string) int
   func (c *Config) GetDuration(key string) time.Duration
   func (c *Config) GetBool(key string) bool
   ```
   - Support YAML format (using `gopkg.in/yaml.v3`)
   - Support JSON format
   - Auto-detect format from extension

2. **Implement environment variable interpolation**
   ```go
   func ExpandEnv(value string) string
   // Expands ${VAR} and $VAR patterns
   // Example: "Bearer ${TOKEN}" → "Bearer abc123"
   ```

3. **Implement config file discovery**
   ```go
   func FindConfigFile(name string) (string, error)
   // Search order:
   // 1. Current directory: ./name.yaml, ./name.yml, ./name.json
   // 2. XDG config: ~/.config/toolname/config.yaml
   // 3. Home directory: ~/.toolname.yaml
   ```

4. **Add flag integration helpers**
   ```go
   func BindFlag(config *Config, key string, flagValue interface{})
   // Allows config file values to provide defaults for CLI flags
   ```

#### Tests to Write
- YAML config parsing
- JSON config parsing
- Environment variable expansion
- Config file discovery in various locations
- Flag binding

#### Deliverables
- `gokit/cli/config.go` with config loading and env expansion
- `gokit/cli/config_test.go` with unit tests
- Add dependency: `gopkg.in/yaml.v3`

---

### Stage 0.5: Statistics Package (gokit/stats/)

**Goal:** Create thread-safe statistics collection utilities.

#### Tasks

1. **Implement histogram (gokit/stats/histogram.go)**
   ```go
   type Histogram struct {
       mu     sync.Mutex
       values []time.Duration
       sorted bool
   }

   func NewHistogram() *Histogram
   func (h *Histogram) Add(d time.Duration)
   func (h *Histogram) Percentile(p float64) time.Duration  // p50, p95, p99
   func (h *Histogram) Min() time.Duration
   func (h *Histogram) Max() time.Duration
   func (h *Histogram) Mean() time.Duration
   func (h *Histogram) StdDev() time.Duration
   func (h *Histogram) Count() int64
   ```
   - Sort values lazily for percentile calculations
   - Consider memory-efficient bucketed histogram for high-volume data

2. **Implement atomic counters (gokit/stats/counter.go)**
   ```go
   type Counter struct {
       value int64  // Use atomic operations
   }

   func NewCounter() *Counter
   func (c *Counter) Inc()
   func (c *Counter) Add(n int64)
   func (c *Counter) Value() int64
   func (c *Counter) Reset() int64  // Returns value before reset

   // CounterMap for tracking multiple named counters
   type CounterMap struct {
       mu       sync.RWMutex
       counters map[string]*Counter
   }

   func NewCounterMap() *CounterMap
   func (m *CounterMap) Inc(key string)
   func (m *CounterMap) Get(key string) int64
   func (m *CounterMap) All() map[string]int64
   ```

3. **Implement speed tracker (gokit/stats/speed.go)**
   ```go
   type SpeedTracker struct {
       mu         sync.Mutex
       samples    []speedSample
       windowSize time.Duration  // Rolling window for calculation
   }

   type speedSample struct {
       bytes     int64
       timestamp time.Time
   }

   func NewSpeedTracker(windowSize time.Duration) *SpeedTracker
   func (s *SpeedTracker) Add(bytes int64)
   func (s *SpeedTracker) BytesPerSecond() float64
   func (s *SpeedTracker) Format() string  // "12.4 MB/s"
   ```
   - Prune old samples outside window
   - Handle edge cases (no samples, single sample)

#### Tests to Write
- Histogram percentile accuracy
- Counter thread safety (concurrent increments)
- Speed tracker calculation
- Memory efficiency for large datasets

#### Deliverables
- `gokit/stats/histogram.go`
- `gokit/stats/counter.go`
- `gokit/stats/speed.go`
- `gokit/stats/*_test.go` for each

---

### Stage 0.6: Worker Pool Package (gokit/pool/)

**Goal:** Create a generic, reusable worker pool with graceful shutdown.

#### Tasks

1. **Implement generic worker pool**
   ```go
   type Pool[T any, R any] struct {
       workers    int
       jobs       chan T
       results    chan R
       workerFunc func(context.Context, T) R
       wg         sync.WaitGroup
   }

   func New[T any, R any](workers int, fn func(context.Context, T) R) *Pool[T, R]
   func (p *Pool[T, R]) Start(ctx context.Context)
   func (p *Pool[T, R]) Submit(job T)
   func (p *Pool[T, R]) Results() <-chan R
   func (p *Pool[T, R]) Wait()
   func (p *Pool[T, R]) Close()
   ```

2. **Add graceful shutdown**
   - Respect context cancellation
   - Drain jobs queue on shutdown
   - Wait for in-flight work to complete

3. **Add pool options**
   ```go
   type PoolOption func(*poolConfig)

   func WithJobBuffer(size int) PoolOption
   func WithResultBuffer(size int) PoolOption
   func WithErrorHandler(fn func(error)) PoolOption
   ```

4. **Migrate existing blitz worker pool**
   - Review current implementation in blitz
   - Extract common patterns
   - Ensure backward compatibility

#### Tests to Write
- Basic job submission and result collection
- Context cancellation handling
- Graceful shutdown with pending jobs
- Concurrent safety

#### Deliverables
- `gokit/pool/pool.go`
- `gokit/pool/pool_test.go`

---

### Stage 0.7: Network Utilities Package (gokit/netutil/)

**Goal:** Create reusable network helper functions.

#### Tasks

1. **Implement DNS resolver utilities (gokit/netutil/resolver.go)**
   ```go
   func ResolveHost(ctx context.Context, host string) ([]net.IP, error)
   func ResolveHostPreferIPv4(ctx context.Context, host string) (net.IP, error)
   func IsIPv4(ip net.IP) bool
   func IsIPv6(ip net.IP) bool
   ```

2. **Implement timeout helpers (gokit/netutil/timeout.go)**
   ```go
   type TimeoutConfig struct {
       Connect time.Duration
       Read    time.Duration
       Write   time.Duration
       Total   time.Duration
   }

   func DefaultTimeoutConfig() TimeoutConfig
   func (tc TimeoutConfig) DialContext(ctx context.Context, network, addr string) (net.Conn, error)
   ```

3. **Implement URL utilities**
   ```go
   func ResolveURL(base, ref string) (string, error)  // Handle relative URLs
   func ExtractHost(urlStr string) (string, error)
   func NormalizeURL(urlStr string) (string, error)
   ```

#### Tests to Write
- DNS resolution (may need mocking)
- URL resolution with various base/ref combinations
- Timeout behavior

#### Deliverables
- `gokit/netutil/resolver.go`
- `gokit/netutil/timeout.go`
- `gokit/netutil/*_test.go`

---

### Stage 0.8: Blitz Module Skeleton

**Goal:** Set up the blitz module to use gokit packages.

#### Tasks

1. **Create main.go with flag parsing**
   ```go
   package main

   import (
       "flag"
       "fmt"
       "os"

       "github.com/NickDiPreta/gokit/cli"
   )

   var (
       numRequests = flag.Int("n", 100, "Number of requests")
       concurrency = flag.Int("c", 10, "Concurrent workers")
       url         = flag.String("url", "", "Target URL")
   )

   func main() {
       flag.Parse()
       if *url == "" {
           fmt.Fprintln(os.Stderr, cli.Error("URL is required"))
           os.Exit(1)
       }
       // Placeholder for load test logic
   }
   ```

2. **Verify gokit imports work**
   ```bash
   cd blitz
   go build .  # Should compile successfully
   ```

3. **Create placeholder structure for later phases**
   ```
   blitz/
   ├── go.mod
   ├── main.go
   └── internal/
       └── runner/
           └── runner.go  # Placeholder for load test runner
   ```

#### Deliverables
- Working `blitz` binary that uses gokit packages
- Verified import paths work across workspace

---

### Stage 0.9: Testing and Documentation

**Goal:** Ensure all packages have tests and basic documentation.

#### Tasks

1. **Run all tests**
   ```bash
   go test ./gokit/...
   ```

2. **Add package documentation**
   - Each package should have a `doc.go` with package-level documentation
   - Key exported types and functions should have doc comments

3. **Create example usage**
   ```go
   // gokit/cli/example_test.go
   func ExampleTable() {
       t := cli.NewTable("Name", "Status", "Duration")
       t.AddRow("Request 1", "OK", "45ms")
       t.AddRow("Request 2", "Failed", "1.2s")
       t.Render()
   }
   ```

4. **Verify cross-module imports**
   ```bash
   # From blitz directory
   go build .

   # From root
   go work sync
   go build ./...
   ```

#### Deliverables
- All tests passing
- Package documentation in place
- Example usage for key packages

---

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
| `cli/config.go` | Config file loading, env expansion |
| `stats/histogram.go` | Latency histograms, percentile calculation |
| `stats/counter.go` | Thread-safe atomic counters |
| `stats/speed.go` | Download speed tracking |
| `pool/pool.go` | Generic worker pool (migrate existing code) |
| `netutil/resolver.go` | DNS resolution helpers |
| `netutil/timeout.go` | Timeout configuration helpers |

### Phase 0 Checklist

- [ ] Stage 0.1: Go workspace initialized with all modules
- [ ] Stage 0.2: CLI output package with colors, tables, JSON
- [ ] Stage 0.3: Progress display with bar, spinner, live updates
- [ ] Stage 0.4: Config loading with YAML/JSON and env expansion
- [ ] Stage 0.5: Stats package with histogram, counters, speed tracker
- [ ] Stage 0.6: Generic worker pool with graceful shutdown
- [ ] Stage 0.7: Network utilities for DNS and timeouts
- [ ] Stage 0.8: Blitz module skeleton using gokit
- [ ] Stage 0.9: All tests passing, documentation complete

### Concepts Covered

- Go workspaces (`go.work`)
- Multi-module repositories
- `flag` package for CLI arguments
- ANSI escape codes for colors/cursor control
- `sync.Mutex` and `sync/atomic` for thread safety
- `io.Writer` interfaces for flexible output
- Go generics for reusable worker pool
- Context cancellation patterns
- YAML parsing with `gopkg.in/yaml.v3`

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

## Phase 4 — audiotrack: Audiobook Library & Progress Tracker

A personal audiobook management system that stores your audiobook files and remembers exactly where you left off.

### Stage 4.1: Core Data Model and Storage

**Goal:** Define the data structures and storage backend for audiobooks and progress.

#### Data Model

```go
type Audiobook struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Author      string    `json:"author"`
    Narrator    string    `json:"narrator"`
    Duration    time.Duration `json:"duration"`
    FilePath    string    `json:"file_path"`
    CoverPath   string    `json:"cover_path"`
    AddedAt     time.Time `json:"added_at"`
    Chapters    []Chapter `json:"chapters"`
}

type Chapter struct {
    Title    string        `json:"title"`
    StartPos time.Duration `json:"start_pos"`
}

type Progress struct {
    BookID      string        `json:"book_id"`
    Position    time.Duration `json:"position"`
    UpdatedAt   time.Time     `json:"updated_at"`
    Completed   bool          `json:"completed"`
    PlaybackSpeed float64     `json:"playback_speed"`
}
```

#### Storage Options

1. **SQLite** — Simple, file-based, good for single-user
2. **BoltDB** — Pure Go, embedded key-value store
3. **JSON files** — Simplest, human-readable

#### Features
- Initialize library database
- CRUD operations for audiobooks
- Store/retrieve progress for each book
- Metadata extraction from audio files

#### Concepts
- Database design patterns
- `database/sql` or embedded databases
- Audio metadata parsing (ID3 tags, etc.)

---

### Stage 4.2: File Upload and Library Management

**Goal:** `audiotrack add ~/audiobooks/book.mp3` or `audiotrack import ~/audiobooks/`

#### Features
- Add individual audiobook files
- Bulk import from directory
- Supported formats: MP3, M4A, M4B, FLAC, OGG
- Automatic metadata extraction (title, author, duration, cover art)
- Organize files into library structure (optional)

#### CLI Commands

```bash
# Add a single audiobook
audiotrack add book.mp3

# Import all audiobooks from a directory
audiotrack import ~/Downloads/audiobooks/

# List all audiobooks in library
audiotrack list

# Show details for a book
audiotrack info <book-id>

# Remove from library
audiotrack remove <book-id>
```

#### Metadata Extraction

```go
func extractMetadata(filePath string) (*Audiobook, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    // Use github.com/dhowden/tag or similar
    m, err := tag.ReadFrom(f)
    if err != nil {
        return nil, err
    }

    return &Audiobook{
        Title:  m.Title(),
        Author: m.Artist(),
        // ...
    }, nil
}
```

#### Concepts
- File handling and organization
- Audio metadata libraries (`github.com/dhowden/tag`)
- Directory traversal
- FFprobe integration for duration/chapters

---

### Stage 4.3: Progress Tracking

**Goal:** Save and restore listening position for each audiobook.

#### Features
- Record current position: `audiotrack save-progress <book-id> 1h23m45s`
- Get current position: `audiotrack where <book-id>`
- Mark as complete: `audiotrack complete <book-id>`
- Resume info: `audiotrack resume` (shows where you left off)

#### CLI Commands

```bash
# Save your current position
audiotrack save-progress "atomic-habits" 2h15m30s

# Check where you are in a book
audiotrack where "atomic-habits"
# Output: Atomic Habits — 2h15m30s / 5h30m00s (41% complete)

# See all in-progress books
audiotrack resume
# Output:
# Currently listening:
# 1. Atomic Habits — 2h15m30s / 5h30m00s (41%)
# 2. The Pragmatic Programmer — 45m00s / 8h15m00s (9%)

# Mark as finished
audiotrack complete "atomic-habits"
```

#### Progress Storage

```go
type ProgressStore interface {
    SaveProgress(bookID string, position time.Duration) error
    GetProgress(bookID string) (*Progress, error)
    ListInProgress() ([]Progress, error)
    MarkComplete(bookID string) error
}
```

#### Concepts
- State persistence
- Time duration parsing and formatting
- Progress calculation and display

---

### Stage 4.4: Web Interface (Optional)

**Goal:** Browser-based UI for managing library and tracking progress.

#### Features
- View library with cover art
- Search and filter audiobooks
- Update progress via web form
- Mobile-friendly responsive design
- Embedded web server: `audiotrack serve`

#### Architecture

```
┌─────────────────────────────────────────┐
│            Web Browser                   │
└────────────────────┬────────────────────┘
                     │ HTTP
┌────────────────────▼────────────────────┐
│         Go HTTP Server                   │
│  ┌─────────────┐  ┌─────────────────┐   │
│  │  REST API   │  │  Static Files   │   │
│  │  /api/books │  │  (HTML/CSS/JS)  │   │
│  └──────┬──────┘  └─────────────────┘   │
│         │                                │
│  ┌──────▼──────┐                        │
│  │   Storage   │                        │
│  └─────────────┘                        │
└─────────────────────────────────────────┘
```

#### API Endpoints

```
GET    /api/books              — List all audiobooks
GET    /api/books/:id          — Get audiobook details
POST   /api/books              — Add new audiobook
DELETE /api/books/:id          — Remove audiobook
GET    /api/books/:id/progress — Get progress
PUT    /api/books/:id/progress — Update progress
GET    /api/resume             — List in-progress books
```

#### Concepts
- `net/http` server
- REST API design
- HTML templates or embedded SPA
- File serving

---

### Stage 4.5: Sync and Backup

**Goal:** Sync progress across devices.

#### Features
- Export library/progress to JSON: `audiotrack export backup.json`
- Import from backup: `audiotrack import-data backup.json`
- Optional cloud sync (S3, Google Drive, Dropbox)
- Conflict resolution for progress updates

#### Sync Strategies

1. **Manual export/import** — Simple JSON backup
2. **File-based sync** — Use existing cloud storage (Dropbox, iCloud)
3. **Server sync** — Self-hosted sync server (stretch goal)

#### Concepts
- JSON serialization
- Cloud storage APIs
- Conflict resolution strategies

---

### Stage 4.6: Audio Player Integration (Stretch Goal)

**Goal:** Optionally play audiobooks directly and auto-save progress.

#### Features
- Built-in CLI player: `audiotrack play <book-id>`
- Auto-save progress every 30 seconds
- Keyboard controls (pause, skip, speed)
- Chapter navigation

#### Player Options

1. **Shell out to mpv/VLC** — Simplest approach
2. **Use oto/beep** — Pure Go audio playback
3. **Web-based player** — HTML5 audio in web UI

```go
// Using external player
func playWithMPV(filePath string, startPos time.Duration) error {
    cmd := exec.Command("mpv",
        "--start="+formatDuration(startPos),
        filePath,
    )
    return cmd.Run()
}
```

#### Concepts
- Audio playback libraries
- External process integration
- Real-time progress tracking

---

### Stage 4.7: Polish and Features

#### Tasks
- Comprehensive `--help` output
- README with usage examples
- Statistics (listening time this week/month, books completed)
- Search by title/author
- Tags and collections
- Reading lists / "want to listen"

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
| http package | blitz, streamrip, audiotrack |
| crypto/aes | streamrip |
| Text parsing | streamrip |
| File I/O | streamrip, audiotrack |
| os/exec | streamrip (FFmpeg), audiotrack (mpv) |
| database/sql | audiotrack |
| HTML templates | audiotrack |
| REST API design | audiotrack |
| Audio metadata | audiotrack |

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
        │
        ▼
audiotrack
├── Reuses HTTP server patterns
├── Reuses output formatting
├── Adds: database/storage layer
├── Adds: REST API design
├── Adds: audio metadata parsing
└── Adds: web UI with templates
```

---

## Splitting Into Separate Repos

When tools are stable:

1. Publish gokit first, tag v0.1.0
2. Update each tool's go.mod to require published gokit
3. Create separate repos for each tool
4. Users install via:
   - `go install github.com/NickDiPreta/blitz@latest`
   - `go install github.com/NickDiPreta/streamrip@latest`
   - `go install github.com/NickDiPreta/netscan@latest`
   - `go install github.com/NickDiPreta/audiotrack@latest`
