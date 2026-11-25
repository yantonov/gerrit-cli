# Setup & Troubleshooting

## Install

Install Gerrit CLI from the [repo](https://github.com/yantonov/gerrit-cli).

## Configure

```bash
export GERRIT_URL=your-gerrit-instance.com

gerrit-cli keychain set username   # hidden prompt
gerrit-cli keychain set password   # hidden prompt
```

## Check keychain status

```bash
gerrit-cli keychain status
```

Example output when fully configured:

```
Host: your-gerrit-instance.com
Username: set
Password: set
```

If `Username` or `Password` shows `not set`, re-run the corresponding `keychain set` command above.

## Check that the credentials work

`keychain status` reports only what is stored. To find out whether Gerrit accepts it, run:

```bash
gerrit-cli keychain check
```

Example output when the credentials are accepted:

```
[OK] login succeeded for host your-gerrit-instance.com as Your Name
```

If a value is missing, the command stops before contacting Gerrit and names the missing fields. If Gerrit rejects the credentials (HTTP 401), it reports that the stored password is stale and where to get a new one.

## Troubleshooting

- **`Host` is wrong or missing** — `GERRIT_URL` isn't exported in the current shell; re-run `export GERRIT_URL=...`.
- **Auth errors on commands (`detail`, `files`, etc.)** — run `gerrit-cli keychain check`. It reports the missing field, or that Gerrit rejects the stored credentials, so there is no need to guess by re-setting them. Re-set only what it names, e.g. after a password rotation.
- **`resolve-change-number` / `resolve-change-id` fail on a valid URL** — confirm the URL's host matches `GERRIT_URL`; the CLI won't resolve URLs from a different Gerrit instance.
- **A `--change-id`/`--review-url` command errors with "requires --change-id or --review-url" or "specify either ... not both"** — exactly one of the two flags must be passed; check for a missing or duplicated flag.
