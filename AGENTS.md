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

## Docker Deployment (v2.4.1+)

### Docker Commands for Agents
Agents can deploy and manage cmdcenter via Docker:

```bash
# Build Docker image
docker compose build

# Start container
docker compose up -d

# Stop container
docker compose down

# View logs
docker compose logs -f cmdcenter

# Restart container
docker compose restart

# Execute command inside container
docker exec -it cmdcenter sh
```

### Docker Configuration
Required files:
- `Dockerfile` - Multi-stage Go build configuration (for compose.yml)
- `compose.yml` - Build-based deployment (development)
- `compose.binary.yml` - Binary-based deployment (production)
- `config/config.json` - Persistent configuration (volume mounted)

**Important**: Docker configuration is cmdcenter's responsibility. Hotify only manages which compose file to use (similar to Coolify).

### Volume Mount
- Host directory: `./config`
- Container directory: `/root/.cmdcenter`
- Purpose: Persist commands across container restarts

### Traefik Integration
Docker deployment includes Traefik labels for automatic reverse proxy:
- Automatic SSL via Let's Encrypt
- Domain routing (e.g., cmdcenter.intrane.fr)
- HTTP to HTTPS redirect

Enable Traefik Docker provider if needed:
```bash
hotify-cli docker enable-traefik
sudo systemctl restart traefik
```

### Docker Deployment Workflow
```bash
# 1. Stop non-docker process if running
pkill cmdcenter

# 2. Ensure config exists
mkdir -p config
echo '{"title":"Command Center","subtitle":"Generic command execution dashboard","commands":[]}' > config/config.json

# 3. Choose compose file and start
# For development (builds from source):
docker compose -f compose.yml up -d

# For production (uses pre-built binary):
docker compose -f compose.binary.yml up -d

# 4. Enable Traefik provider (if not already)
hotify-cli docker enable-traefik
sudo systemctl restart traefik

# 5. Verify deployment
curl http://localhost:3031/api/health
curl https://cmdcenter.intrane.fr/api/health
```

### Choosing the Right Compose File
- **compose.yml**: Use for development, builds from source (~80s build, ~20MB image)
- **compose.binary.yml**: Use for production, uses pre-built binary (no build, ~74MB image)
- Hotify can switch between compose files (cmdcenter provides the files)

### API Access in Docker
API endpoints work the same in Docker:
- Local: `http://localhost:3031/api/*`
- Via Traefik: `https://cmdcenter.intrane.fr/api/*`

Config management via API persists to volume mount:
```bash
# Add command via API (persists to host config/)
curl -X POST http://localhost:3031/api/config/commands/add \
  -H 'Content-Type: application/json' \
  -d '{"id":"test","name":"Test","command":"echo hello","icon":"🧪"}'
```

## Notes for Agents
- Always use `cmdcenter init` to create initial config (non-Docker)
- For Docker, ensure `config/config.json` exists with proper structure
- Use unique IDs for commands to avoid conflicts
- The `supports-args` flag enables argument input modal in UI and CLI args support
- Config changes via CLI automatically update the running daemon (non-Docker)
- Config changes via API persist to volume mount in Docker
- Logs: `/tmp/cmdcenter.log` (non-Docker) or `docker compose logs` (Docker)
- Docker deployment requires Docker and Docker Compose
- Traefik integration requires hotify-cli v2.4.0+ and Traefik v3.6+
- Known bug: Config file may have `"commands": null` after removal - should be `"commands": []`
