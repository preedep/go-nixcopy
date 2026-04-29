# ADR-0001: SFTP Auth Unit Test Strategy

**Status:** Accepted  
**Date:** 2026-04-29

## Context

`SFTPStorage.Connect` builds SSH auth methods from config: password, private key (with optional passphrase), or both. This logic had zero unit-test coverage — only the integration test exercised it, and only via password auth.

We needed to add unit tests covering all auth paths without requiring a live SSH server.

Two approaches were considered:

**Option A — Extract helper + in-memory keys**  
Pull the auth-method building into an unexported `buildAuthMethods(cfg)` function. Test it directly from a `package storage` internal test file using RSA keys generated in-memory with `crypto/rsa`.

**Option B — In-process SSH server**  
Spin up a real SSH server inside the test using `golang.org/x/crypto/ssh` server APIs. Test the full `Connect` dial against it.

## Decision

**Option A.** `buildAuthMethods` is pure config-wiring — it reads a key file, parses it, and returns `[]ssh.AuthMethod`. There is no network behavior to verify at this layer. Testing it against a live server would be testing `golang.org/x/crypto/ssh` library internals, not our code.

The actual SSH handshake ("does the server accept our key?") belongs in the integration test, where a real SFTP daemon is running. We extend the integration test to support `SFTP_PRIVATE_KEY_PATH` when that is needed.

## Consequences

- Unit tests run in milliseconds with no ports, no race conditions from network, no new dependencies.
- Failures point directly at auth-selection logic.
- The SSH handshake itself is not covered by unit tests — accepted, because it is covered by the integration test path and the `golang.org/x/crypto/ssh` library is not ours to test.
- Future contributors proposing an in-process SSH server should refer to this ADR before adding the complexity.
