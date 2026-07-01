# Gerrit CLI

A Go-based command-line tool for interacting with Gerrit REST API.  
The intent was to use it as a skill for claude code, etc.  
Most commands are read-only, with a small write command for publishing review comments.

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

Credentials are stored per Gerrit host, so switching `GERRIT_URL` to a different instance uses a different set of stored credentials. Use `gerrit-cli keychain status` to check what's currently configured, `gerrit-cli keychain remove <username|password>` to drop a single value, or `gerrit-cli keychain clear` to drop both.

## Build

```bash
go build -o gerrit-cli
```

## Usage

```bash
./gerrit-cli <command> [args...]
```

### Available Commands

- `keychain set username` - Store the username for the current `GERRIT_URL` host (read from an interactive, hidden prompt)
- `keychain set password` - Store the password for the current `GERRIT_URL` host (read from an interactive, hidden prompt)
- `keychain remove <username|password>` - Remove a single stored value for the current host
- `keychain clear` - Remove both stored values for the current host
- `keychain status` - Show whether username/password are set for the current host (values are never printed)
- `get-change <change_id>` - Get detailed change information
- `get-files <change_id>` - Get list of files in a change
- `get-commit <change_id>` - Get commit message
- `get-diff <change_id> <file_path>` - Get file diff
- `get-messages <change_id>` - Get review messages
- `get-patch <change_id>` - Get full patch
- `get-moab-numbers <change_id>` - Extract MOAB numbers from review messages
- `get-publish-version <change_id>` - Extract published artifact versions from review messages
- `post-comment <change_id> <comment>` - Publish a top-level review comment
- `resolve-change-number <url>` - Extract the change number from a Gerrit URL
- `resolve-change-id <url>` - Resolve Gerrit URL to commit Change-Id via Gerrit API

### Examples

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

# Get change details
./gerrit-cli get-change I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

# Get files in a change
./gerrit-cli get-files I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

# Get commit message
./gerrit-cli get-commit I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

# Get diff for a specific file
./gerrit-cli get-diff I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3 src/main.go

# Get review messages
./gerrit-cli get-messages I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

# Get full patch
./gerrit-cli get-patch I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

# Get MOAB numbers
./gerrit-cli get-moab-numbers I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

Example:
```
{
  "CSHARP": "123",
  "JAVA": "234"
}
```

# Get published artifact versions
./gerrit-cli get-publish-version I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3

Example:
```
{
  "CSHARP": "1.1948302.1.44265-review",
  "JAVA": "2.5550123.1.999-review"
}
```

# Publish a top-level review comment
./gerrit-cli post-comment I3ea8ccae945a1a1a0c52aab84bb1d2c1830bb2e3 "Looks good to me"

# Extract change number from Gerrit URL
./gerrit-cli resolve-change-number https://your-gerrit-instance.com/c/namespace/project/+/1234567

# Resolve Gerrit URL to commit Change-Id via Gerrit API
./gerrit-cli resolve-change-id https://your-gerrit-instance.com/c/namespace/project/+/1234567
```
