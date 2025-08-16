# forwardemail-cli — Architecture

*Status: draft v0.1*

## 1) Purpose & Goals

`forwardemail-cli` is a Go-based command-line interface to manage Forward Email accounts and resources through their public REST API. The tool targets:

- **Operators / DevOps**: automate domain and alias management, download logs, verify DNS/SMTP, send or inspect outbound emails.
- **Developers**: scriptable, predictable CLI with stable output formats for CI/CD, GitOps, and provisioning flows.

**Design principles**

- Minimal dependencies; prefer stdlib.
- Clear separation between **SDK (api client)** and **CLI commands**.
- Consistent UX: idempotent operations where possible, safe-by-default flags, human and machine-friendly output.
- Extensibility for upcoming API surface (Contacts/Calendars/Messages/Folders).

---

## 2) Scope (v1)

- **Accounts**: get/update profile.
- **Domains**: CRUD, verify DNS/SMTP, catch‑all passwords, invites, members, settings (protections, quotas, webhooks, retention).
- **Aliases**: CRUD, recipients (emails/FQDN/IP/webhook URLs), IMAP+PGP flags, quotas, vacation responder, password generation.
- **Emails** (outbound): list/get/delete, send via structured fields or RFC822 `--raw`, check daily limit.
- **Logs**: request deliverability logs (csv.gz by email delivery); respect 10 req/day limit.
- **Encrypt**: helper to encrypt plaintext TXT strings for DNS.

**Out of scope (v1)**

- Interactive TUI; IMAP/CardDAV/CalDAV/Messages/Folders (planned in roadmap).
- Account sign-up flow.

---

## 3) Non‑Functional Requirements

- **Portability**: Linux/macOS/Windows (amd64/arm64).
- **Performance**: async pagination where safe; streaming writers for large payloads.
- **Reliability**: exponential backoff with jitter; context timeouts; retries on idempotent GET/HEAD and safe POSTs where documented.
- **Security**: never print secrets; on-disk secrets 0600; optional OS keyring.
- **Observability**: structured logs (stderr), trace-friendly request IDs, `--verbose` and `--debug`.
- **Deterministic output**: stable key order for JSON, predictable tables/CSV.

---

## 4) Security Model

- **Auth**: HTTP Basic with **API key as username**, empty password (service-wide).
  - Some alias-scoped endpoints (future Messages/Contacts/Calendars) will require alias username/password instead of the API key. The SDK exposes `WithCredentials(username, password)` to override per-call.
- **Secret sources** (merged with precedence): flag → env → config → keyring.
- **Env vars**: `FORWARDEMAIL_API_KEY`, `FORWARDEMAIL_USERNAME`, `FORWARDEMAIL_PASSWORD` (alias mode), `FORWARDEMAIL_BASE_URL` (default `https://api.forwardemail.net`).
- **Config path**: `${XDG_CONFIG_HOME:-~/.config}/forwardemail/config.yaml` or Windows `%AppData%\forwardemail\config.yaml`.
- **Profiles**: multiple named credential sets in config; select via `--profile` or `FORWARDEMAIL_PROFILE`.
- **Redaction**: redact tokens in logs and panic reports; mask values in error context.

---

## 5) High‑Level Architecture

```
cmd/forwardemail-cli          // Cobra command tree
pkg/cli                       // command wiring, flag parsing, output formatting
pkg/api                       // typed SDK (HTTP client, models, services)
pkg/config                    // config loading, profiles, keyring integration
pkg/output                    // table/json/yaml/csv renderers
pkg/pager                     // pagination helpers & iterators
pkg/validate                  // input validation (emails, domains, URLs, sizes)
pkg/x/httpretry               // retry transport, backoff, rate limiting
pkg/x/keyring                 // optional OS keyring wrapper
internal/testutil             // golden snapshots, mock server, fixtures
```

**Layering**

- `pkg/api`: pure Go SDK; no CLI concerns.
- `pkg/cli`: composes `pkg/api` + IO; returns domain errors with exit codes.
- `pkg/output`: isolated renderers so `--output json` is consistent across commands.

---

## 6) Command Surface (v1)

```
forwardemail-cli (alias: fe)
  account get|update --email --given-name --family-name --avatar-url

  domains list [--q --sort --paginate]
          create <domain> [flags]
          get <domain>
          verify-records <domain>
          verify-smtp <domain>
          update <domain> [flags]
          delete <domain>
          catchall <domain> passwords list|create [--password --description] | delete <token-id>
          invites <domain> accept|create --email --group admin|user | remove --email
          members <domain> update <member-id> --group admin|user | remove <member-id>

  aliases list <domain> [--q --name --recipient --sort --paginate]
          create <domain> [--name --recipients ... --labels ... --is-enabled ...
                           --error-code-if-disabled 250|421|550 --has-imap ...
                           --has-pgp ... --public-key ... --max-quota ...
                           --vacation-enable --vacation-start --vacation-end
                           --vacation-subject --vacation-message]
          get <domain> (--id <id> | --name <name>)
          update <domain> --id <id> [same flags as create]
          delete <domain> --id <id>
          gen-password <domain> --id <id> [--password --override --email-instructions <addr>]

  emails  limit
          list [--q --domain --sort --paginate]
          send [--from --to --cc --bcc --subject --text --html --attach path ...
                --priority high|normal|low --header "X-...: v" | --raw path.eml]
          get <id>
          delete <id>

  logs    download [--domain ... --q ... --bounce-category ... --response-code ...]

  dns     encrypt --input "plaintext"

Global flags:
  --profile, --output (table|json|yaml|csv), --no-headers, --compact,
  --timeout, --page-size, --paginate, --verbose, --debug
```

---

## 7) API Coverage Matrix (v1)

| Area      | Endpoint                                                       | CLI mapping               |          |          |
| --------- | -------------------------------------------------------------- | ------------------------- | -------- | -------- |
| Account   | `GET/PUT /v1/account`                                          | \`account get             | update\` |          |
| Emails    | `GET /v1/emails/limit`                                         | `emails limit`            |          |          |
|           | `GET/POST /v1/emails`                                          | \`emails list             | send\`   |          |
|           | `GET/DELETE /v1/emails/:id`                                    | \`emails get              | delete\` |          |
| Domains   | `GET/POST /v1/domains`                                         | \`domains list            | create\` |          |
|           | `GET/PUT/DELETE /v1/domains/:domain`                           | \`domains get             | update   | delete\` |
|           | `GET /v1/domains/:domain/verify-records`                       | `domains verify-records`  |          |          |
|           | `GET /v1/domains/:domain/verify-smtp`                          | `domains verify-smtp`     |          |          |
| Catch‑all | `GET/POST /v1/domains/:domain/catch-all-passwords`             | `domains catchall ...`    |          |          |
|           | `DELETE /.../catch-all-passwords/:token_id`                    | `domains catchall delete` |          |          |
| Invites   | `GET/POST/DELETE /v1/domains/:domain/invites`                  | `domains invites ...`     |          |          |
| Members   | `PUT/DELETE /v1/domains/:domain/members/:member_id`            | `domains members ...`     |          |          |
| Aliases   | `GET/POST /v1/domains/:domain/aliases`                         | \`aliases list            | create\` |          |
|           | `GET/PUT/DELETE /v1/domains/:domain/aliases/:alias_id`         | \`aliases get             | update   | delete\` |
|           | `POST /v1/domains/:domain/aliases/:alias_id/generate-password` | `aliases gen-password`    |          |          |
| Encrypt   | `POST /v1/encrypt`                                             | `dns encrypt`             |          |          |

> Note: Endpoints accept domain **name or ID**. Some alias `GET` supports lookup by `name`; we expose `--name` for convenience.

---

## 8) Data Model (Go)

- **Domain**: id, name, plan, protections (spam/phishing/virus flags), recipient verification, MX checks policy, bounce webhook, outbound retention days, per-alias quota, custom SMTP port, invites, members.
- **Alias**: id, domain\_id, name, recipients (emails/FQDN/IP/webhook URLs), labels, description, enabled flag, disable code (250/421/550), IMAP/PGP flags and public key, quota, vacation responder fields, created/updated timestamps.
- **Email**: id, from, to/cc/bcc, subject, text/html/attachments, headers, priority, status (pending/queued/deferred/sent), created/updated.
- **CatchAllToken**: token\_id, description, created\_at.
- **Invite/Member**: email, group/role, status.

The SDK will expose typed request/response structs mirroring server fields with `json` tags and `omitempty`. Unknown JSON preserved via `map[string]any` for forward compatibility where useful.

---

## 9) Configuration & Credentials

**Resolution order:** CLI flags → Env vars → Profile config → Defaults → Interactive (only if `--interactive` is set).

**Config file schema (YAML):**

```yaml
current_profile: default
profiles:
  default:
    base_url: https://api.forwardemail.net
    api_key: env:FORWARDEMAIL_API_KEY   # references env var
    timeout: 15s
    output: table
  staging:
    base_url: https://staging.api.forwardemail.net
    api_key: keyring:forwardemail/staging
```

**Keyring** (optional): `profiles.<name>.api_key: keyring:<service>/<account>` stored via OS keyring. CLI offers `fe auth login --store keyring` to save it.

---

## 10) HTTP Client & Transport

- **Base**: `net/http` client with `Transport` tuned for keep-alives; default timeout 15s, override via `--timeout`.
- **Auth**: `req.SetBasicAuth(username, password)` where username=`apiKey`, password="" (or alias creds for future endpoints).
- **Retry policy**:
  - Idempotent methods (`GET`, `DELETE`, some `PUT`) retried on network errors and `429/5xx` with exponential backoff + jitter (`1s → 30s`, max 5 tries).
  - Non-idempotent (`POST /v1/emails`) retried only on explicit safe error class (e.g., connect timeouts before write) unless `--force-retry`.
- **Rate limiting**: token bucket limiter (configurable), respecting `Retry-After` if present.
- **User-Agent**: `forwardemail-cli/<version> (<os>/<arch>) go/<version>`.
- **Pagination**: helper that reads `X-Page-*` & RFC5988 `Link` headers; exposes `Pager.Next(ctx)` iterator and `--paginate` flag with page-size clamped to server limits.

---

## 11) Validation & UX Details

- **Domains**: FQDN validation, punycode support.
- **Emails**: RFC5322-lite validation; allow `Name <addr>` but server ultimately validates.
- **Recipients**: accept comma/space/newline lists; classify items into email/FQDN/IP/URL (webhook) at parse time.
- **Sizes** (quotas): accept units (`MB`, `GB`); normalize to server-required format.
- **Vacation**: ISO8601 timestamps or `YYYY-MM-DD`; localize to UTC when sending.
- **Disable codes**: enforce enum {250, 421, 550}.

---

## 12) Output & Formatting

- **Formats**: `table` (default), `json`, `yaml`, `csv` (where meaningful).
- **Stability**: JSON key ordering stable via canonical marshalling.
- **Humanization**: sizes and durations human-readable in `table`; raw values in `json`/`yaml`.
- **Wide/Compact**: auto-detect TTY; `--no-headers` for scripting.

---

## 13) Errors & Exit Codes

- `0` success.
- `1` generic error.
- `2` validation/usage error.
- `3` auth/permission error.
- `4` not found.
- `5` rate-limit or quota exceeded (e.g., logs 10/day).
- `6` network/timeout.

Errors are printed to **stderr** in compact human text and to **stdout** only for requested output payloads.

---

## 14) Logging & Tracing

- `--verbose` adds request method/path and elapsed time.
- `--debug` dumps headers (with redaction) and response codes; may include compact JSON bodies when non-2xx.
- Correlation IDs: echo `X-Request-ID` if present.
- Optional OpenTelemetry hooks (env-driven) for traces and metrics.

---

## 15) Implementation Patterns

### Cobra wiring

- Each subcommand constructs a `*api.Client` via a shared `Factory` that resolves config/profile/env and injects transports.
- Command handlers:
  1. Parse/validate flags.
  2. Call service method (`client.Domains.Create(ctx, req)`).
  3. Render with `pkg/output` based on `--output`.

### API client shape

```go
// pkg/api/client.go
type Client struct {
    HTTP *http.Client
    BaseURL *url.URL
    auth AuthProvider
    Domains *DomainService
    Aliases *AliasService
    Emails *EmailService
    Account *AccountService
    Logs *LogService
    Crypto *CryptoService
}

type AuthProvider interface { Apply(req *http.Request) }

type BasicAuth struct{ Username, Password string }
func (b BasicAuth) Apply(r *http.Request) { r.SetBasicAuth(b.Username, b.Password) }
```

### Pagination helper

```go
type PageInfo struct{ Current, Items, Pages int; Next, Prev string }

type Pager[T any] struct {
    Fetch func(ctx context.Context, page, limit int) ([]T, *PageInfo, error)
}
```

---

## 16) Testing Strategy

- **Unit tests** for parsers, validators, renderers.
- **SDK contract tests** against a **mock server** that simulates:
  - Success/4xx/5xx, timeouts, `Retry-After`, pagination headers, log quota errors.
- **CLI e2e** with golden snapshots for stdout/stderr/exit codes.
- **Record/replay** (opt-in) with `go-vcr` for selective real API calls (redacted), gated in CI.

**Coverage targets**: 80% packages, 100% for validators/formatters.

---

## 17) CI/CD & Releases

- **CI**: Go 1.23 matrix (linux/mac/win), `go vet`, `staticcheck`, unit+e2e tests, race detector.
- **Releases**: GoReleaser → GitHub Releases (archives + checksums), Homebrew tap, Scoop manifest, AUR recipe.
- **Versioning**: SemVer; include API compatibility notes in CHANGELOG.

---

## 18) Telemetry (Opt‑in)

- Anonymous usage metrics (command name, success/fail, duration, OS/arch).
- Disabled by default; enable via `FORWARDEMAIL_CLI_TELEMETRY=1`.
- Respect `DO_NOT_TRACK`.

---

## 19) Dependency Policy

- Core: stdlib + `spf13/cobra` (CLI), `spf13/pflag`, `spf13/viper` (config) optional.
- Optional: `99designs/keyring` (OS keyring), `hashicorp/go-retryablehttp` or custom backoff, `golang.org/x/net/idna` (punycode).

---

## 20) Performance & Concurrency

- **Parallel listing**: fetch next pages concurrently up to `--concurrency` (default 2) when `--paginate` is enabled and server exposes `Link` relations. Preserve output ordering.
- **Streaming**: large payloads (e.g., `emails send --raw`, `logs download`) use streaming IO, show progress if TTY.

---

## 21) Roadmap (post‑v1)

- **Contacts/Calendars** (CardDAV/CalDAV) CRUD when API stabilizes; alias-credential support in SDK.
- **Messages/Folders** abstractions (IMAP/POP3-like).
- **TUI** mode for quick triage.
- **Plugin system** (exec hooks) for custom post-processing of list outputs.

---

## 22) Risks & Mitigations

- **API drift**: pin to specific API version in User-Agent; add integration tests that watch for schema changes.
- **Rate limits / Quotas**: surface clear exit code 5 and actionable messages; exponential backoff honoring `Retry-After`.
- **Secrets leakage**: aggressive redaction, structured logging, `--redact=false` forbidden.
- **Windows path/encoding**: CI runners for Windows; normalize path handling.

---

## 23) Appendix A — Example Workflows

### Create domain with protections and bounce webhook

```
fe domains create example.com \
  --plan enhanced_protection \
  --protect-spam --protect-phishing --protect-virus \
  --recipient-verification \
  --bounce-webhook https://ops.example.com/hooks/forwardemail \
  --retention-days 30
```

### Create an alias that blackholes

```
fe aliases create example.com \
  --name noreply \
  --recipients blackhole@example.com \
  --is-enabled=false --error-code-if-disabled=250
```

### Send an email with structured fields

```
fe emails send --from notif@example.com --to user@dest.tld \
  --subject "Deploy done" --text "All green" --header "X-Env: prod"
```

### Download deliverability logs (guarding daily quota)

```
fe logs download --domain example.com --q "to:user@dest.tld"
```

---

## 24) Appendix B — Minimal SDK Snippet

```go
func NewClient(baseURL, username, password string, httpc *http.Client) (*Client, error) {
    if httpc == nil { httpc = defaultHTTPClient() }
    u, _ := url.Parse(baseURL)
    c := &Client{HTTP: httpc, BaseURL: u, auth: BasicAuth{username, password}}
    c.Domains = &DomainService{c}
    c.Aliases = &AliasService{c}
    c.Emails  = &EmailService{c}
    c.Account = &AccountService{c}
    c.Logs    = &LogService{c}
    c.Crypto  = &CryptoService{c}
    return c, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, v any) error {
    c.auth.Apply(req)
    req.Header.Set("Accept", "application/json")
    resp, err := c.HTTP.Do(req.WithContext(ctx))
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 { return decodeAPIError(resp) }
    if v != nil { return json.NewDecoder(resp.Body).Decode(v) }
    return nil
}
```

---

## 25) Acceptance Criteria (v1.0.0)

- All commands in §6 implemented; `--output json` round-trippable; pagination works against large lists.
- Retries/backoff working; integration tests prove 429/5xx handling.
- Secrets never shown in logs; config + keyring functional.
- Packages adhere to GoDoc; `fe --help` complete with examples.
- Installable via Homebrew on macOS and via binary tarballs for Linux/Windows
-
