# OOOMFS

> **⚠️ Experiment Notice:** This is an experimental project. API and behavior may change. Contributions and feedback welcome!

A single-binary local development environment providing Redis, S3, and SMTP servers with a web dashboard.

(halfway through i remembered that docker compose is a thing but it's too late to stop)

## Features

- **Redis** - Embedded Redis server (port 6379)
- **S3** - S3-compatible object storage (port 9000)
- **SMTP** - Email capture server (port 1025)
- **Dashboard** - Web UI for monitoring and control (port 8080)
- **Persistence** - Data saved automatically across restarts
- **Graceful Shutdown** - Clean shutdown saves all data

## Quick Start

```bash
# Build (includes embedded React frontend)
go build -o ooomfs ./cmd/server

# Run (starts all services)
./ooomfs

# Access dashboard
open http://localhost:8080
```

## Frontend Development

For local frontend development with hot reload:

```bash
# Terminal 1: Go backend
go run ./cmd/server

# Terminal 2: Frontend dev server
cd frontend && npm install && npm run dev
```

The frontend runs on `http://localhost:5173` with API requests proxied to the backend.

See [docs/frontend-integration.md](docs/frontend-integration.md) for full frontend development guide.

## Command Line Options

```bash
./ooomfs [options]

Options:
  -reset           Clear all persisted data before starting
  -config <file>   Path to config file
  -no-persist      Disable data persistence
  -redis-port <n>  Redis port (default: 6379)
  -s3-port <n>     S3 port (default: 9000)
  -smtp-port <n>   SMTP port (default: 1025)
  -dashboard-port <n>  Dashboard port (default: 8080)
  -redis-host <h>  Redis host (default: 127.0.0.1)
  -s3-host <h>     S3 host (default: 127.0.0.1)
  -smtp-host <h>   SMTP host (default: 127.0.0.1)
  -dashboard-host <h>  Dashboard host (default: 127.0.0.1)
```

## Environment Variables

```bash
OOOMFS_REDIS_PORT=6379
OOOMFS_S3_PORT=9000
OOOMFS_SMTP_PORT=1025
OOOMFS_DASHBOARD_PORT=8080
OOOMFS_REDIS_HOST=127.0.0.1
OOOMFS_S3_HOST=127.0.0.1
OOOMFS_SMTP_HOST=127.0.0.1
OOOMFS_DASHBOARD_HOST=127.0.0.1
OOOMFS_PERSIST_DIR=.ooomfs
OOOMFS_PERSIST_ENABLED=true

# DEVSTACK_* also works for backward compatibility
DEVSTACK_REDIS_PORT=6379
```

## Configuration File

Create `ooomfs.yaml` in your project directory:

```yaml
redis:
  port: 6379
  host: 127.0.0.1
  persist:
    enabled: true
    file: "~/.ooomfs/redis.dump"

s3:
  port: 9000
  host: 127.0.0.1
  persist:
    enabled: true
    directory: "~/.ooomfs/s3-data"

smtp:
  port: 1025
  host: 127.0.0.1
  persist:
    enabled: true
    directory: "~/.ooomfs/smtp-emails"

dashboard:
  port: 8080
  host: 127.0.0.1

persist:
  enabled: true
  dir: "~/.ooomfs"
```

## API Endpoints

### Status
```bash
curl http://localhost:8080/api/status
```

### Redis
```bash
# Start/Stop/Restart
curl -X POST http://localhost:8080/api/redis/start
curl -X POST http://localhost:8080/api/redis/stop
curl -X POST http://localhost:8080/api/redis/restart

# List keys
curl http://localhost:8080/api/redis/keys

# Delete key
curl -X DELETE http://localhost:8080/api/redis/key/mykey

# Force save
curl -X POST http://localhost:8080/api/redis/save
```

### S3
```bash
# Start/Stop/Restart
curl -X POST http://localhost:8080/api/s3/start
curl -X POST http://localhost:8080/api/s3/stop
curl -X POST http://localhost:8080/api/s3/restart

# List buckets
curl http://localhost:8080/api/s3/buckets

# Create bucket
curl -X PUT http://localhost:8080/api/s3/bucket/my-bucket

# Delete bucket
curl -X DELETE http://localhost:8080/api/s3/bucket/my-bucket

# List objects in bucket
curl http://localhost:8080/api/s3/bucket/my-bucket/objects

# Delete object
curl -X DELETE http://localhost:8080/api/s3/bucket/my-bucket/object/my-file.txt
```

### SMTP
```bash
# Start/Stop/Restart
curl -X POST http://localhost:8080/api/smtp/start
curl -X POST http://localhost:8080/api/smtp/stop
curl -X POST http://localhost:8080/api/smtp/restart

# List emails
curl http://localhost:8080/api/smtp/emails

# Get email details
curl http://localhost:8080/api/smtp/email/1

# Delete email
curl -X DELETE http://localhost:8080/api/smtp/email/1

# Clear all emails
curl -X POST http://localhost:8080/api/smtp/clear
```

### Logs
```bash
curl http://localhost:8080/api/logs
```

### Persist
```bash
# Force save all data
curl -X POST http://localhost:8080/api/persist
```

## Data Persistence

Data is persisted to `.ooomfs/` directory by default:

```
.ooomfs/
├── redis.dump       # Redis data
├── s3-data/        # S3 objects
└── smtp-emails/    # Captured emails
```

To disable persistence:
```bash
./ooomfs --no-persist
```

To reset all data:
```bash
./ooomfs --reset
```

## Architecture

```
cmd/server/main.go          # Entry point
internal/
├── config/                # Configuration loading
├── services/
│   ├── redis/             # Redis service
│   ├── s3/                # S3 service
│   └── smtp/              # SMTP service
└── dashboard/
    ├── server.go          # HTTP server + embedded frontend
    ├── embed.go           # go:embed declaration
    └── assets/            # Embedded React frontend (built from frontend/dist)
```

## Dependencies

- [miniredis](https://github.com/alicebob/miniredis) - Embedded Redis
- [gofakes3](https://github.com/johannesboyne/gofakes3) - S3 server
- Go standard library - SMTP server

## Contributing

This is an experiment! If you find it useful (or broken), contributions are welcome:
- Open issues for bugs or feature ideas
- PRs for improvements
- Feedback on API design and UX

## License

MIT
