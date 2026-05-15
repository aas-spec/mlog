# mlog

Simple logger with daily file rotation, log levels, and multiple named loggers.

## Versions

### v1 — stable, no breaking changes (existing projects)

```go
import "github.com/aas-spec/mlog"
```

### v2 — fixed, recommended for new projects

```go
import "github.com/aas-spec/mlog/v2"
```

**Fixes in v2:**
- Correct file cleanup: sorts all log files and keeps only the last `StoreDays` (v1 used a broken offset loop)
- Glob mask uses each logger's own path, not the default logger path
- `Panic` correctly spreads variadic args (v1 wrapped them in a slice)
- `Printf`/`Logf`/`Outf` parameter renamed from `fmt` to `format` (shadowed the `fmt` package)
- `GetTimeStamp` is now exported
- `DisableStdout=1` env variable suppresses stdout output
- New: `SetLogUseOwnDir`, `StdPrintf`, `DefUseOwnDir` constant

## API

Both versions share the same calling interface.

```go
// Default logger
mlog.Log("message")
mlog.Logln("message")
mlog.Logf("value: %d", 42)
mlog.Panic("fatal error")

// Level-aware (skips if level > configured level)
mlog.LLog(3, "debug message")
mlog.LLogf(3, "value: %d", 42)

// Named logger
mlog.Out("myservice", "message")
mlog.Outf("myservice", "value: %d", 42)
mlog.LOut("myservice", 3, "debug")

// Configuration
mlog.SetLogLevel("myservice", 3)
mlog.SetStoreDays("myservice", 14)

// v2 only
mlog.SetLogUseOwnDir("myservice", true) // logs/myservice/app-date.log
mlog.StdPrintf("plain: %s", "text")     // stdout only, with timestamp
```

## Log file location

Logs are written to `<executable_dir>/logs/<executable_name>[-LoggerID]_<date>.log`

## Environment

| Variable         | Values | Effect              |
|------------------|--------|---------------------|
| `DisableStdout`  | `1`    | Write to file only (v2) |
