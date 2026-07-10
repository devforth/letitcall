# Let It Call

A SvelteKit and Go scheduling application. The production image compiles the Svelte portal, embeds it in the Go binary, and serves the UI and JSON API from one HTTP process.

## Run locally

Start the API on the portal's development-proxy port:

```sh
cd api
dotenvx run -f .env.local -- go run ./cmd/server
```

In a second terminal, start the portal:

```sh
cd portal
npm install
npm run dev -- --host 127.0.0.1 --port 41783 --strictPort
```

Opening the repository in VS Code automatically starts both development tasks on ports `41783` (portal) and `41784` (API). The automatic API task loads the committed `api/.env.local` settings through dotenvx.

There is no signup route. When the users table is empty, the API creates its first user from `FIRSTUSER__CREDENTIALS__EMAIL` and `FIRSTUSER__CREDENTIALS__PASSWORD`. Later users are created from Dashboard → Users.

## Docker

```sh
docker build -t letitcall .
docker run --rm -p 8080:80 \
  -v letitcall-data:/data \
  -e FIRSTUSER__CREDENTIALS__EMAIL=owner@example.com \
  -e FIRSTUSER__CREDENTIALS__PASSWORD='replace-with-at-least-12-characters' \
  letitcall
```

The server listens on `HTTP__PORT` (default `80`). LevelDB data is stored at `STORAGE__LEVELDB__PATH` (default `./data`, `/data` in Docker). Each logical table has its own LevelDB database.

## Configuration

All backend settings use structured uppercase environment variables:

- `HTTP__PORT` (default `80`)
- `HTTP__PUBLIC__URL`
- `HTTP__READ__TIMEOUT`, `HTTP__WRITE__TIMEOUT`, `HTTP__IDLE__TIMEOUT`, `HTTP__SHUTDOWN__TIMEOUT`
- `STORAGE__LEVELDB__PATH`
- `FIRSTUSER__CREDENTIALS__EMAIL`, `FIRSTUSER__CREDENTIALS__PASSWORD`
- `LOGIN__SESSION__TTL`, `LOGIN__SESSION__COOKIE__SECURE` (defaults to `true` for an HTTPS public URL)
- `LOGIN__PASSWORD__MAX_ATTEMPTS`, `LOGIN__PASSWORD__LOCKOUT`
- `LOGIN__OAUTH__GOOGLE__CLIENT_ID`, `LOGIN__OAUTH__GOOGLE__CLIENT_SECRET`
- `LOGIN__OAUTH__GOOGLE__REDIRECT_URL` (derived from `HTTP__PUBLIC__URL` when omitted)
- `LOGIN__OAUTH__GOOGLE__TOKEN_ENCRYPTION_KEY` (base64-encoded 32-byte key)

If any Google OAuth setting is present, all required Google settings must resolve successfully. The requested scopes include identity and permission to manage Google Calendar events. OAuth state uses PKCE, and Google tokens are encrypted at rest.

## Test and publish

```sh
cd api && go test ./...
cd ../portal && npm run check && npm run build
```

`publish.sh` creates and pushes multi-platform version and `latest` tags in one Buildx invocation:

```sh
./publish.sh 0.1.0 ghcr.io/example/letitcall
```
