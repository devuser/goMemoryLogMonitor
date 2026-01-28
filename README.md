# MemoryLogMonitor

MemoryLogMonitor is a lightweight, real-time log monitoring system that receives logs via TCP and provides a modern web interface for querying, filtering, and sorting logs with in-memory storage.

![MemoryLogMonitor UI](https://github.com/user-attachments/assets/27c28342-9385-4549-ae6a-26e3eb61cc9a)

## Features

- **TCP Log Receiver**: Receives logs via TCP on port 9090
- **In-Memory Storage**: Stores logs in memory with LRU eviction (max 10,000 entries)
- **Thread-Safe**: Uses `sync.RWMutex` for concurrent access
- **Real-Time Monitoring**: Auto-refreshes every 5 seconds
- **Advanced Filtering**: Filter by keyword and date range
- **Sorting**: Sort logs by newest or oldest first
- **Pagination**: Paginated view for efficient browsing
- **Status API**: Monitor system status and log count
- **Modern Web UI**: Vue3 + TypeScript frontend with responsive design
- **Single Binary**: Frontend embedded in Go binary using embed

## Technology Stack

- **Backend**: Go with Gin framework
- **Frontend**: Vue3, TypeScript, Vite
- **Architecture**: TCP server (9090) + HTTP server (8080)
- **Thread Safety**: sync.RWMutex for concurrent access
- **Storage**: In-memory with LRU eviction

## Quick Start

### Prerequisites

- Go 1.20 or later
- Node.js 18+ and npm (for building the frontend)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/devuser/MemoryLogMonitor.git
cd MemoryLogMonitor
```

2. Build the frontend:
```bash
cd frontend
npm install
npm run build
cd ..
```

3. Copy the built frontend to the http package:
```bash
mkdir -p http/frontend/dist
cp -r frontend/dist/* http/frontend/dist/
```

4. Build the Go binary:
```bash
go build -o memorylogmonitor
```

5. Run the application:
```bash
./memorylogmonitor
```

The application will start:
- TCP server on port 9090 (receives logs)
- HTTP server on port 8080 (web UI and API)

### Sending Logs

Send logs via TCP using netcat, telnet, or any TCP client:

```bash
# Using netcat
echo "Application started successfully" | nc localhost 9090

# Using telnet
telnet localhost 9090
# Then type your log messages

# Using a script
for i in {1..10}; do
  echo "Log entry $i: Sample message" | nc localhost 9090
done
```

### Web Interface

Open your browser and navigate to:
```
http://localhost:8080
```

Features available in the UI:
- **Search**: Filter logs by keyword
- **Date Range**: Filter by start and end date
- **Sort**: Toggle between newest/oldest first
- **Pagination**: Navigate through pages of logs
- **Auto-refresh**: Logs update every 5 seconds
- **Status**: View system status and total log count

## API Endpoints

### GET /api/logs

Retrieve logs with optional filtering, sorting, and pagination.

**Query Parameters:**
- `keyword` (string): Filter by keyword in message
- `start` (RFC3339): Filter by start time
- `end` (RFC3339): Filter by end time
- `sort` (string): Sort order - "asc" or "desc" (default: "desc")
- `page` (int): Page number (default: 1)
- `pageSize` (int): Number of logs per page (default: 100, max: 1000)

**Example:**
```bash
curl "http://localhost:8080/api/logs?keyword=error&page=1&pageSize=50&sort=desc"
```

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2026-01-28T03:39:32Z",
      "message": "Log entry 5: Sample application log message"
    }
  ],
  "total": 10,
  "page": 1,
  "pageSize": 50
}
```

### GET /api/status

Get system status and log count.

**Example:**
```bash
curl http://localhost:8080/api/status
```

**Response:**
```json
{
  "status": "running",
  "logCount": 10,
  "timestamp": "2026-01-28T03:39:54Z"
}
```

## Architecture

### Components

1. **LogStore** (`logstore/logstore.go`)
   - Thread-safe log storage using `sync.RWMutex`
   - LRU eviction when capacity is exceeded
   - Filtering and retrieval methods

2. **TCP Server** (`tcp/server.go`)
   - Listens on port 9090
   - Accepts log messages (one per line)
   - Stores logs with current timestamp

3. **HTTP Server** (`http/server.go`)
   - Serves on port 8080
   - RESTful API endpoints
   - Embedded Vue3 frontend

4. **Frontend** (`frontend/`)
   - Vue3 + TypeScript SPA
   - Real-time log viewing
   - Advanced filtering and pagination

### Data Flow

```
TCP Client → TCP Server (9090) → LogStore ← HTTP Server (8080) → Web UI/API
```

## Configuration

Default settings are defined in `main.go`:

```go
const (
    MaxLogEntries = 10000  // Maximum log entries to store
    TCPPort = 9090         // TCP server port
    HTTPPort = 8080        // HTTP server port
)
```

To change these, modify the constants and rebuild.

## Development

### Building Frontend

```bash
cd frontend
npm install
npm run dev  # Development server
npm run build  # Production build
```

### Building Backend

```bash
go build -o memorylogmonitor
```

### Running Tests

```bash
go test ./...
```

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.
