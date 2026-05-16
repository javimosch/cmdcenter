# cmdcenter CLI Cheatsheet

## Overview
cmdcenter is a generic command execution dashboard with configurable commands via JSON config file. Use this skill for CLI-based command management and execution.

## Quick Start
```bash
# Initialize config file
cmdcenter init

# List all commands
cmdcenter list

# Add a command
cmdcenter add --id <id> --name "<name>" --command "<command>" --icon "<emoji>"

# Execute a command
cmdcenter run --id <id>

# Start web UI
cmdcenter start -daemon
```

## Command Reference

### Config Management
```bash
# Initialize config file with empty template
cmdcenter init

# Add a new command
cmdcenter add --id my-cmd --name "My Command" --command "echo hello" --icon "🔧"

# Add command with argument support
cmdcenter add --id df --name "Disk Free" --command "df" --icon "💾" --supports-args

# Edit existing command
cmdcenter edit --id my-cmd --command "echo hello world"

# Remove a command
cmdcenter remove --id my-cmd

# List all configured commands
cmdcenter list
```

### Command Execution
```bash
# Execute command by ID
cmdcenter run --id my-cmd

# Execute with extra arguments (if supports-args: true)
cmdcenter run --id df --args "-h /"
```

### Daemon Management
```bash
# Start HTTP server (UI) in foreground
cmdcenter start

# Start as daemon (background)
cmdcenter start -daemon

# Start on custom port
cmdcenter start -port 3000

# Stop daemon
cmdcenter stop

# Check daemon status
cmdcenter status

# Show version
cmdcenter version
```

## Config File
Location: `~/.cmdcenter/config.json`

Structure:
```json
{
  "title": "Command Center",
  "subtitle": "Generic command execution dashboard",
  "commands": [
    {
      "id": "unique-id",
      "name": "Display Name",
      "description": "What this command does",
      "icon": "🔧",
      "command": "shell command to execute",
      "supports_args": false
    }
  ]
}
```

## Common Patterns

### System Monitoring Commands
```bash
cmdcenter add --id disk-usage --name "Disk Usage" --command "df -h" --icon "💾"
cmdcenter add --id mem-usage --name "Memory Usage" --command "free -h" --icon "🧠"
cmdcenter add --id cpu-usage --name "CPU Usage" --command "top -bn1 | head -20" --icon "⚡"
```

### Deployment Commands
```bash
cmdcenter add --id deploy --name "Deploy" --command "./deploy.sh" --icon "🚀"
cmdcenter add --id rollback --name "Rollback" --command "./rollback.sh" --icon "↩️"
```

### File Operations
```bash
# With argument support for flexibility
cmdcenter add --id ls --name "List Files" --command "ls" --icon "📁" --supports-args
cmdcenter add --id grep --name "Search Files" --command "grep" --icon "🔍" --supports-args
```

## Web UI
When daemon is running, access UI at: `http://localhost:3031`

UI Features:
- Click command buttons to execute
- Manage commands via "Manage Commands" button
- Edit raw config JSON directly
- View and search daemon logs
- Non-blocking toast notifications
- Custom confirmation dialogs

## Tips
- Use unique IDs for commands (avoid spaces, use hyphens)
- Enable `--supports-args` for commands that need flexible arguments
- The web UI automatically reloads when config changes via CLI
- Daemon logs are stored at `/tmp/cmdcenter.log`
- Commands with `supports_args: true` show argument input modal in UI
