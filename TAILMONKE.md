# Tailmonke TUI

A rich, real-time terminal UI for monitoring ping logs created by pingmonke.

## Features

### Display

- **Color-coded ping results:**
  - 🔴 **Red**: Timeout (no response)
  - 🟠 **Yellow/Orange**: High latency (≥100ms)
  - 🟢 **Green**: Good latency (<100ms)
  - 🟣 **Magenta**: Other statuses

- **Persistent headers** - Column headers always visible at the top
- **Running summary statistics** - Bottom of screen shows:
  - Total pings
  - Count of OK, Delayed, and Timeout pings
  - Average latency

- **Auto-scrolling tail** - Displays configurable number of most recent pings
- **Real-time updates** - Automatically detects when the log file is updated
- **File selection** - Auto-detects current period's log or specify manually
- **New file detection** - Notifies when a newer log file becomes available with option to switch (press **N**)

### Keyboard Controls

In interactive mode (default):

- **Ctrl+C** or **Q** - Exit the application
- **F5** or **Ctrl+R** - Regenerate the pings-summary.csv for the current period
- **N** - Switch to a newly detected log file (notification appears when available)

### Command Line Options

```bash
# Default: auto-detect latest log file and run interactive TUI
tailmonke --config config.yaml

# Specify a log file explicitly
tailmonke --file ~/ping-logs/2026-01-07-pings.csv

# Non-interactive mode (plain text output)
tailmonke --file ~/ping-logs/2026-01-07-pings.csv --non-interactive
```

### Configuration

In `config.yaml`, under the `tailmonke` section:

```yaml
tailmonke:
  # Number of log lines to display in the TUI (default: 20)
  lines_to_display: 20
```

## Architecture

- **bubbletea** - Modern TUI framework for Go
- **Real-time file monitoring** - Checks for log file updates every second
- **Async summary generation** - F5 regenerates summary without blocking UI
- **Cross-platform support** - Works on macOS, Linux, Windows

## Usage Examples

### Monitor the current debug session

```bash
tailmonke
```

### View a specific historical log

```bash
tailmonke --file ~/ping-logs/2026-01-07-pings.csv
```

### Display in non-interactive mode (CI/scripting)

```bash
tailmonke --file ~/ping-logs/2026-01-07-pings.csv --non-interactive
```

### Monitor and regenerate summaries as events occur

```bash
tailmonke --config config.yaml
# Press F5 to regenerate the summary file when events are detected
```
