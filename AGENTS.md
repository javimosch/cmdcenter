# cmdcenter - Agent Documentation

## Overview
cmdcenter is a generic command execution dashboard with configurable commands via JSON config file. It provides both a web UI and CLI for managing and executing shell commands.

## Agent Usage

### CLI Commands for Agents
Agents can use cmdcenter via CLI without needing the web UI:

```bash
# Initialize config file
cmdcenter init

# Add a command
cmdcenter add --id <id> --name "<name>" --command "<command>" --icon "<emoji>" --supports-args

# Edit a command
cmdcenter edit --id <id> --name "<name>" --command "<command>" --icon "<emoji>" --supports-args

# Remove a command
cmdcenter remove --id <id>

# List all commands
cmdcenter list

# Execute a command
cmdcenter run --id <id> [--args "<extra_args>"]

# Start/stop daemon
cmdcenter start -daemon
cmdcenter stop
cmdcenter status
```

### Config File Location
Config is stored at: `~/.cmdcenter/config.json`

### Config Structure
```json
{
  "title": "Command Center",
  "subtitle": "Generic command execution dashboard",
  "commands": [
    {
      "id": "command-id",
      "name": "Command Name",
      "description": "Command description",
      "icon": "🔧",
      "command": "shell command here",
      "supports_args": false
    }
  ]
}
```

### Key Features for Agents
- **No default commands**: Ships with empty config, agents must add commands
- **Argument support**: Commands can support extra arguments via `--supports-args` flag
- **CLI-first**: All operations available via CLI for automated agent workflows
- **Config management**: Full CRUD operations on commands via CLI
- **Execution**: Execute commands by ID with optional extra arguments

### Common Agent Workflows

1. **Setup commands for a specific task**:
```bash
cmdcenter add --id deploy --name "Deploy App" --command "./deploy.sh" --icon "🚀"
cmdcenter add --id backup --name "Backup DB" --command "pg_dump db > backup.sql" --icon "💾"
```

2. **Execute commands with arguments**:
```bash
# For commands with supports-args: true
cmdcenter run --id df --args "-h /"
```

3. **Manage configurations programmatically**:
```bash
# Update existing command
cmdcenter edit --id deploy --command "./deploy.sh --prod"

# Remove unused commands
cmdcenter remove --id old-command
```

### Daemon Mode
Agents can run cmdcenter in daemon mode for background operation:
```bash
cmdcenter start -daemon
cmdcenter stop
```

### Web UI
Web UI available at `http://localhost:3031` when daemon is running. Provides:
- Command execution buttons
- Command management (add/edit/delete)
- Raw config editing
- Log viewing and management
- Toast notifications (non-blocking)
- Custom confirmation dialogs

## Notes for Agents
- Always use `cmdcenter init` to create initial config
- Use unique IDs for commands to avoid conflicts
- The `supports-args` flag enables argument input modal in UI and CLI args support
- Config changes via CLI automatically update the running daemon
- Logs are stored at `/tmp/cmdcenter.log`
