# jarakey-shared-middleware — Agent & Reviewer Contract

This file tells any AI assistant (the `@claude` PR reviewer wired via
`.github/workflows/pr-linter.yml`, or an editor agent) how to work in this repository. It is
loaded automatically from the repo root. If it isn't picked up, start with:
_"Follow the rules in AGENT.md."_

---

## What this repo is

A **shared Go library** (`APP_TYPE: library`) imported by **every** Jarakey backend service
(gateway, user, code, property, billing, notification, files, tracing). It provides the
standard middleware stack: auth/JWT, CORS, security headers, correlation IDs, Prometheus
metrics, circuit breaker, retry, health check.

Packages: `middleware/` (the HTTP middleware), `utils/` (JWT manager and helpers),
`types/` (shared claims/user/role types).

## The prime directive — backward compatibility

Because this library is imported by every service, **any breaking change ripples to the whole
platform**. Treat backward compatibility as the top constraint:

- **Never** change an exported function signature, its behaviour, or a response shape without
  a migration path. Add a new function rather than repurpose an existing one.
- Middleware that sets response headers, context keys (`c.Set("user_id", …)`, `user_role`,
  `user_email`, `org_id`, `user_permissions`), or status codes must keep those keys/values
  stable — downstream handlers read them by name.
- Config is read from the environment. New behaviour must default to the **current** behaviour
  when the new env var is unset.
- A version bump follows semver: behaviour-preserving fix = patch, additive = minor,
  breaking = major (and call it out loudly in the PR).

## Security invariants (do not regress)

1. **CORS never pairs a wildcard origin with credentials.** `Access-Control-Allow-Origin: *`
   together with `Access-Control-Allow-Credentials: true` is spec-invalid (browsers drop it)
   and unsafe. Reflect the request `Origin` (+ `Vary: Origin`) when present; only use `*`
   for origin-less callers and then WITHOUT credentials. See `middleware/CORS`.
2. **`Access-Control-Allow-Methods` lists every method the platform exposes — including
   `PATCH`.** A missing method silently breaks that method's browser preflight (QA S8).
3. **JWT secret**: `resolveJWTSecret()` warns when it falls back to the built-in development
   default. That default is public and forgeable — it must never be used in a deployed
   environment. Do not remove the warning; do not add new hardcoded secrets.
4. **Security headers** (`nosniff`, `DENY`, HSTS, `Referrer-Policy`) are set by
   `SecurityHeaders()`; keep them.
5. Auth failures return `401` (unauthenticated) / `403` (authenticated but unauthorized) with
   `{ error, message }`. Never leak token contents or stack traces in the body.

## Performance

- Resolve config and build reusable objects **once at middleware setup**, not per request —
  e.g. the JWT manager is built in `AuthRequired()` before the handler closure, not inside it.
  Avoid per-request `os.Getenv` and allocations on the hot path.

## Tests

- Every exported behaviour has a table-driven test in the same package. CI enforces
  `COVERAGE_THRESHOLD: 80` on Go `1.23` (`.github/workflows/ci.yml`).
- Change a header/status/context-key contract → update its assertion in the same PR.
- Concurrency tests synchronise via channels — no `time.Sleep`, no `require.Eventually`
  (`go test -race` must stay clean).

## PR review checklist (for the @claude reviewer)

Read `gh pr diff`, then check, in priority order:

1. **Backward compatibility** — any exported signature/behaviour/header/context-key change?
   Flag it and demand a migration path or major-version call-out.
2. **Security** — the five invariants above; injection or auth bypass; secrets in code.
3. **Correctness** — logic errors, unhandled errors, nil derefs, wrong status codes.
4. **Tests** — new/changed behaviour covered; coverage not dropping below 80%; race-clean.
5. **Performance** — per-request work that could be hoisted to setup.

Be concise and actionable. Prioritise critical issues. If a prior `@claude` comment exists,
only check whether it was addressed — don't repeat it.
