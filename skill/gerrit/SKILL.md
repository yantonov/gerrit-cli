---
name: gerrit
description: Use when looking up a Gerrit change (by change ID or URL, e.g. review.crto.in/c/...) - viewing diffs, commit messages, review comments, or searching changes by query. Read-only.
---

# Gerrit CLI

Read-only inspection and search of Gerrit changes via `gerrit-cli`.

Setup, keychain configuration, and troubleshooting: [reference/setup.md](reference/setup.md).

## Change identification

Every `<change_id>`-based command below requires exactly one of `--change-id` or `--review-url` (never both):

- `--change-id <id>` - the Gerrit change ID, a numeric change number, or a full `project~branch~Change-Id` triplet
- `--review-url <url>` - a Gerrit change URL; the change number is resolved from it automatically

## Commands

- `get-change (--change-id <id> | --review-url <url>)` - change details
- `get-files (--change-id <id> | --review-url <url>)` - files in a change
- `get-commit (--change-id <id> | --review-url <url>)` - commit message
- `get-diff (--change-id <id> | --review-url <url>) <file_path>` - diff for a file
- `get-messages (--change-id <id> | --review-url <url>)` - review messages
- `get-patch (--change-id <id> | --review-url <url>)` - full patch
- `get-moab-numbers (--change-id <id> | --review-url <url>)` - MOAB numbers from review messages
- `get-publish-version (--change-id <id> | --review-url <url>)` - published artifact versions from review messages
- `is-verified (--change-id <id> | --review-url <url>)` - whether the Verified label is set on the latest patch set
- `resolve-change-number <url>` - change number from a Gerrit URL
- `query-changes --query <query>` - search changes by a Gerrit query string (e.g. a JIRA key, project+branch filters, free text)
- `resolve-change-id <url>` - resolve a Gerrit URL to its commit Change-Id
- `shell <show|install>` - print or install bash completion
- `--version` - print the CLI version, Go version, and platform

## Examples

```bash
gerrit-cli get-change --change-id I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3
gerrit-cli get-diff --change-id I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3 src/main.go
gerrit-cli get-messages --review-url https://your-gerrit-instance.com/c/namespace/project/+/1234567
gerrit-cli resolve-change-number https://your-gerrit-instance.com/c/namespace/project/+/1234567
gerrit-cli query-changes --query "QUEUE-123"
gerrit-cli query-changes --query "project:my-project+branch:main+status:open"
```
