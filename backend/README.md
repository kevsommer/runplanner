# RunPlanner Auth Skeleton (Go + Gin)

## Prereqs
- Go 1.22+

## Run
```bash
go mod tidy
go run ./cmd/server
```

The server listens on :8080. Override with `PORT`.

## Test with curl (session cookies handled by curl's cookie jar)
```bash
# Register (auto-login)
curl -i -c cookies.txt -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"supersecret"}' \
  http://localhost:8080/api/auth/register

# Who am I?
curl -i -c cookies.txt -b cookies.txt http://localhost:8080/api/auth/me

# Logout
curl -i -X POST -c cookies.txt -b cookies.txt http://localhost:8080/api/auth/logout

# Login
curl -i -X POST -c cookies.txt -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"supersecret"}' \
  http://localhost:8080/api/auth/login
```

## Notes
- This uses **cookie sessions** (server-side) via gin-contrib/sessions.
- Users are stored **in-memory** (lost on restart). Next step: swap `store.UserStore` to SQLite/Postgres.
- Passwords hashed with bcrypt.
- Controllers never expose `PasswordHash` because `json:"-"` on the model field.
