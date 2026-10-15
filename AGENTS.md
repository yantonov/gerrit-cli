# AGENTS.md

## Project
A Go CLI that reads Gerrit changes — diffs, commit messages, review comments —
and posts review comments.
Stack: Go 1.25, no framework, no database, no cache/queue.

## Starting a session
Work lives in `bl` tickets. One `backlog.md` (from `.tickets/`) is committed
into the repo.
- `bl available` — tickets you can claim (open, deps satisfied)
- `bl started` — tickets in progress
- `bl show <id> --json` — one ticket in full
- `bl help` — full command reference

## Commands
Every action is a `make` target:
- `setup` — pull dependencies
- `build` — compile to `target/gerrit-cli`
- `test` — run all tests
- `types` — run `go vet`
- `lint` — run `golangci-lint`
- `check` — full verification (lint → types → tests)

## Definition of Done
A feature is done only when:
- `make check` passes (including every test that existed before this session).
- Every acceptance criterion on the ticket passes.
- The ticket is closed with `bl done <id>`.

## Hard constraints (MUST / MUST NOT)
1. MUST: use HTTP Basic auth with credentials from the OS keychain for every Gerrit API call.
2. MUST: keep every credential in `GERRIT_URL` and the OS keychain. Never write them to a file.
3. MUST: when you change `skill/gerrit/`, update its reference docs at the same time.

## Work rules
- **WIP=1.** Claim one ticket at a time with `bl start <id>`. Close it with `bl done <id>` before you claim another.
- Stay inside the claimed ticket. If you find another problem, file it with `bl add "<title>"`. Then continue.
- Priority order: correctness → performance → style.
- Let the linter own style. Run `make lint` instead of listing style rules.

## Escalate to human (stop and ask)
- Adding a new external dependency (`go get`).
- Changing the Gerrit API surface this CLI exposes.
- Changing the keychain storage format or credential flow.
- Deleting anything that cannot be recovered.

## Details (read when needed)
- `skill/gerrit/reference/setup.md`: setup and keychain troubleshooting
