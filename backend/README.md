# RunPlanner

A training plan builder for runners. Create plans, schedule workouts, and track progress.

- **Backend**: Go + Gin, SQLite, cookie-based sessions
- **Frontend**: Vue 3 + TypeScript, PrimeVue, Vite

## Local Development

### Backend

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Runs on http://localhost:8080.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Runs on http://localhost:5173 with API requests proxied to the backend.

## Docker Deployment

### Quick Start

```bash
# Set a secure session secret
export SESSION_SECRET=$(openssl rand -hex 32)

# Build and run
docker compose up -d --build
```

The app is available on port 3000. Point your reverse proxy (e.g. global nginx) to `localhost:3000`.

### Global Nginx Example

Add a server block for your domain in `/etc/nginx/sites-available/runplanner`:

```nginx
server {
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Then enable it and set up SSL:

```bash
sudo ln -s /etc/nginx/sites-available/runplanner /etc/nginx/sites-enabled/
sudo certbot --nginx -d yourdomain.com
sudo nginx -t && sudo systemctl reload nginx
```

### Useful Commands

```bash
docker compose up -d --build   # Build and start
docker compose logs -f         # View logs
docker compose down            # Stop
docker compose down -v         # Stop and delete database volume
```

## MCP Server

`backend/cmd/mcp` exposes the training plans over the [Model Context
Protocol](https://modelcontextprotocol.io), so an MCP client can view plans and
create plans and workouts. It talks to the running backend over its REST API and
signs in as one user, so start the backend first.

| Tool | Purpose |
|---|---|
| `list_training_plans` | List the user's plans with ids, dates and km totals |
| `get_training_plan` | Week-by-week, day-by-day view of one plan |
| `create_training_plan` | Create an empty plan (optionally with the race-day workout) |
| `create_workouts` | Add workouts in bulk, positioned by week and day of week |
| `create_workout` | Add a single workout on a calendar date |

Tools that take a `planId` fall back to the user's active plan when it is omitted.

### Configuration

| Variable | Default | Description |
|---|---|---|
| `RUNPLANNER_EMAIL` | _(required)_ | Account to sign in as |
| `RUNPLANNER_PASSWORD` | _(required)_ | That account's password |
| `RUNPLANNER_API_URL` | `http://localhost:8080` | Backend base URL |

### Against a deployed server

The MCP server runs on your own machine and reaches the backend over HTTP, so
point it at the public URL — `/api/` is already proxied there by the frontend
nginx, and nothing extra needs to be exposed:

```bash
export RUNPLANNER_API_URL=https://yourdomain.com
```

Use `https`: the server signs in with the account password, and over plain HTTP
to a remote host that travels in the clear (it warns on startup if you do).

### Using it

The repo ships a `.mcp.json` at the root, so in Claude Code it is enough to export
the credentials before starting:

```bash
export RUNPLANNER_EMAIL=you@example.com
export RUNPLANNER_PASSWORD=...
claude
```

For other clients, build the binary and point them at it:

```bash
cd backend && go build -o bin/runplanner-mcp ./cmd/mcp
```

```json
{
  "mcpServers": {
    "runplanner": {
      "command": "/absolute/path/to/runplanner/backend/bin/runplanner-mcp",
      "env": {
        "RUNPLANNER_EMAIL": "you@example.com",
        "RUNPLANNER_PASSWORD": "...",
        "RUNPLANNER_API_URL": "http://localhost:8080"
      }
    }
  }
}
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `SESSION_SECRET` | `change-me-in-production` | Session cookie encryption key |
| `DATABASE_URL` | `file:data/runplanner.db?...` | SQLite connection string |
| `PORT` | `8080` | Backend port (internal) |
| `CORS_ORIGINS` | _(none)_ | Extra allowed origins, comma-separated |
