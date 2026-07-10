# Let It Call

A SvelteKit and Go scheduling application. The production image compiles the Svelte portal, embeds it in the Go binary, and serves the UI and JSON API from one HTTP process.



## Docker

```sh
docker build -t letitcall .
docker run --rm -p 8080:80 \
  -v letitcall-data:/data \
  -e FIRSTUSER__CREDENTIALS__EMAIL=owner@example.com \
  -e FIRSTUSER__CREDENTIALS__PASSWORD='replace-with-at-least-12-characters' \
  letitcall
```

The server listens on `HTTP__PORT` (default `80`). LevelDB data is stored at `STORAGE__LEVELDB__PATH` (`/data` in Docker). Each logical table has its own LevelDB database, and user avatars are stored under its `content/avatars` subdirectory.

## Configuration

All backend settings use structured uppercase environment variables:

| Name | What it does | Default value |
| --- | --- | --- |
| `HTTP__PORT` | Sets the HTTP server port. | `80` |
| `HTTP__BASE__PATH` | Sets the URL path prefix where the application is served. A configured value must start with `/`, for example `/letitcall`. | Empty (served at `/`) |
| `STORAGE__LEVELDB__PATH` | Sets the directory containing the LevelDB databases. | `./data` (`/data` in Docker) |
| `FIRSTUSER__CREDENTIALS__EMAIL` | Sets the email of the user created when the users table is empty. Must be set together with the first-user password. | Not set |
| `FIRSTUSER__CREDENTIALS__PASSWORD` | Sets the password of the user created when the users table is empty. Must be set together with the first-user email. | Not set |
| `LOGIN__SESSION__TTL` | Sets how long an authenticated session remains valid. | `24h` |
| `LOGIN__PASSWORD__MAX_ATTEMPTS` | Sets the number of failed password attempts allowed before login is temporarily locked. | `5` |
| `LOGIN__PASSWORD__LOCKOUT` | Sets how long a password login remains locked after reaching the failed-attempt limit. | `15m` |
| `LOGIN__OAUTH__GOOGLE__CLIENT_ID` | Sets the Google OAuth client ID and enables Google login when the client secret is also set. | Not set |
| `LOGIN__OAUTH__GOOGLE__CLIENT_SECRET` | Sets the Google OAuth client secret. Must be set together with the client ID. | Not set |
| `MAILING__SENDING__MAILGUN__API_KEY` | Enables Mailgun booking notifications. Must be set with the Mailgun domain and sender. | Not set (email delivery is skipped) |
| `MAILING__SENDING__MAILGUN__DOMAIN` | Sets the Mailgun sending domain. | Not set |
| `MAILING__SENDING__MAILGUN__FROM` | Sets the Mailgun From address, such as `Let It Call <bookings@example.com>`. | Not set |

To enable Google OAuth, set both Google settings and add this authorized redirect URI to the Google client:

```text
https://your-host.example{HTTP__BASE__PATH}/auth/google/callback
```

For example, a HTTP__BASE__PATH of `/letitcall` produces `https://your-host.example/letitcall/auth/google/callback`. Omit the base path portion when the application is served at `/`. A TLS-terminating reverse proxy must preserve the request host and set `X-Forwarded-Proto: https` so the server forms the same redirect URI.

The requested scopes include identity and permission to manage Google Calendar events. OAuth state uses PKCE. Google tokens are encrypted with a random key generated on first use and kept in `google-token.key` under `STORAGE__LEVELDB__PATH`; persist the data directory across restarts.

When a booking is created, Mailgun delivery and Google Calendar delivery run in parallel. Email is sent to every event-type recipient when Mailgun is configured. A Google Calendar event is added separately to each recipient whose user account is Google-connected; recipients without a Google connection are silently skipped.

## For Developers

VS Code starts the portal on `41783` and API on `41784`. To run manually:

```sh
# Terminal 1
cd api
go run ./cmd/server

# Terminal 2
cd portal
pnpm install
pnpm run dev --host 127.0.0.1 --port 41783 --strictPort
```

Open `http://127.0.0.1:41783`; default login is `admin` / `admin`. There is no signup: the first user comes from `FIRSTUSER__CREDENTIALS__EMAIL` and `FIRSTUSER__CREDENTIALS__PASSWORD`; add later users in Users. Event types use immutable slugs, recipients, timezone-based weekly availability, and UTC bookings; manage them in Scheduling, book at `/book/{event-slug}`, and fetch public data from `/api/public/event-types/{event-slug}`.

### Google OAuth test credentials

Create a Web OAuth client, add the Google account as an OAuth test user, and register:

Enable the [Google Calendar API](https://console.cloud.google.com/apis/library/calendar-json.googleapis.com) for your Google Cloud Console project after creating the OAuth credentials.

```text
http://127.0.0.1:41783/auth/google/callback
```

```dotenv
# api/.env.local
LOGIN__OAUTH__GOOGLE__CLIENT_ID=your-client-id
LOGIN__OAUTH__GOOGLE__CLIENT_SECRET=your-client-secret
```

Do not commit credentials. Restart `Dev: Go API (41784)` and create a user with the Google account's email before signing in; Google login does not create users.

### Test and publish

```sh
cd api && go test ./...
cd ../portal && pnpm run check && pnpm run build
```

```sh
# Set VERSION and PACKAGE_NAME in publish.sh first.
./publish.sh
```

