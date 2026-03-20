# lm-cli — LINE Messaging API CLI

A command-line interface for the LINE Messaging API. Enables DevOps engineers, bot developers, and marketers to manage LINE Official Accounts from the terminal.

## Features

- **3 authentication modes**: Long-term token, stateless (CI/CD-friendly), JWT v2
- **All message types**: push, multicast, broadcast, narrowcast, reply
- **Output formats**: table / json / yaml / csv
- **CI/CD-ready**: `--no-input`, `LM_TOKEN` env var, meaningful exit codes
- **Multi-profile support**: Manage prod/staging/test environments

## Installation

```bash
# Via Go
go install github.com/crowdy/lm-cli@latest

# Binary (GitHub Releases)
# https://github.com/crowdy/lm-cli/releases
```

## Quick Start

```bash
# Authenticate (long-term token)
lm auth login

# Check bot info
lm bot info

# Send a message
lm message push <userId> "Hello!"

# JSON output
lm bot info --format json

# CI/CD with env var
LM_TOKEN=xxx lm message push <userId> "Deploy complete"
```

## Commands

| Command | Description |
|---------|-------------|
| `lm auth login` | Configure authentication |
| `lm auth status` | Show authentication status |
| `lm auth list` | List profiles |
| `lm auth switch <profile>` | Switch active profile |
| `lm message push <userId> <text>` | Push message to user |
| `lm message multicast <text> --to id,id` | Send to multiple users |
| `lm message broadcast <text>` | Broadcast to all followers |
| `lm message reply <replyToken> <text>` | Reply to webhook event |
| `lm bot info` | Get bot profile |
| `lm bot quota` | Get message quota |
| `lm bot consumption` | Get usage consumption |
| `lm user profile <userId>` | Get user profile |
| `lm user followers` | List follower IDs |
| `lm group info <groupId>` | Get group info |
| `lm group members <groupId>` | List group members |
| `lm group leave <groupId>` | Leave group |
| `lm richmenu list` | List rich menus |
| `lm richmenu create --file menu.json` | Create rich menu |
| `lm richmenu upload <id> image.jpg` | Upload image |
| `lm webhook get` | Get webhook URL |
| `lm webhook set <url>` | Set webhook URL |
| `lm webhook test` | Test webhook |
| `lm audience list` | List audience groups |
| `lm insight followers` | Follower statistics |
| `lm content get <messageId>` | Download content |

## Configuration

Config directory: `~/.config/lm/` (override with `LM_CONFIG_DIR`)

### Environment Variables

| Variable | Description |
|----------|-------------|
| `LM_TOKEN` | Direct token (bypasses auth) |
| `LM_CHANNEL_ID` | Channel ID |
| `LM_CHANNEL_SECRET` | Channel secret |
| `LM_PROFILE` | Profile name |
| `LM_FORMAT` | Output format (table/json/yaml/csv) |
| `LM_CONFIG_DIR` | Config directory path |
| `LM_NO_INPUT` | Non-interactive mode (1 or true) |
| `LM_ENDPOINT` | Override API base URL (testing) |
| `LM_DEBUG` | Debug logging (1, true, api) |

## Authentication

### Long-term Token (Simplest)
```bash
lm auth login
# Enter Channel ID and long-term token
```

### Stateless Token (CI/CD Recommended)
Auto-issues and refreshes tokens from Channel ID + Secret.
```bash
lm auth login --type stateless
# Enter Channel ID and Channel Secret
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Auth error (401/403) |
| 3 | Not found (404) |
| 4 | Validation error |
| 5 | LINE API error |
| 6 | Network error |
| 10 | Cancelled (Ctrl+C) |
| 11 | Rate limit (429) |

## Development

```bash
make build    # Build ./lm
make test     # go test ./...
make lint     # golangci-lint run ./...
make coverage # Generate coverage report
```

## License

MIT
