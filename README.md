# Gerrit CLI

A Go-based command-line tool for interacting with Gerrit REST API.  
The intent was to use it as a skill for claude code, etc.  
Most commands are read-only, with a small write command for publishing review comments.

## Table of Contents

- [Setup](#setup)
- [Build](#build)
- [Usage](#usage)
  - [keychain](#keychain)
  - get-change
  - get-files
  - get-commit
  - get-diff
  - get-messages
  - get-patch
  - [get-moab-numbers](#get-moab-numbers)
  - [get-publish-version](#get-publish-version)
  - [is-verified](#is-verified)
  - post-comment
  - resolve-change-number
  - resolve-change-id
  - [shell](#shell)
  - [--version](#--version)

## Setup

1. Set your Gerrit base URL (flexible format):
```bash
# Simple hostname (https:// will be added automatically)
export GERRIT_URL=your-gerrit-instance.com

# Or with explicit protocol
export GERRIT_URL=https://your-gerrit-instance.com

# HTTP is also supported
export GERRIT_URL=http://your-gerrit-instance.com
```

2. Get your HTTP credentials from: `https://your-gerrit-instance.com/settings/#HTTPCredentials`

3. Store your credentials in the OS keychain, scoped to the host from `GERRIT_URL`. Both values are read from an interactive, hidden prompt (input is not echoed):
```bash
gerrit-cli keychain set username
# Username: <hidden prompt>

gerrit-cli keychain set password
# Password: <hidden prompt>
```

Credentials are stored per Gerrit host, so switching `GERRIT_URL` to a different instance uses a different set of stored credentials. See [keychain](#keychain) below for the full set of `keychain` subcommands.

The keychain is the macOS Keychain, the Windows Credential Manager, or the freedesktop Secret Service on Linux. If none is reachable, every command that touches credentials fails with a message naming the platform and how to get a working store - on Linux that usually means starting a keyring daemon such as gnome-keyring, and running under a D-Bus session (`dbus-run-session -- gerrit-cli <command>`) when connected over SSH.

## Build

```bash
sh bin/build.sh        # -> target/gerrit-cli, with the version stamped from `git describe`
```

## Usage

```bash
./gerrit-cli <command> [args...]
```

Every `<change_id>`-based command below requires exactly one of `--change-id` (the Gerrit change ID, numeric change number, or full `project~branch~Change-Id` triplet) or `--review-url` (a Gerrit change URL, from which the change number is resolved automatically). Passing neither, or both, is an error.

### keychain

Manage the credentials used to authenticate against the host from `GERRIT_URL`.

- `keychain set username` - Store the username for the current `GERRIT_URL` host (read from an interactive, hidden prompt)
- `keychain set password` - Store the password for the current `GERRIT_URL` host (read from an interactive, hidden prompt)
- `keychain remove <username|password>` - Remove a single stored value for the current host
- `keychain clear` - Remove both stored values for the current host
- `keychain status` - Show whether username/password are set for the current host (values are never printed)

```bash
# Store credentials for the current GERRIT_URL host
./gerrit-cli keychain set username
./gerrit-cli keychain set password

# Check what's currently stored
./gerrit-cli keychain status

# Remove a single stored value
./gerrit-cli keychain remove password

# Remove both stored values
./gerrit-cli keychain clear
```

### get-moab-numbers

`get-moab-numbers (--change-id <change_id> | --review-url <url>)` - Extract MOAB numbers from review messages

```bash
./gerrit-cli get-moab-numbers --change-id I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3
```

Example output:
```
{
  "CSHARP": "123",
  "JAVA": "234"
}
```

### get-publish-version

`get-publish-version (--change-id <change_id> | --review-url <url>)` - Extract published artifact versions from review messages

```bash
./gerrit-cli get-publish-version --change-id I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3
```

Example output:
```
{
  "CSHARP": "1.1948302.1.44265-review",
  "JAVA": "2.5550123.1.999-review"
}
```

### is-verified

`is-verified (--change-id <change_id> | --review-url <url>)` - Check whether the Verified label is set on the latest patch set

```bash
./gerrit-cli is-verified --change-id I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3
```

Example output:
```
{
  "by": "svc-moab2gerrit",
  "status": "verified",
  "verified": true
}
```

### shell

Print or install bash completion for `gerrit-cli`.

- `shell show` - Print the bash completion script for `gerrit-cli`
- `shell install` - Install the bash completion script to where bash loads it from automatically

```bash
# Print the bash completion script (review before trusting it, or source it directly)
. <(./gerrit-cli shell show)

# Install the completion script to wherever bash will load it from automatically
./gerrit-cli shell install
```

Shell completion is only supported for bash, detected through `$SHELL`. `shell install` writes to
`~/.local/share/bash-completion/completions/gerrit-cli`, or, under Git for Windows' bundled MSYS bash
(detected through `$MSYSTEM`), to `~/bash_completion.d/gerrit-cli.bash`, since that is the only path
MSYS bash's own `git-prompt.sh` auto-sources.

### --version

`--version` (or `version`) - Print the version this binary was built from, plus the Go version and platform

```bash
./gerrit-cli --version
```
