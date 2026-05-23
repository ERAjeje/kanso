# Plan: fix-couchdb-jwt-auth

## Problem
PouchDB ↔ CouchDB sync is broken because CouchDB has no `[jwt_keys]` configuration to validate the Bearer JWT tokens sent by PouchDB. All sync attempts fail with 401 (unauthenticated), keeping data only in local IndexedDB.

Logs confirm: no `_changes` replication events, only `PUT /{db} 412` (benign — DB creation retry).

## Root Cause
CouchDB Docker container runs stock config — no `local.ini` mounted with JWT auth settings. The architecture documented in ARCHITECTURE.md requires:
1. `[chttpd]` with `jwt_authentication_handler`
2. `[jwt_keys]` with base64-encoded HMAC secret matching `JWT_SECRET`
3. `_security` docs on each DB (since `COUCHDB_USER` activates `require_valid_user=true`)

None of these were implemented.

## Changes

### 1. CREATE `infra/couchdb/local.ini`
CouchDB JWT authentication config with base64-encoded `JWT_SECRET`.

### 2. EDIT `infra/docker-compose.yml`
Mount `./couchdb/local.ini:/opt/couchdb/etc/local.d/10-jwt-auth.ini`

### 3. EDIT `backend/cmd/kanso-api/main.go`
Add `_security` doc PUT to `ensureCouchDBDatabases()` so authenticated users (via JWT) can access their databases.

## Verification
1. `docker compose up -d` — CouchDB should start with JWT config
2. `curl -s -H "Authorization: Bearer $(curl -s -X POST ...)" http://localhost:5984/registros/_all_docs` → should return data, not 401
3. Browser console should show sync state transitions (offline → syncing → online)
4. `go build ./...` + `go vet ./...` + `go test ./...` — must pass
5. `npx tsc --noEmit` + `npx vitest run` — must pass
