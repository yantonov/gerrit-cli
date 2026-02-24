# Gerrit CLI

A Go-based command-line tool for interacting with Gerrit REST API.  
The intent was to use it as a skill for claude code, etc.  
Intentionally only read access is supported now.  

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

3. Set the authentication environment variable:
```bash
export SECRET_GERRIT_AUTH_TOKEN=$(echo -n 'username:password' | base64)
```

## Build

```bash
go build -o gerrit-cli
```

## Usage

```bash
./gerrit-cli <command> [args...]
```

### Available Commands

- `get-change <change_id>` - Get detailed change information
- `get-files <change_id>` - Get list of files in a change
- `get-commit <change_id>` - Get commit message
- `get-diff <change_id> <file_path>` - Get file diff
- `get-messages <change_id>` - Get review messages
- `get-patch <change_id>` - Get full patch
- `get-moab-numbers <change_id>` - Extract MOAB numbers from review messages
- `resolve-url <url>` - Resolve Gerrit URL to change ID

### Examples

```bash
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
  "CSHARP": 123,
  "JAVA": 234
}
```

# Resolve Gerrit URL to change ID
./gerrit-cli resolve-url https://your-gerrit-instance.com/c/namespace/project/+/1234567
```
