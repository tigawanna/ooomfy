# DevStack Manager - Frontend Integration Guide

This guide explains how to build a frontend application that interfaces with the DevStack Manager API.

## Table of Contents

1. [API Overview](#api-overview)
2. [Setting Up Your Project](#setting-up-your-project)
3. [API Reference](#api-reference)
4. [Real-time Updates](#real-time-updates)
5. [Example Implementations](#example-implementations)
6. [Frontend Development](#frontend-development)
7. [Embedding Process](#embedding-process)

---

## API Overview

The DevStack Manager exposes REST APIs for managing local development services:

- **Redis** - In-memory data store
- **S3** - Object storage (MinIO-compatible)
- **SMTP** - Mail catch server (MailHog-compatible)

### Base URLs

| Service | URL |
|---------|-----|
| Dashboard API | `http://localhost:8080/api/` |
| Logs Stream | `http://localhost:8081/api/` |

---

## Setting Up Your Project

### Using Fetch (Vanilla JS)

```javascript
const API_BASE = 'http://localhost:8080/api';

// GET request
async function getStatus() {
  const res = await fetch(`${API_BASE}/status`);
  return res.json();
}

// POST request
async function startService(service) {
  await fetch(`${API_BASE}/${service}/start`, { method: 'POST' });
}
```

### Using React Query (Recommended)

```javascript
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

const API_BASE = '/api';  // Use relative URL if proxy configured

// Query for status
function useStatus() {
  return useQuery({
    queryKey: ['status'],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/status`);
      return res.json();
    },
    refetchInterval: 3000,  // Poll every 3 seconds
  });
}

// Mutation for actions
function useServiceControl(service) {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (action) => {
      await fetch(`${API_BASE}/${service}/${action}`, { method: 'POST' });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['status'] });
    },
  });
}
```

### Vite Proxy Configuration

To avoid CORS issues, configure your Vite dev server:

```javascript
// vite.config.js
export default defineConfig({
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

---

## API Reference

### Service Status

#### `GET /api/status`

Returns the running status of all services.

**Response:**
```json
{
  "redis": "running",
  "s3": "running",
  "smtp": "running",
  "dashboard": "running"
}
```

### Redis API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/redis/` | GET | Get Redis stats |
| `/api/redis/keys` | GET | List all keys |
| `/api/redis/{key}` | GET | Get key value |
| `/api/redis/{key}` | DELETE | Delete a key |
| `/api/redis/start` | POST | Start Redis |
| `/api/redis/stop` | POST | Stop Redis |
| `/api/redis/restart` | POST | Restart Redis |
| `/api/redis/save` | POST | Save data to disk |

**Example - List keys:**
```javascript
const keys = await fetch('/api/redis/keys').then(r => r.json());
// Returns: ["key1", "key2", ...]
```

**Example - Get key value:**
```javascript
const data = await fetch('/api/redis/mykey').then(r => r.json());
// Returns: { key: "mykey", value: "..." }
```

### S3 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/s3/` | GET | Get S3 stats |
| `/api/s3/buckets` | GET | List buckets |
| `/api/s3/bucket/{bucket}/objects` | GET | List objects in bucket |
| `/api/s3/bucket/{bucket}` | PUT | Create bucket |
| `/api/s3/bucket/{bucket}` | DELETE | Delete bucket |
| `/api/s3/bucket/{bucket}/object/{key}` | DELETE | Delete object |
| `/api/s3/start` | POST | Start S3 |
| `/api/s3/stop` | POST | Stop S3 |
| `/api/s3/restart` | POST | Restart S3 |

**Example - List buckets:**
```javascript
const buckets = await fetch('/api/s3/buckets').then(r => r.json());
// Returns: [{ name: "my-bucket", creationDate: "..." }, ...]
```

**Example - Create bucket:**
```javascript
await fetch('/api/s3/bucket/my-new-bucket', { method: 'PUT' });
```

### SMTP API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/smtp/` | GET | Get SMTP stats |
| `/api/smtp/emails` | GET | List all emails |
| `/api/smtp/email/{id}` | GET | Get email by ID |
| `/api/smtp/email/{id}` | DELETE | Delete email |
| `/api/smtp/start` | POST | Start SMTP |
| `/api/smtp/stop` | POST | Stop SMTP |
| `/api/smtp/restart` | POST | Restart SMTP |
| `/api/smtp/clear` | POST | Clear all emails |

**Example - List emails:**
```javascript
const emails = await fetch('/api/smtp/emails').then(r => r.json());
// Returns: [{ id: "...", from: "...", to: [...], subject: "...", ... }, ...]
```

### Logs API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/logs` | GET | Get all logs (from dashboard server) |
| `/api/logs/all` | GET | Get all logs (from main server) |
| `/api/logs/stream` | GET | Stream logs via SSE |

**Example - Get logs:**
```javascript
const logs = await fetch('http://localhost:8081/api/logs/all').then(r => r.json());
// Returns: ["[12:34:56] Message 1", "[12:34:57] Message 2", ...]
```

### Persist API

#### `POST /api/persist`

Saves all data (Redis, S3) to disk.

```javascript
await fetch('/api/persist', { method: 'POST' });
```

---

## Real-time Updates

### Log Streaming (Server-Sent Events)

To receive real-time log updates:

```javascript
const eventSource = new EventSource('http://localhost:8081/api/logs/stream');

eventSource.onmessage = (event) => {
  console.log('New log:', event.data);
  // Append to your logs display
};

eventSource.onerror = (err) => {
  console.error('SSE error:', err);
};

// Cleanup on unmount
eventSource.close();
```

### Polling for Status

For reactive status updates in React:

```javascript
import { useEffect } from 'react';

function useServiceStatus(refetchInterval = 3000) {
  const [status, setStatus] = useState(null);
  
  useEffect(() => {
    const fetchStatus = async () => {
      try {
        const res = await fetch('/api/status');
        const data = await res.json();
        setStatus(data);
      } catch (err) {
        console.error('Failed to fetch status:', err);
      }
    };
    
    fetchStatus();
    const interval = setInterval(fetchStatus, refetchInterval);
    
    return () => clearInterval(interval);
  }, [refetchInterval]);
  
  return status;
}
```

---

## Example Implementations

### Status Card Component (React)

```tsx
import { useServiceControl } from '@/hooks/use-api';

export function StatusCard({ name, status, port, service }) {
  const { start, stop, restart } = useServiceControl(service);
  const isRunning = status === 'running';

  return (
    <div className="card">
      <h3>{name}</h3>
      <p>Port: {port}</p>
      <span className={isRunning ? 'status-run' : 'status-stop'}>
        {isRunning ? 'Running' : 'Stopped'}
      </span>
      
      <div className="actions">
        {isRunning ? (
          <button onClick={() => stop.mutate()}>Stop</button>
        ) : (
          <button onClick={() => start.mutate()}>Start</button>
        )}
        <button onClick={() => restart.mutate()}>Restart</button>
      </div>
    </div>
  );
}
```

### Logs Panel (React)

```tsx
import { useEffect, useState } from 'react';

export function LogsPanel() {
  const [logs, setLogs] = useState([]);

  useEffect(() => {
    // Initial fetch
    fetch('http://localhost:8081/api/logs/all')
      .then(r => r.json())
      .then(setLogs);

    // SSE stream
    const eventSource = new EventSource('http://localhost:8081/api/logs/stream');
    eventSource.onmessage = (event) => {
      setLogs(prev => [...prev, event.data]);
    };

    return () => eventSource.close();
  }, []);

  return (
    <div className="logs-panel">
      {logs.map((log, i) => (
        <div key={i} className="log-entry">{log}</div>
      ))}
    </div>
  );
}
```

### Redis Explorer

```tsx
import { useRedisKeys, useRedisKey, useDeleteRedisKey } from '@/hooks/use-api';

export function RedisExplorer() {
  const { data: keys = [] } = useRedisKeys();
  const deleteKey = useDeleteRedisKey();

  return (
    <div>
      <h2>Redis Keys ({keys.length})</h2>
      <ul>
        {keys.map(key => (
          <li key={key}>
            {key}
            <button onClick={() => deleteKey.mutate(key)}>Delete</button>
          </li>
        ))}
      </ul>
    </div>
  );
}
```

### S3 Browser

```tsx
import { useS3Buckets, useS3Objects, useS3CreateBucket, useS3DeleteBucket } from '@/hooks/use-api';

export function S3Browser() {
  const { data: buckets = [] } = useS3Buckets();
  const [selectedBucket, setSelectedBucket] = useState(null);
  const { data: objects = [] } = useS3Objects(selectedBucket);
  const createBucket = useS3CreateBucket();
  const deleteBucket = useS3DeleteBucket();

  const handleCreate = async () => {
    const name = prompt('Bucket name:');
    if (name) createBucket.mutate(name);
  };

  return (
    <div>
      <h2>S3 Buckets</h2>
      <button onClick={handleCreate}>Create Bucket</button>
      
      <div className="buckets">
        {buckets.map(bucket => (
          <div key={bucket.name} onClick={() => setSelectedBucket(bucket.name)}>
            <h3>{bucket.name}</h3>
            <button onClick={() => deleteBucket.mutate(bucket.name)}>Delete</button>
          </div>
        ))}
      </div>

      {selectedBucket && (
        <div className="objects">
          <h3>Objects in {selectedBucket}</h3>
          {objects.map(obj => (
            <div key={obj.key}>{obj.key} ({obj.size} bytes)</div>
          ))}
        </div>
      )}
    </div>
  );
}
```

### SMTP Email Viewer

```tsx
import { useSMTPEmails, useSMTPEmail, useDeleteSMTPEmail, useClearSMTPEmails } from '@/hooks/use-api';

export function EmailViewer() {
  const { data: emails = [] } = useSMTPEmails();
  const { data: selectedEmail } = useSMTPEmail(selectedId);
  const [selectedId, setSelectedId] = useState(null);
  const deleteEmail = useDeleteSMTPEmail();
  const clearEmails = useClearSMTPEmails();

  return (
    <div className="email-viewer">
      <div className="email-list">
        <h2>Emails ({emails.length})</h2>
        <button onClick={() => clearEmails.mutate()}>Clear All</button>
        
        {emails.map(email => (
          <div 
            key={email.id} 
            className={selectedId === email.id ? 'selected' : ''}
            onClick={() => setSelectedId(email.id)}
          >
            <strong>{email.from}</strong>
            <p>{email.subject}</p>
            <small>{email.date}</small>
          </div>
        ))}
      </div>

      {selectedEmail && (
        <div className="email-content">
          <h3>{selectedEmail.subject}</h3>
          <p><strong>From:</strong> {selectedEmail.from}</p>
          <p><strong>To:</strong> {selectedEmail.to.join(', ')}</p>
          <hr />
          <div dangerouslySetInnerHTML={{ __html: selectedEmail.html || selectedEmail.text }} />
        </div>
      )}
    </div>
  );
}
```

---

## Error Handling

All API errors return appropriate HTTP status codes:

- `200` - Success
- `404` - Resource not found
- `500` - Server error

```javascript
async function safeFetch(url, options = {}) {
  try {
    const res = await fetch(url, options);
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}: ${await res.text()}`);
    }
    return await res.json();
  } catch (err) {
    console.error('API Error:', err);
    throw err;
  }
}
```

---

## Quick Reference

### Endpoints Summary

```
GET    /api/status
GET    /api/redis/
GET    /api/redis/keys
GET    /api/redis/:key
DELETE /api/redis/:key
POST   /api/redis/start|stop|restart|save
GET    /api/s3/
GET    /api/s3/buckets
GET    /api/s3/bucket/:bucket/objects
PUT    /api/s3/bucket/:bucket
DELETE /api/s3/bucket/:bucket
DELETE /api/s3/bucket/:bucket/object/:key
POST   /api/s3/start|stop|restart
GET    /api/smtp/
GET    /api/smtp/emails
GET    /api/smtp/email/:id
DELETE /api/smtp/email/:id
POST   /api/smtp/start|stop|restart|clear
POST   /api/persist
GET    /api/logs
GET    /api/logs/all
GET    /api/logs/stream
```

### Common Ports

| Service | Port |
|---------|------|
| Dashboard | 8080 |
| Logs Server | 8081 |
| Redis | 6379 |
| S3 (MinIO) | 9000 |
| SMTP (MailHog) | 1025 |

---

## OpenAPI Specification

The full OpenAPI 3.0 specification is available at: `docs/openapi.json`

You can view it interactively using:
- [Swagger Editor](https://editor.swagger.io/)
- [Stoplight](https://stoplight.io/)
- [Redoc](https://redocly.github.io/redoc/)

---

## Frontend Development

### Local Development Workflow

When developing the frontend, you have two options:

#### Option 1: Standalone Development (Recommended)

Develop with hot reload by running the frontend dev server separately:

```bash
# Terminal 1: Start the Go backend
go run ./cmd/server

# Terminal 2: Start the frontend dev server
cd frontend
npm install
npm run dev
```

The frontend runs on `http://localhost:5173` with Vite's proxy forwarding API requests to the Go backend on port 8080.

#### Option 2: Against Running Backend

If the Go backend is already running:

```bash
cd frontend
npm run dev
```

The Vite proxy automatically forwards `/api` requests to `http://localhost:8080`.

### Vite Proxy Configuration

The `vite.config.ts` is pre-configured with proxy settings:

```typescript
// vite.config.ts
server: {
  port: 5173,
  proxy: {
    '/api/logs/stream': {
      target: 'http://localhost:8081',
      changeOrigin: true,
    },
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
},
```

---

## Embedding Process

The frontend is embedded into the Go binary using `go:embed`. This means:

- **Single binary**: No external files needed at runtime
- **Portable**: Works from any directory
- **Fast startup**: No file system checks for frontend assets

### How It Works

1. **Frontend is built** to `frontend/dist/`
2. **Files are copied** to `internal/dashboard/assets/`
3. **Go embeds** the assets directory at compile time
4. **Server serves** files from the embedded file system

### File Locations

```
ooomfy/
├── frontend/
│   └── dist/                    # Built frontend (gitignored)
│       ├── index.html
│       └── assets/
│           ├── index-xxx.css
│           └── index-xxx.js
│
├── internal/dashboard/
│   ├── assets/                  # Embedded copy of frontend/dist
│   │   ├── index.html
│   │   └── assets/
│   ├── embed.go                 # go:embed declaration
│   └── server.go                # Serves embedded files
│
└── cmd/server/main.go           # Entry point
```

### Build Commands

#### Development Build (with embedded frontend)

```bash
# 1. Build frontend
cd frontend && npm run build

# 2. Copy to embed directory (handled by build script)
cp -r frontend/dist internal/dashboard/assets

# 3. Build Go binary
go build -o ooomfs ./cmd/server
```

Or use the combined build script:

```bash
# From project root
make build

# Or manually:
npm run build --prefix frontend && cp -r frontend/dist internal/dashboard/assets && go build -o ooomfs ./cmd/server
```

#### Quick Rebuild (if frontend already built)

If you've already built the frontend and just need to rebuild the Go binary:

```bash
go build -o ooomfs ./cmd/server
```

### For CI/CD Pipelines

In automated builds:

```bash
#!/bin/bash
set -e

# Build frontend
cd frontend && npm ci && npm run build && cd ..

# Copy to embed directory
rm -rf internal/dashboard/assets
cp -r frontend/dist internal/dashboard/assets

# Build Go binary
CGO_ENABLED=0 go build -ldflags="-s -w" -o ooomfs ./cmd/server
```

### Troubleshooting

**"Frontend not embedded" error:**
- Ensure `frontend/dist/` has been built
- Run: `cp -r frontend/dist internal/dashboard/assets`
- Rebuild the Go binary

**Changes not appearing:**
- Always rebuild the Go binary after frontend changes
- Run: `go build -o ooomfs ./cmd/server`

**Hot reload not working:**
- Hot reload only works in standalone dev mode (`npm run dev`)
- The embedded version does not support hot reload
