# Boilerplate CLI UI Go

Go CLI with HTTP server and simple UI, compilable to binary with daemon start/stop commands.

## Features

- **CLI Interface**: Command-line tool with daemon management
- **HTTP Server**: Built-in web server with simple UI
- **Daemon Mode**: Background process management
- **Process Control**: Start, stop, and status commands
- **Simple UI**: Clean web interface for server status
- **Binary Compilation**: Single executable binary

## Usage

### Build

```bash
chmod +x build.sh
./build.sh
```

### CLI Commands

**Start HTTP server (foreground):**
```bash
./boilerplate-cli-ui-go-optimized start
```

**Start HTTP server on custom port:**
```bash
./boilerplate-cli-ui-go-optimized start -port 3000
```

**Start as daemon (background):**
```bash
./boilerplate-cli-ui-go-optimized start -daemon
```

**Start daemon on custom port:**
```bash
./boilerplate-cli-ui-go-optimized start -port 3000 -daemon
```

**Stop daemon:**
```bash
./boilerplate-cli-ui-go-optimized stop
```

**Check daemon status:**
```bash
./boilerplate-cli-ui-go-optimized status
```

**Show version:**
```bash
./boilerplate-cli-ui-go-optimized version
```

## Web Interface

When the server is running, access the UI at:
- `http://localhost:8080` (default port)
- `http://localhost:3000` (if started with -port 3000)

### API Endpoints

- `GET /` - Web UI
- `GET /api/status` - Server status (JSON)
- `GET /api/health` - Health check (JSON)

## Daemon Management

The daemon mode allows the HTTP server to run in the background:

- **PID File**: `/tmp/boilerplate-cli-ui-go.pid`
- **Log File**: `/tmp/boilerplate-cli-ui-go.log`
- **Process Control**: SIGTERM for graceful shutdown

## Architecture

```
CLI (boilerplate-cli-ui-go)
├── main.go - CLI commands and flag parsing
├── server.go - HTTP server and web UI
└── daemon.go - Process management
```

## Binary Size

- **Default**: ~2MB
- **Optimized**: ~1.3MB (with -ldflags "-s -w")

## Comparison with boilerplate-go

| Feature | boilerplate-go | boilerplate-cli-ui-go |
|---------|---------------|----------------------|
| CLI Commands | greet, version, help | start, stop, status, version, help |
| HTTP Server | No | Yes (with UI) |
| Daemon Mode | No | Yes (background process) |
| Binary Size | ~1.3MB | ~1.3MB |
| Use Case | Simple CLI | CLI with web interface |

## Requirements

- Go 1.21+

## Examples

### Development Workflow
```bash
# Build the binary
./build.sh

# Start server in foreground for development
./boilerplate-cli-ui-go-optimized start

# In another terminal, test the API
curl http://localhost:8080/api/status

# Stop with Ctrl+C
```

### Production Workflow
```bash
# Start as daemon
./boilerplate-cli-ui-go-optimized start -daemon

# Check status
./boilerplate-cli-ui-go-optimized status

# View logs
tail -f /tmp/boilerplate-cli-ui-go.log

# Stop when done
./boilerplate-cli-ui-go-optimized stop
```

## Use Cases

- **CLI Tools**: Add web interface to existing CLI tools
- **Microservices**: Small HTTP services with CLI management
- **Admin Panels**: Simple admin interfaces for system tools
- **Development**: Quick prototyping of CLI + web applications
- **Monitoring**: Status dashboards for long-running processes

## Future Enhancements

- [ ] Add configuration file support
- [ ] Add authentication for web UI
- [ ] Add HTTPS support
- [ ] Add systemd service file generation
- [ ] Add more API endpoints
- [ ] Add database integration
- [ ] Add metrics/monitoring

## License

This boilerplate is provided as-is for educational and development purposes.