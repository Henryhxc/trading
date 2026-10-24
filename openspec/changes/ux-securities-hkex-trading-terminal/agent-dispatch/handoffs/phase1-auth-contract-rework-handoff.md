# Phase 1 Auth Contract Rework Handoff

Date: 2026-10-24 (Asia/Hong_Kong)
Status: `READY_FOR_INDEPENDENT_REVIEW`

## Scope

This bounded rework addresses the three findings in `agent-dispatch/reviews/phase1-review.md`:

- Login now maps only `sql.ErrNoRows` (missing identity) to `ErrCredentials`; unexpected repository errors propagate to the HTTP layer and become the existing generic 500 response.
- Registration and login share an OpenAPI `Password` schema documenting the server's 12-72 UTF-8 byte policy, bcrypt's 72-byte limit, and the fact that JSON Schema `minLength`/`maxLength` count Unicode characters rather than UTF-8 bytes.
- Register and login now list the implemented 500 `InternalServerError` response in OpenAPI.

No product decision 0.4, task checkbox, phase status, orchestration state, Phase 2 work, database migration, frontend source, or runtime service was changed.

## Files Changed

- `apps/backend/internal/service/auth/service.go`: add explicit 12-72 byte password constants and distinguish `sql.ErrNoRows` from unexpected repository failures.
- `apps/backend/internal/service/auth/service_test.go`: add service behavior assertions for missing identities and preserved storage errors.
- `apps/backend/internal/httpapi/server_test.go`: add register/login generic-500 assertions and preserve invalid-identity 401 assertion.
- `docs/api/openapi.yaml`: reference the shared password schema, add the 500 responses, and define the byte/character semantics.
- `tests/phase1_rework_contract_test.py`: add durable OpenAPI/password/500 contract guards.

## RED Evidence

Tests were added before the production change and run against the pre-fix implementation.

```text
cd apps/backend
go test ./internal/service/auth ./internal/httpapi -run 'Test(LoginMapsMissingIdentityToInvalidCredentials|LoginPreservesUnexpectedRepositoryError|AuthUnexpectedErrorsReturnGenericInternalServerError|LoginInvalidIdentityReturnsUnauthorized)' -count=1
FAIL: TestLoginPreservesUnexpectedRepositoryError
      Login() error = invalid credentials, want repository error database connection reset
```

The missing-identity, HTTP 500, and HTTP 401 assertions were otherwise able to execute; the failure specifically demonstrated that every repository lookup error was incorrectly converted to credentials failure.

```text
python3 tests/phase1_rework_contract_test.py
ERROR: test_openapi_documents_bcrypt_password_bytes_and_auth_server_errors
      Password schema was absent from docs/api/openapi.yaml
```

The initial contract test invocation used the script entry point; no harness/import failure was used as RED evidence.

## GREEN Evidence

Focused tests after the minimal fix:

```text
cd apps/backend && go test ./internal/service/auth ./internal/httpapi -count=1
ok   ux-securities/backend/internal/service/auth
ok   ux-securities/backend/internal/httpapi

python3 tests/phase1_rework_contract_test.py
Ran 4 tests ... OK
```

Full and contract checks:

```text
cd apps/backend && go test ./... -count=1 && go vet ./...
PASS (all Go packages; no vet findings)

python3 tests/phase1_source_guard_test.py
Ran 14 tests ... OK

python3 tests/phase1_rework_contract_test.py
Ran 4 tests ... OK

openspec validate ux-securities-hkex-trading-terminal --strict
Change 'ux-securities-hkex-trading-terminal' is valid
```

`docs/api/openapi.yaml` also parsed successfully with PyYAML. `git diff --check` could not be claimed because this workspace has no Git metadata. The frontend was not affected by this backend/OpenAPI-only change, so no frontend build was required in this lane.

## API and Password Semantics

- Register: password must be 12-72 UTF-8 bytes; service rejects outside this range with `ErrInvalidInput`, mapped to HTTP 400.
- Login: bcrypt comparison remains the credential check; missing identity (`sql.ErrNoRows`) and password mismatch map to `ErrCredentials`/HTTP 401. A password beyond bcrypt's 72-byte input limit therefore remains a safe invalid-credentials response.
- Repository failures other than `sql.ErrNoRows` propagate from `Login` and map to `{code: "internal_error", ...}` with HTTP 500. Storage details are not exposed.
- OpenAPI uses `minLength: 12` and `maxLength: 72` as portable character-level schema guards, while its description explicitly requires server-side UTF-8 byte validation for non-ASCII passwords.

## Residual Risks

- The OpenAPI byte policy is documented text plus a schema guard because standard JSON Schema length keywords are character-based; clients must still handle server validation for multi-byte characters.
- Runtime browser, Compose, and phpMyAdmin evidence remain the responsibility of the main Phase 1 gate. This handoff does not claim those checks or authorize phase acceptance.
- No real broker, production database, or external provider was contacted.
