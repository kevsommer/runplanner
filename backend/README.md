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

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `SESSION_SECRET` | `change-me-in-production` | Session cookie encryption key |
| `DATABASE_URL` | `file:data/runplanner.db?...` | SQLite connection string |
| `PORT` | `8080` | Backend port (internal) |
| `CORS_ORIGINS` | _(none)_ | Extra allowed origins, comma-separated |
