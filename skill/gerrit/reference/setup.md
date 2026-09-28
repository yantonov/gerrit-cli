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
Host: review.crto.in
Username: set
Password: set
```

If `Username` or `Password` shows `not set`, re-run the corresponding `keychain set` command above.

## Troubleshooting

- **`Host` is wrong or missing** — `GERRIT_URL` isn't exported in the current shell; re-run `export GERRIT_URL=...`.
- **Auth errors on commands (`get-change`, etc.)** — run `gerrit-cli keychain status` to confirm credentials are set, then re-set them if stale (e.g. after a password rotation).
- **`resolve-change-number` / `resolve-change-id` fail on a valid URL** — confirm the URL's host matches `GERRIT_URL`; the CLI won't resolve URLs from a different Gerrit instance.
- **A `--change-id`/`--review-url` command errors with "requires --change-id or --review-url" or "specify either ... not both"** — exactly one of the two flags must be passed; check for a missing or duplicated flag.
