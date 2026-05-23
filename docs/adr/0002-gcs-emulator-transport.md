# ADR-0002: GCS emulator HTTP transport for integration tests

## Status

Accepted

## Context

Integration tests for GCS use [fake-gcs-server](https://github.com/fsouza/fake-gcs-server) as a local emulator. The Go GCS SDK (`cloud.google.com/go/storage`) uses **two separate HTTP APIs**:

- **JSON API** — metadata operations (`Attrs`, `List`, `Delete`, uploads). Redirected to a custom host via `option.WithEndpoint("http://localhost:4443/storage/v1/")`.
- **XML API** — streaming downloads (`obj.NewReader`). The SDK constructs the URL as `http://<endpoint-host>/<bucket>/<object>`, which fake-gcs-server does **not** serve. It serves downloads at `/download/storage/v1/b/<bucket>/o/<object>?alt=media`.

Simply setting `option.WithEndpoint` is not enough: JSON API calls succeed but `NewReader` returns "object doesn't exist" because the XML API path is wrong.

## Decision

Inject a custom `http.Client` via `option.WithHTTPClient` when a custom endpoint is configured with `application_default` (emulator) auth. The client uses `gcsEmulatorTransport`, which:

1. Rewrites the `Host` and `Scheme` of every outbound request to the emulator address.
2. Detects XML API paths (no `/storage/`, `/upload/`, or `/download/` prefix) and rewrites them to the fake-gcs download path: `/download/storage/v1/b/<bucket>/o/<object>?alt=media`.

This approach is confined to `gcs.go` and only activates when `Endpoint` is set with `application_default` or empty auth — it has no effect on real GCS usage (explicit credentials bypass the rewrite path entirely).

## Alternatives considered

- **`STORAGE_EMULATOR_HOST` env var** — the Go GCS SDK v1.x honours this for the XML API only; JSON API calls (`Attrs`, `List`) still go to googleapis.com. Rejected because it leaves half the operations broken.
- **Patch `NewReader` to use `mediaLink` from `Attrs`** — would require `Read()` to make two round-trips (Attrs then download). Rejected as more invasive than the transport approach.
- **Switch to fake-gcs XML-only mode** — fake-gcs requires the full JSON API for writes; an XML-only setup would break `Write`. Rejected.

## Consequences

- `gcs.go` gains a small private type (`gcsEmulatorTransport`) that is never instantiated in production use.
- `GCS_ENDPOINT` passed to integration tests must include the `/storage/v1/` suffix (e.g. `http://localhost:4443/storage/v1/`) so `option.WithEndpoint` routes JSON calls correctly; the transport then handles the XML path rewrite.
- Real GCS users are unaffected: explicit auth types (`service_account`, `access_token`, `impersonate`) do not trigger `WithoutAuthentication` or the custom transport.
