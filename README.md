# cmdcenter

Generic command execution dashboard with configurable commands via JSON config file. Provides both a modern web UI and CLI for managing and executing shell commands.

## Features

- 🎯 **Generic Command Center**: Execute any shell commands from a web UI
- ⚙️ **Configurable Commands**: Define commands in `~/.cmdcenter/config.json`
- 🔧 **CLI Management**: Full CRUD operations via command-line interface
- 📋 **Command Logging**: All command executions logged with timestamps and status
- 🔍 **Search Logs**: Search through daemon logs with filtering
- 🗑️ **Log Management**: Clear logs when needed
- 🎨 **Modern UI**: Argentine flag color scheme with bento-grid layout
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile
- 🚀 **Daemon Mode**: Run as background service
- 💬 **Non-blocking Toasts**: Custom toast notifications instead of native alerts
- 🔄 **Argument Support**: Commands can accept extra arguments
- 🤖 **Agent-first**: Designed for AI/agent workflows with comprehensive CLI

## Installation

### From Source
```bash
git clone https://github.com/yourusername/cmdcenter.git
cd cmdcenter
go build -o cmdcenter
./cmdcenter init
```

### Via Supercli
```bash
sc install cmdcenter
```

## Quick Start

### Initialize Config
```bash
cmdcenter init
```

This creates `~/.cmdcenter/config.json` with an empty commands array.

### Add Commands
```bash
# Add a simple command
cmdcenter add --id status --name "System Status" --command "uptime" --icon "📊"

# Add command with argument support
cmdcenter add --id df --name "Disk Free" --command "df" --icon "💾" --supports-args
```

### Start Web UI
```bash
# Start in foreground
cmdcenter start

# Start as daemon
cmdcenter start -daemon

# Start on custom port
cmdcenter start -port 3000
```

Access the UI at `http://localhost:3031`

### Execute Commands
```bash
# Via CLI
cmdcenter run --id status

# With arguments (if supports-args enabled)
cmdcenter run --id df --args "-h /"
```

## CLI Usage

### Config Commands
```bash
cmdcenter init                          # Initialize config file
cmdcenter add --id <id> --name <name> --command <cmd> [--icon <emoji>] [--supports-args]
cmdcenter edit --id <id> [--name <name>] [--command <cmd>] [--icon <emoji>] [--supports-args]
cmdcenter remove --id <id>
cmdcenter list                            # List all commands
```

### Execution Commands
```bash
cmdcenter run --id <id> [--args <args>]  # Execute a command
```

### Server Commands
```bash
cmdcenter start [-port <port>] [-daemon]  # Start HTTP server
cmdcenter stop                            # Stop daemon
cmdcenter status                          # Check daemon status
cmdcenter version                         # Show version
```

## Configuration

Config file location: `~/.cmdcenter/config.json`

### Example Config
```json
{
  "title": "Command Center",
  "subtitle": "Generic command execution dashboard",
  "commands": [
    {
      "id": "status",
      "name": "System Status",
      "description": "Check system status",
      "icon": "📊",
      "command": "uptime",
      "supports_args": false
    },
    {
      "id": "df",
      "name": "Disk Free",
      "description": "Show disk free space",
      "icon": "💾",
      "command": "df",
      "supports_args": true
    }
  ]
}
```

### Command Fields
- `id`: Unique identifier (required)
- `name`: Display name (required)
- `command`: Shell command to execute (required)
- `description`: Command description
- `icon`: Emoji icon
- `supports_args`: Enable argument input modal/CLI args support

## Web UI Features

### Command Execution
- Click command buttons to execute
- Commands with `supports_args: true` show argument input modal
- Live command output display
- Success/error feedback with custom toasts

### Command Management
- Add/Edit/Delete commands via UI
- Raw JSON config editing
- Live config reload
- Form validation

### Log Management
- View daemon logs in real-time
- Search/filter logs
- Auto-refresh every 10 seconds
- Clear logs on demand

### UI Features
- Non-blocking toast notifications
- Custom confirmation dialogs
- Responsive bento-grid layout
- Argentine flag color scheme
- Mobile-friendly design

## Development

### Build
```bash
./build.sh
```

This creates two binaries:
- `cmdcenter-default`: Standard build
- `cmdcenter-optimized`: Optimized build

### Project Structure
```
cmdcenter/
├── main.go           # CLI entry point
├── server.go         # HTTP server and UI
├── daemon.go         # Daemon management
├── config.go         # Config file handling
├── build.sh          # Build script
├── AGENTS.md         # Agent documentation
├── README.md         # This file
└── .agents/
    └── skills/
        └── cmdcenter-cheatsheet/  # Agent skill
```

## API Endpoints

### Command Execution
- `POST /api/command` - Execute command (JSON: `{"command": "id", "args": "optional"}`)

### Config Management
- `GET /api/config` - Get current config
- `POST /api/config/reload` - Reload config from file
- `GET /api/config/commands` - Get all commands
- `POST /api/config/commands/add` - Add command
- `POST /api/config/commands/edit` - Edit command
- `POST /api/config/commands/remove` - Remove command
- `GET /api/config/raw` - Get raw config JSON
- `POST /api/config/raw` - Save raw config JSON

### Logs
- `GET /api/logs` - Get daemon logs (optional `?search=query`)
- `POST /api/logs/clear` - Clear daemon logs

### System
- `GET /api/status` - Server status
- `GET /api/health` - Health check

## Agent Usage

cmdcenter is designed to be agent-first. AI agents can use the CLI for all operations without needing the web UI. See `AGENTS.md` for detailed agent documentation.

### Example Agent Workflow
```bash
# Setup commands for a task
cmdcenter add --id deploy --name "Deploy" --command "./deploy.sh" --icon "🚀"
cmdcenter add --id test --name "Run Tests" --command "./test.sh" --icon "✅"

# Execute commands
cmdcenter run --id deploy
cmdcenter run --id test

# Cleanup
cmdcenter remove --id deploy
```

## License

MIT

## Contributing

Contributions welcome! Please read AGENTS.md for agent-specific guidelines.
