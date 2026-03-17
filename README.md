# WebUntis Go CLI

A Go CLI application optimized for AI agents to interact with the WebUntis API.

## Features

- **Compact JSON Output:** Optimized for token saving (no whitespace by default).
- **Session Persistence:** Login once, use session token for subsequent commands.
- **Filtering:** `--fields` flag to request only specific data.
- **WebUntis JSON-RPC 2.0:** Full client implementation.

## Installation

```bash
go mod tidy
go build -o webuntis cmd/webuntis/main.go
```

## Usage

### Authentication

```bash
./webuntis login --server demo.webuntis.com --school demo --username test --password secret
```

### Basic Commands

```bash
# List classes (compact JSON)
./webuntis classes

# List teachers with filtering
./webuntis teachers --fields id,name,foreName

# List subjects
./webuntis subjects

# List rooms
./webuntis rooms
```

### Timetable

```bash
# Get timetable for class ID 1 for today
./webuntis timetable --class 1

# Get timetable for specific date
./webuntis timetable --class 1 --date 2026-03-07

# Get timetable for range
./webuntis timetable --class 1 --date 2026-03-07 --end-date 2026-03-14
```

### Output Formatting

- Use `--pretty` for human-readable output.
- Use `--fields field1,field2` to reduce output size.

## Project Structure

- `cmd/webuntis`: Main entry point.
- `pkg/api`: JSON-RPC client and models.
- `pkg/cli`: Cobra commands and logic.
- `pkg/config`: Viper configuration and session storage.
