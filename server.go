package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	server *http.Server
	mu     sync.Mutex
)

const (
	maxLogLines = 2000
)

type Status struct {
	Status    string    `json:"status"`
	Port      int       `json:"port"`
	Uptime    string    `json:"uptime"`
	StartTime time.Time `json:"start_time"`
}

var serverStatus Status
var appConfig    Config

func startServer(port int) {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
	}

	serverStatus = Status{
		Status:    "running",
		Port:      port,
		StartTime: time.Now(),
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/", handleHome)
	mux.HandleFunc("/api/status", handleStatusAPI)
	mux.HandleFunc("/api/health", handleHealthAPI)
	mux.HandleFunc("/api/command", handleCommandAPI)
	mux.HandleFunc("/api/logs", handleLogsAPI)
	mux.HandleFunc("/api/logs/clear", handleClearLogsAPI)
	mux.HandleFunc("/api/config", handleConfigAPI)
	mux.HandleFunc("/api/config/reload", handleConfigReloadAPI)
	mux.HandleFunc("/api/config/commands", handleCommandsAPI)
	mux.HandleFunc("/api/config/commands/add", handleAddCommandAPI)
	mux.HandleFunc("/api/config/commands/edit", handleEditCommandAPI)
	mux.HandleFunc("/api/config/commands/remove", handleRemoveCommandAPI)
	mux.HandleFunc("/api/config/raw", handleRawConfigAPI)

	server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Server starting on http://localhost:%d", port)
	log.Printf("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

func loadConfig() error {
	// Set default config
	appConfig = Config{
		Title:    "Command Center",
		Subtitle: "Generic command execution dashboard",
		Commands: []Command{},
	}

	// Try to load from config file
	config, err := loadConfigFile()
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Config file not found, using defaults")
			return nil
		}
		return err
	}

	appConfig = *config
	log.Printf("Loaded config with %d commands", len(appConfig.Commands))
	return nil
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Command Center</title>
    <style>
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap');
        
        :root {
            --bg-primary: #FFFFFF;
            --bg-secondary: #F9F9F8;
            --border-color: #EAEAEA;
            --text-primary: #111111;
            --text-secondary: #787774;
            --accent-black: #111111;
            --accent-gray: #333333;
        }
        
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
            line-height: 1.6;
        }
        
        .container {
            max-width: 1200px;
            width: 100%;
        }
        
        .header {
            margin-bottom: 40px;
            padding-bottom: 24px;
            border-bottom: 1px solid var(--border-color);
        }
        
        .header h1 {
            font-size: 32px;
            font-weight: 600;
            letter-spacing: -0.02em;
            color: var(--text-primary);
            margin-bottom: 8px;
        }
        
        .header .subtitle {
            font-size: 16px;
            color: var(--text-secondary);
            font-weight: 400;
        }
        
        .category-tabs {
            display: flex;
            gap: 8px;
            margin-bottom: 32px;
            flex-wrap: wrap;
        }
        
        .category-tab {
            background: var(--bg-secondary);
            color: var(--text-secondary);
            border: 1px solid var(--border-color);
            padding: 8px 16px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        
        .category-tab:hover {
            background: var(--border-color);
            color: var(--text-primary);
        }
        
        .category-tab.active {
            background: var(--accent-black);
            color: white;
            border-color: var(--accent-black);
        }
        
        .bento-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 16px;
            margin-bottom: 24px;
        }
        
        .card {
            background: var(--bg-primary);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 24px;
            transition: all 0.2s ease;
        }
        
        .card:hover {
            box-shadow: 0 2px 8px rgba(0,0,0,0.04);
        }
        
        .card-full {
            grid-column: span 2;
        }
        
        .card h3 {
            color: var(--text-primary);
            margin-bottom: 16px;
            font-size: 15px;
            font-weight: 600;
            letter-spacing: -0.01em;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        
        .command-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 10px;
        }
        
        .cmd-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 12px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
            text-align: center;
        }
        
        .cmd-btn:hover { 
            background: var(--accent-gray);
        }
        
        .cmd-btn:active { 
            transform: scale(0.98);
        }
        
        .cmd-btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none;
        }
        
        .cmd-btn.success {
            background: #346538;
        }
        
        .cmd-btn.success:hover {
            background: #2a522e;
        }
        
        .cmd-btn.error {
            background: #9F2F2D;
        }
        
        .cmd-btn.error:hover {
            background: #8a2624;
        }
        
        .link-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 12px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
            text-decoration: none;
            display: inline-block;
            text-align: center;
            width: 100%;
        }
        .link-btn:hover {
            background: var(--accent-gray);
        }
        .link-btn:active { 
            transform: scale(0.98);
        }
        .log-output {
            background: var(--bg-secondary);
            color: var(--text-primary);
            padding: 16px;
            border-radius: 6px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 12px;
            min-height: 160px;
            max-height: 320px;
            overflow-y: auto;
            white-space: pre-wrap;
            word-break: break-all;
            line-height: 1.5;
            border: 1px solid var(--border-color);
        }
        .search-box {
            display: flex;
            gap: 8px;
            margin-bottom: 16px;
        }
        .search-input {
            flex: 1;
            padding: 10px 12px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            font-size: 13px;
            font-family: inherit;
            background: var(--bg-primary);
            transition: all 0.15s ease;
        }
        .search-input:focus {
            outline: none;
            border-color: var(--accent-black);
            background: var(--bg-primary);
        }
        .refresh-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 10px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .refresh-btn:hover {
            background: var(--accent-gray);
        }
        .clear-btn {
            background: #9F2F2D;
            color: white;
            border: none;
            padding: 10px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .clear-btn:hover {
            background: #8a2624;
        }
        .manage-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 10px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .manage-btn:hover {
            background: var(--accent-gray);
        }
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.4);
            backdrop-filter: blur(4px);
            z-index: 1000;
        }
        #commandModal {
            z-index: 1100;
        }
        #rawConfigModal {
            z-index: 1200;
        }
        #commandArgsModal {
            z-index: 1300;
        }
        .toast-container {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 2000;
            display: flex;
            flex-direction: column;
            gap: 8px;
        }
        .toast {
            background: var(--bg-primary);
            border-radius: 6px;
            padding: 12px 16px;
            box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
            display: flex;
            align-items: center;
            gap: 10px;
            min-width: 280px;
            max-width: 360px;
            animation: slideIn 0.2s ease;
            border: 1px solid var(--border-color);
        }
        .toast.success {
            border-left: 3px solid #346538;
        }
        .toast.error {
            border-left: 3px solid #9F2F2D;
        }
        .toast.warning {
            border-left: 3px solid #956400;
        }
        .toast-icon {
            font-size: 16px;
        }
        .toast-message {
            flex: 1;
            font-size: 13px;
            color: var(--text-primary);
            font-weight: 400;
        }
        .toast-close {
            background: none;
            border: none;
            color: var(--text-secondary);
            cursor: pointer;
            font-size: 16px;
            padding: 2px;
            line-height: 1;
        }
        .toast-close:hover {
            color: var(--text-primary);
        }
        @keyframes slideIn {
            from {
                transform: translateX(100%);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }
        @keyframes slideOut {
            from {
                transform: translateX(0);
                opacity: 1;
            }
            to {
                transform: translateX(100%);
                opacity: 0;
            }
        }
        .confirm-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.4);
            backdrop-filter: blur(4px);
            z-index: 1400;
            display: none;
            align-items: center;
            justify-content: center;
        }
        .confirm-overlay.active {
            display: flex;
        }
        .confirm-dialog {
            background: var(--bg-primary);
            border-radius: 8px;
            padding: 24px;
            max-width: 400px;
            width: 90%;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
            border: 1px solid var(--border-color);
        }
        .confirm-title {
            font-size: 16px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 12px;
            letter-spacing: -0.01em;
        }
        .confirm-message {
            font-size: 14px;
            color: var(--text-secondary);
            line-height: 1.5;
            margin-bottom: 20px;
        }
        .confirm-actions {
            display: flex;
            gap: 8px;
            justify-content: flex-end;
        }
        .confirm-btn {
            padding: 10px 16px;
            border-radius: 4px;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            cursor: pointer;
            transition: all 0.15s ease;
            border: none;
        }
        .confirm-btn-cancel {
            background: var(--bg-secondary);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
        }
        .confirm-btn-cancel:hover {
            background: var(--border-color);
        }
        .confirm-btn-confirm {
            background: var(--accent-black);
            color: white;
        }
        .confirm-btn-confirm:hover {
            background: var(--accent-gray);
        }
        .confirm-btn-danger {
            background: #9F2F2D;
            color: white;
        }
        .confirm-btn-danger:hover {
            background: #8a2624;
        }
        .modal.active {
            display: block;
        }
        .modal-content {
            background: var(--bg-primary);
            border-radius: 8px;
            padding: 24px;
            max-width: 600px;
            width: 90%;
            max-height: 80vh;
            overflow-y: auto;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
        }
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 16px;
            border-bottom: 1px solid var(--border-color);
        }
        .modal-header h2 {
            color: var(--text-primary);
            font-size: 18px;
            font-weight: 600;
            letter-spacing: -0.01em;
        }
        .close-btn {
            background: none;
            border: none;
            font-size: 20px;
            cursor: pointer;
            color: var(--text-secondary);
            padding: 4px;
            line-height: 1;
        }
        .close-btn:hover {
            color: var(--text-primary);
        }
        .form-group {
            margin-bottom: 16px;
        }
        .form-group label {
            display: block;
            margin-bottom: 6px;
            color: var(--text-secondary);
            font-size: 13px;
            font-weight: 500;
        }
        .form-group input,
        .form-group textarea,
        .form-group select {
            width: 100%;
            padding: 10px 12px;
            border: 1px solid var(--border-color);
            border-radius: 4px;
            font-size: 13px;
            font-family: inherit;
            background: var(--bg-primary);
            transition: all 0.15s ease;
            box-sizing: border-box;
        }
        .form-group input:focus,
        .form-group textarea:focus,
        .form-group select:focus {
            outline: none;
            border-color: var(--accent-black);
        }
        .form-group textarea {
            min-height: 60px;
            resize: vertical;
        }
        .form-actions {
            display: flex;
            gap: 8px;
            justify-content: flex-end;
            margin-top: 20px;
            padding-top: 16px;
            border-top: 1px solid var(--border-color);
        }
        .submit-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 10px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .submit-btn:hover {
            background: var(--accent-gray);
        }
        .cancel-btn {
            background: var(--bg-secondary);
            color: var(--text-primary);
            border: 1px solid var(--border-color);
            padding: 10px 16px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 13px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .cancel-btn:hover {
            background: var(--border-color);
        }
        .command-list {
            margin-top: 16px;
            max-height: 50vh;
            overflow-y: auto;
            padding-right: 4px;
        }
        .command-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px;
            background: var(--bg-secondary);
            border-radius: 6px;
            margin-bottom: 8px;
            border: 1px solid var(--border-color);
        }
        .command-item-info {
            flex: 1;
        }
        .command-item-name {
            font-weight: 500;
            color: var(--text-primary);
            font-size: 14px;
        }
        .command-item-id {
            color: var(--text-secondary);
            font-size: 11px;
            margin-bottom: 4px;
            font-family: 'JetBrains Mono', monospace;
        }
        .command-item-desc {
            color: var(--text-secondary);
            font-size: 12px;
        }
        .command-item-actions {
            display: flex;
            gap: 6px;
        }
        .edit-btn {
            background: var(--accent-black);
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .edit-btn:hover {
            background: var(--accent-gray);
        }
        .delete-btn {
            background: #9F2F2D;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            font-family: inherit;
            transition: all 0.15s ease;
        }
        .delete-btn:hover {
            background: #8a2624;
        }
        .daemon-log-output {
            background: var(--bg-secondary);
            color: var(--text-primary);
            padding: 16px;
            border-radius: 6px;
            font-family: 'JetBrains Mono', monospace;
            font-size: 12px;
            min-height: 300px;
            max-height: 500px;
            overflow-y: auto;
            white-space: pre-wrap;
            word-break: break-all;
            line-height: 1.5;
            border: 1px solid var(--border-color);
        }
        .log-stats {
            font-size: 12px;
            color: var(--text-secondary);
            margin-top: 12px;
            font-weight: 400;
        }
        @media (max-width: 768px) {
            .bento-grid {
                grid-template-columns: 1fr;
            }
            .card-full {
                grid-column: span 1;
            }
            .command-grid {
                grid-template-columns: repeat(2, 1fr);
            }
            .container {
                padding: 24px 20px;
            }
        }
        @media (max-width: 480px) {
            .command-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 id="appTitle">Command Center</h1>
            <p class="subtitle" id="appSubtitle">Generic command execution dashboard</p>
        </div>

        <div class="category-tabs" id="categoryTabs">
            <button class="category-tab active" data-category="all">All</button>
        </div>

        <div class="bento-grid">
            <div class="card card-full">
                <h3>Commands & Links</h3>
                <div class="command-grid" id="commandGrid">
                    <div class="cmd-btn">Loading commands...</div>
                </div>
            </div>
            
            <div class="card card-full">
                <h3>Command Output</h3>
                <div class="log-output" id="logOutput">Ready to execute commands...</div>
            </div>
            
            <div class="card card-full">
                <h3>Daemon Logs</h3>
                <div class="search-box">
                    <input type="text" class="search-input" id="logSearch" placeholder="Search logs..." onkeyup="filterLogs()">
                    <button class="refresh-btn" onclick="loadDaemonLogs()">Refresh</button>
                    <button class="clear-btn" onclick="clearLogs()">Clear</button>
                    <button class="manage-btn" onclick="openCommandManager()">Manage Commands</button>
                </div>
                <div class="daemon-log-output" id="daemonLogOutput">Loading logs...</div>
                <div class="log-stats" id="logStats"></div>
            </div>
        </div>

        <!-- Command Management Modal -->
        <div class="modal" id="commandModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2 id="modalTitle">Add Command</h2>
                    <button class="close-btn" onclick="closeCommandModal()">&times;</button>
                </div>
                <form id="commandForm" onsubmit="handleCommandSubmit(event)">
                    <input type="hidden" id="editCommandId" value="">
                    <div class="form-group">
                        <label for="commandId">ID *</label>
                        <input type="text" id="commandId" required placeholder="e.g., my-command">
                    </div>
                    <div class="form-group">
                        <label for="commandName">Name *</label>
                        <input type="text" id="commandName" required placeholder="e.g., My Command">
                    </div>
                    <div class="form-group">
                        <label for="commandIcon">Icon (optional)</label>
                        <input type="text" id="commandIcon" placeholder="e.g., 📋">
                    </div>
                    <div class="form-group">
                        <label for="commandDescription">Description</label>
                        <textarea id="commandDescription" placeholder="Brief description of what this command does"></textarea>
                    </div>
                    <div class="form-group">
                        <label for="commandType">Type *</label>
                        <select id="commandType" required>
                            <option value="command">Command</option>
                            <option value="link">Link</option>
                        </select>
                    </div>
                    <div class="form-group" id="commandGroup">
                        <label for="commandCommand">Command *</label>
                        <textarea id="commandCommand" placeholder="e.g., echo 'Hello World'"></textarea>
                    </div>
                    <div class="form-group" id="urlGroup" style="display: none;">
                        <label for="commandURL">URL *</label>
                        <input type="url" id="commandURL" placeholder="e.g., https://example.com">
                    </div>
                    <div class="form-group">
                        <label for="commandCategory">Category</label>
                        <input type="text" id="commandCategory" placeholder="e.g., Development, Monitoring, Tools">
                    </div>
                    <div class="form-group" id="supportsArgsGroup">
                        <label style="display: flex; align-items: center; gap: 8px;">
                            <input type="checkbox" id="commandSupportsArgs" style="width: auto;">
                            Enable argument support for this command
                        </label>
                        <small style="color: var(--text-secondary); display: block; margin-top: 4px; font-size: 12px;">When enabled, clicking this command will show a modal to input extra arguments</small>
                    </div>
                    <div class="form-group" id="argsDescriptionGroup" style="display: none;">
                        <label for="commandArgsDescription">Arguments Description</label>
                        <textarea id="commandArgsDescription" placeholder="e.g., --headed true/false for HEADED mode, --args for additional arguments"></textarea>
                        <small style="color: var(--text-secondary); display: block; margin-top: 4px; font-size: 12px;">Document the possible arguments this command accepts</small>
                    </div>
                    <div class="form-group">
                        <label for="commandEnv">Environment Variables</label>
                        <textarea id="commandEnv" placeholder="e.g., NODE_ENV=production,API_KEY=secret"></textarea>
                        <small style="color: var(--text-secondary); display: block; margin-top: 4px; font-size: 12px;">Comma-separated KEY=VALUE pairs for environment variables</small>
                    </div>
                    <div class="form-actions">
                        <button type="button" class="cancel-btn" onclick="closeCommandModal()">Cancel</button>
                        <button type="submit" class="submit-btn">Save Command</button>
                    </div>
                </form>
            </div>
        </div>

        <!-- Command List Modal -->
        <div class="modal" id="commandListModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Manage Commands</h2>
                    <button class="close-btn" onclick="closeCommandListModal()">&times;</button>
                </div>
                <div style="display: flex; gap: 8px; margin-bottom: 20px;">
                    <button class="submit-btn" onclick="openAddCommandModal()" style="flex: 1;">Add Command</button>
                    <button class="refresh-btn" onclick="reloadConfig()" style="flex: 1;">Reload Config</button>
                    <button class="manage-btn" onclick="openRawConfigModal()" style="flex: 1;">Edit Raw Config</button>
                </div>
                <div class="command-list" id="commandList"></div>
            </div>
        </div>

        <!-- Raw Config Modal -->
        <div class="modal" id="rawConfigModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2>Edit Raw Config</h2>
                    <button class="close-btn" onclick="closeRawConfigModal()">&times;</button>
                </div>
                <div class="form-group">
                    <label for="rawConfigEditor">Configuration JSON</label>
                    <textarea id="rawConfigEditor" style="font-family: 'Courier New', monospace; min-height: 400px;"></textarea>
                </div>
                <div class="form-actions">
                    <button type="button" class="cancel-btn" onclick="closeRawConfigModal()">Cancel</button>
                    <button type="button" class="submit-btn" onclick="saveRawConfig()">Save & Reload</button>
                </div>
            </div>
        </div>

        <!-- Command Args Modal -->
        <div class="modal" id="commandArgsModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h2 id="argsModalTitle">Execute Command with Args</h2>
                    <button class="close-btn" onclick="closeCommandArgsModal()">&times;</button>
                </div>
                <div class="form-group">
                    <label for="commandArgs">Extra Arguments</label>
                    <input type="text" id="commandArgs" placeholder="e.g., -h, --help, /path/to/file">
                    <small style="color: var(--text-secondary); display: block; margin-top: 4px; font-size: 12px;">These arguments will be appended to the command</small>
                </div>
                <div class="form-group" id="argsDescriptionDisplay" style="display: none;">
                    <label>Available Arguments</label>
                    <div id="argsDescriptionText" style="background: var(--bg-secondary); padding: 12px; border-radius: 4px; font-size: 13px; border: 1px solid var(--border-color); color: var(--text-primary); font-family: 'JetBrains Mono', monospace;"></div>
                </div>
                <div class="form-group">
                    <label>Full Command Preview</label>
                    <div id="commandPreview" style="background: var(--bg-secondary); padding: 12px; border-radius: 4px; font-family: 'JetBrains Mono', monospace; font-size: 12px; border: 1px solid var(--border-color); color: var(--text-primary);"></div>
                </div>
                <div class="form-actions">
                    <button type="button" class="cancel-btn" onclick="closeCommandArgsModal()">Cancel</button>
                    <button type="button" class="submit-btn" onclick="executeCommandWithArgs()">Execute</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Toast Container -->
    <div class="toast-container" id="toastContainer"></div>

    <!-- Confirmation Dialog -->
    <div class="confirm-overlay" id="confirmOverlay">
        <div class="confirm-dialog">
            <div class="confirm-title" id="confirmTitle">Confirm Action</div>
            <div class="confirm-message" id="confirmMessage">Are you sure you want to proceed?</div>
            <div class="confirm-actions">
                <button class="confirm-btn confirm-btn-cancel" id="confirmCancel">Cancel</button>
                <button class="confirm-btn confirm-btn-confirm" id="confirmOk">Confirm</button>
            </div>
        </div>
    </div>

    <script>
        let appConfig = null;
        let currentCategory = 'all';

        // Toast Notification System
        function showToast(message, type, duration) {
            type = type || 'info';
            duration = duration || 3000;

            const container = document.getElementById('toastContainer');
            const toast = document.createElement('div');
            toast.className = 'toast ' + type;

            let icon = '○';
            if (type === 'success') icon = '✓';
            if (type === 'error') icon = '✕';
            if (type === 'warning') icon = '!';

            toast.innerHTML = '<span class="toast-icon">' + icon + '</span><span class="toast-message">' + message + '</span><button class="toast-close" onclick="removeToast(this.parentElement)">×</button>';

            container.appendChild(toast);

            // Auto-remove after duration
            setTimeout(function() {
                removeToast(toast);
            }, duration);
        }

        function removeToast(toast) {
            toast.style.animation = 'slideOut 0.2s ease';
            setTimeout(function() {
                if (toast.parentElement) {
                    toast.parentElement.removeChild(toast);
                }
            }, 200);
        }

        // Custom Confirmation Dialog
        let confirmCallback = null;

        function showConfirm(title, message, onConfirm, isDangerous) {
            return new Promise(function(resolve) {
                const overlay = document.getElementById('confirmOverlay');
                const titleEl = document.getElementById('confirmTitle');
                const messageEl = document.getElementById('confirmMessage');
                const cancelBtn = document.getElementById('confirmCancel');
                const okBtn = document.getElementById('confirmOk');

                titleEl.textContent = title;
                messageEl.textContent = message;

                if (isDangerous) {
                    okBtn.className = 'confirm-btn confirm-btn-danger';
                } else {
                    okBtn.className = 'confirm-btn confirm-btn-confirm';
                }

                // Set up callback
                confirmCallback = function(confirmed) {
                    overlay.classList.remove('active');
                    resolve(confirmed);
                };

                // Show dialog
                overlay.classList.add('active');
            });
        }

        document.getElementById('confirmCancel').addEventListener('click', function() {
            if (confirmCallback) {
                confirmCallback(false);
            }
        });

        document.getElementById('confirmOk').addEventListener('click', function() {
            if (confirmCallback) {
                confirmCallback(true);
            }
        });

        // Close on overlay click
        document.getElementById('confirmOverlay').addEventListener('click', function(e) {
            if (e.target.id === 'confirmOverlay') {
                if (confirmCallback) {
                    confirmCallback(false);
                }
            }
        });

        function loadConfig() {
            fetch('/api/config')
                .then(response => response.json())
                .then(data => {
                    appConfig = data;
                    document.getElementById('appTitle').textContent = '🎯 ' + data.title;
                    document.getElementById('appSubtitle').textContent = data.subtitle;
                    renderCategories(data.commands);
                    renderCommands(data.commands);
                })
                .catch(error => {
                    console.error('Error loading config:', error);
                    document.getElementById('commandGrid').innerHTML = '<div class="cmd-btn">Error loading commands</div>';
                });
        }

        function renderCategories(commands) {
            const categories = new Set();
            commands.forEach(cmd => {
                if (cmd.category) {
                    categories.add(cmd.category);
                }
            });

            const tabsContainer = document.getElementById('categoryTabs');
            tabsContainer.innerHTML = '<button class="category-tab active" data-category="all">All</button>';

            categories.forEach(category => {
                const tab = document.createElement('button');
                tab.className = 'category-tab';
                tab.setAttribute('data-category', category);
                tab.textContent = category;
                tab.onclick = () => filterByCategory(category);
                tabsContainer.appendChild(tab);
            });

            // Add click handler to "All" tab
            tabsContainer.querySelector('[data-category="all"]').onclick = () => filterByCategory('all');
        }

        function filterByCategory(category) {
            currentCategory = category;

            // Update tab styling
            document.querySelectorAll('.category-tab').forEach(tab => {
                tab.classList.remove('active');
                if (tab.getAttribute('data-category') === category) {
                    tab.classList.add('active');
                }
            });

            // Filter commands
            const filteredCommands = category === 'all'
                ? appConfig.commands
                : appConfig.commands.filter(cmd => cmd.category === category);

            renderCommands(filteredCommands);
        }
        
        function renderCommands(commands) {
            const grid = document.getElementById('commandGrid');
            grid.innerHTML = '';

            commands.forEach(cmd => {
                if ((cmd.type === 'link' || cmd.type === '') && cmd.url) {
                    // Render as link
                    const link = document.createElement('a');
                    link.className = 'link-btn';
                    link.textContent = cmd.icon + ' ' + cmd.name;
                    link.href = cmd.url;
                    link.target = '_blank';
                    link.title = cmd.description;
                    grid.appendChild(link);
                } else {
                    // Render as command button
                    const button = document.createElement('button');
                    button.className = 'cmd-btn';
                    button.textContent = cmd.icon + ' ' + cmd.name;
                    button.title = cmd.description;
                    button.setAttribute('data-original-text', cmd.icon + ' ' + cmd.name);
                    
                    // Handle single/double click distinction
                    let clickTimeout;
                    button.onclick = function(e) {
                        if (clickTimeout) {
                            // This is a double-click
                            clearTimeout(clickTimeout);
                            clickTimeout = null;
                            executeCommand(cmd.id, cmd.name, true);
                        } else {
                            // Wait to see if it's a double-click
                            clickTimeout = setTimeout(function() {
                                clickTimeout = null;
                                executeCommand(cmd.id, cmd.name, false);
                            }, 250);
                        }
                    };
                    
                    grid.appendChild(button);
                }
            });
        }
        
        function executeCommand(commandId, commandName, skipArgs = false) {
            const logOutput = document.getElementById('logOutput');
            const button = event.target;

            // Find command in config
            const cmd = appConfig.commands.find(c => c.id === commandId);
            if (cmd && cmd.supports_args && !skipArgs) {
                // Show args modal
                openCommandArgsModal(commandId, commandName, cmd.command, cmd.args_description);
                return;
            }

            logOutput.textContent = 'Executing: ' + commandName + '...\n';
            button.disabled = true;
            button.textContent = '⏳ Running...';

            fetch('/api/command', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ command: commandId })
            })
            .then(response => response.json())
            .then(data => {
                logOutput.textContent = data.output || 'Command completed';
                if (data.success) {
                    button.classList.add('success');
                    button.textContent = '✅ Done';
                } else {
                    button.classList.add('error');
                    button.textContent = '❌ Failed';
                }
            })
            .catch(error => {
                logOutput.textContent = 'Error: ' + error.message;
                button.classList.add('error');
                button.textContent = '❌ Error';
            })
            .finally(() => {
                setTimeout(() => {
                    button.disabled = false;
                    button.classList.remove('success', 'error');
                    button.textContent = button.getAttribute('data-original-text') || commandName;
                }, 2000);
            });
        }

        let currentCommandId = null;
        let currentCommandName = null;
        let currentBaseCommand = null;

        function openCommandArgsModal(commandId, commandName, baseCommand, argsDescription) {
            currentCommandId = commandId;
            currentCommandName = commandName;
            currentBaseCommand = baseCommand;

            document.getElementById('argsModalTitle').textContent = 'Execute: ' + commandName;
            document.getElementById('commandArgs').value = '';

            // Show args description if available
            const argsDescriptionDisplay = document.getElementById('argsDescriptionDisplay');
            const argsDescriptionText = document.getElementById('argsDescriptionText');
            if (argsDescription && argsDescription.trim() !== '') {
                argsDescriptionText.textContent = argsDescription;
                argsDescriptionDisplay.style.display = 'block';
            } else {
                argsDescriptionDisplay.style.display = 'none';
            }

            updateCommandPreview();
            document.body.style.overflow = 'hidden';
            document.getElementById('commandArgsModal').classList.add('active');

            // Add event listener for live preview
            document.getElementById('commandArgs').addEventListener('input', updateCommandPreview);
        }

        function closeCommandArgsModal() {
            document.getElementById('commandArgsModal').classList.remove('active');
            document.body.style.overflow = '';
            document.getElementById('commandArgs').removeEventListener('input', updateCommandPreview);
            currentCommandId = null;
            currentCommandName = null;
            currentBaseCommand = null;
        }

        function updateCommandPreview() {
            const args = document.getElementById('commandArgs').value;
            const preview = document.getElementById('commandPreview');
            if (args) {
                preview.textContent = currentBaseCommand + ' ' + args;
            } else {
                preview.textContent = currentBaseCommand;
            }
        }

        function executeCommandWithArgs() {
            const args = document.getElementById('commandArgs').value;
            const commandId = currentCommandId;
            const fullCommand = currentBaseCommand + (args ? ' ' + args : '');

            const logOutput = document.getElementById('logOutput');
            logOutput.textContent = 'Executing: ' + fullCommand + '...\n';

            closeCommandArgsModal();

            fetch('/api/command', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ command: commandId, args: args })
            })
            .then(response => response.json())
            .then(data => {
                logOutput.textContent = data.output || 'Command completed';
                if (data.success) {
                    showToast('Command executed successfully', 'success');
                } else {
                    showToast('Command failed', 'error');
                }
            })
            .catch(error => {
                logOutput.textContent = 'Error: ' + error.message;
                showToast('Error executing command', 'error');
            });
        }
        
        // Auto-refresh status every 30 seconds
        setInterval(() => {
            fetch('/api/status')
                .then(r => r.json())
                .then(data => {
                    console.log('Status:', data);
                })
                .catch(err => console.error('Error:', err));
        }, 30000);
        
        // Daemon logs functionality
        let daemonLogs = '';
        
        function loadDaemonLogs() {
            const search = document.getElementById('logSearch').value;
            const url = search ? '/api/logs?search=' + encodeURIComponent(search) : '/api/logs';
            
            fetch(url)
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        daemonLogs = data.logs;
                        if (data.logs && data.logs.trim() !== '') {
                            document.getElementById('daemonLogOutput').textContent = data.logs;
                        } else {
                            document.getElementById('daemonLogOutput').textContent = 'No logs available yet. Commands will be logged here when executed.';
                        }
                        document.getElementById('logStats').textContent = 'Lines: ' + (data.count || 0);
                    } else {
                        document.getElementById('daemonLogOutput').textContent = 'Error loading logs: ' + (data.error || 'Unknown error');
                        document.getElementById('logStats').textContent = '';
                    }
                })
                .catch(error => {
                    document.getElementById('daemonLogOutput').textContent = 'Error loading logs: ' + error.message;
                    document.getElementById('logStats').textContent = '';
                });
        }
        
        function filterLogs() {
            loadDaemonLogs();
        }
        
        function clearLogs() {
            showConfirm('Clear Logs', 'Are you sure you want to clear all daemon logs?', null, true).then(function(confirmed) {
                if (confirmed) {
                    fetch('/api/logs/clear', {
                        method: 'POST'
                    })
                    .then(response => response.json())
                    .then(function(data) {
                        if (data.success) {
                            loadDaemonLogs();
                            showToast('Logs cleared successfully', 'success');
                        } else {
                            showToast('Failed to clear logs: ' + (data.error || 'Unknown error'), 'error');
                        }
                    })
                    .catch(function(error) {
                        showToast('Error clearing logs: ' + error.message, 'error');
                    });
                }
            });
        }
        
        // Auto-refresh daemon logs every 10 seconds
        setInterval(function() {
            loadDaemonLogs();
        }, 10000);
        
        // Initial load
        loadConfig();
        loadDaemonLogs();

        // Command Management Functions
        function openCommandManager() {
            document.body.style.overflow = 'hidden';
            document.getElementById('commandListModal').classList.add('active');
            loadCommandList();
        }

        function closeCommandListModal() {
            document.getElementById('commandListModal').classList.remove('active');
            document.body.style.overflow = '';
        }

        function loadCommandList() {
            fetch('/api/config/commands')
                .then(response => response.json())
                .then(commands => {
                    const commandList = document.getElementById('commandList');
                    commandList.innerHTML = '';
                    commands.forEach(cmd => {
                        const item = document.createElement('div');
                        item.className = 'command-item';
                        item.innerHTML = '<div class="command-item-info"><div class="command-item-id">ID: ' + cmd.id + '</div><div class="command-item-name">' + (cmd.icon || '') + ' ' + cmd.name + '</div><div class="command-item-desc">' + (cmd.description || '') + '</div></div><div class="command-item-actions"><button class="edit-btn" onclick="openEditCommandModal(\'' + cmd.id + '\')">Edit</button><button class="delete-btn" onclick="deleteCommand(\'' + cmd.id + '\')">Delete</button></div>';
                        commandList.appendChild(item);
                    });
                })
                .catch(error => {
                    console.error('Error loading commands:', error);
                });
        }

        function openAddCommandModal() {
            document.getElementById('modalTitle').textContent = 'Add Command';
            document.getElementById('editCommandId').value = '';
            document.getElementById('commandId').value = '';
            document.getElementById('commandId').disabled = false;
            document.getElementById('commandName').value = '';
            document.getElementById('commandIcon').value = '';
            document.getElementById('commandDescription').value = '';
            document.getElementById('commandCommand').value = '';
            document.getElementById('commandURL').value = '';
            document.getElementById('commandCategory').value = '';
            document.getElementById('commandType').value = 'command';
            document.getElementById('commandSupportsArgs').checked = false;
            document.getElementById('commandArgsDescription').value = '';
            toggleTypeFields();
            document.body.style.overflow = 'hidden';
            document.getElementById('commandModal').classList.add('active');
        }

        // Type toggle handler
        document.getElementById('commandType').addEventListener('change', toggleTypeFields);
        
        // Supports args checkbox handler
        document.getElementById('commandSupportsArgs').addEventListener('change', toggleArgsDescription);

        function toggleTypeFields() {
            const type = document.getElementById('commandType').value;
            const commandGroup = document.getElementById('commandGroup');
            const urlGroup = document.getElementById('urlGroup');
            const supportsArgsGroup = document.getElementById('supportsArgsGroup');
            const commandCommand = document.getElementById('commandCommand');

            if (type === 'link') {
                commandGroup.style.display = 'none';
                urlGroup.style.display = 'block';
                supportsArgsGroup.style.display = 'none';
                commandCommand.removeAttribute('required');
                document.getElementById('commandURL').setAttribute('required', 'required');
            } else {
                commandGroup.style.display = 'block';
                urlGroup.style.display = 'none';
                supportsArgsGroup.style.display = 'block';
                commandCommand.setAttribute('required', 'required');
                document.getElementById('commandURL').removeAttribute('required');
            }
            // Always call toggleArgsDescription to ensure proper visibility
            toggleArgsDescription();
        }

        function toggleArgsDescription() {
            const supportsArgs = document.getElementById('commandSupportsArgs').checked;
            const argsDescriptionGroup = document.getElementById('argsDescriptionGroup');
            argsDescriptionGroup.style.display = supportsArgs ? 'block' : 'none';
        }

        function openEditCommandModal(commandId) {
            fetch('/api/config/commands')
                .then(response => response.json())
                .then(commands => {
                    const cmd = commands.find(c => c.id === commandId);
                    if (cmd) {
                        document.getElementById('modalTitle').textContent = 'Edit Command';
                        document.getElementById('editCommandId').value = cmd.id;
                        document.getElementById('commandId').value = cmd.id;
                        document.getElementById('commandId').disabled = true;
                        document.getElementById('commandName').value = cmd.name;
                        document.getElementById('commandIcon').value = cmd.icon || '';
                        document.getElementById('commandDescription').value = cmd.description || '';
                        document.getElementById('commandCommand').value = cmd.command || '';
                        document.getElementById('commandURL').value = cmd.url || '';
                        document.getElementById('commandCategory').value = cmd.category || '';
                        document.getElementById('commandType').value = cmd.type || 'command';
                        document.getElementById('commandSupportsArgs').checked = cmd.supports_args || false;
                        document.getElementById('commandArgsDescription').value = cmd.args_description || '';
                        // Convert env object back to comma-separated string
                        let envString = '';
                        if (cmd.env && typeof cmd.env === 'object') {
                            envString = Object.entries(cmd.env).map(function(entry) { return entry[0] + '=' + entry[1]; }).join(',');
                        }
                        document.getElementById('commandEnv').value = envString;
                        toggleTypeFields();
                        document.body.style.overflow = 'hidden';
                        document.getElementById('commandModal').classList.add('active');
                    }
                })
                .catch(error => {
                    console.error('Error loading command:', error);
                });
        }

        function closeCommandModal() {
            document.getElementById('commandModal').classList.remove('active');
            document.body.style.overflow = '';
        }

        function handleCommandSubmit(event) {
            event.preventDefault();
            const editId = document.getElementById('editCommandId').value;
            const type = document.getElementById('commandType').value;
            const commandData = {
                id: document.getElementById('commandId').value,
                name: document.getElementById('commandName').value,
                icon: document.getElementById('commandIcon').value,
                description: document.getElementById('commandDescription').value,
                type: type,
                command: type === 'command' ? document.getElementById('commandCommand').value : '',
                url: type === 'link' ? document.getElementById('commandURL').value : '',
                category: document.getElementById('commandCategory').value,
                supports_args: type === 'command' ? document.getElementById('commandSupportsArgs').checked : false,
                args_description: type === 'command' ? document.getElementById('commandArgsDescription').value : '',
                env: document.getElementById('commandEnv').value
            };

            if (editId) {
                // Edit existing command
                fetch('/api/config/commands/edit', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(commandData)
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        closeCommandModal();
                        loadCommandList();
                        loadConfig();
                        showToast('Command updated successfully', 'success');
                    } else {
                        showToast('Error: ' + data.error, 'error');
                    }
                })
                .catch(error => {
                    console.error('Error updating command:', error);
                    showToast('Error updating command', 'error');
                });
            } else {
                // Add new command
                fetch('/api/config/commands/add', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(commandData)
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        closeCommandModal();
                        loadCommandList();
                        loadConfig();
                        showToast('Command added successfully', 'success');
                    } else {
                        showToast('Error: ' + data.error, 'error');
                    }
                })
                .catch(error => {
                    console.error('Error adding command:', error);
                    showToast('Error adding command', 'error');
                });
            }
        }

        function deleteCommand(commandId) {
            showConfirm('Delete Command', 'Are you sure you want to delete this command?', null, true).then(function(confirmed) {
                if (confirmed) {
                    fetch('/api/config/commands/remove', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ id: commandId })
                    })
                    .then(response => response.json())
                    .then(data => {
                        if (data.success) {
                            loadCommandList();
                            loadConfig();
                            showToast('Command deleted successfully', 'success');
                        } else {
                            showToast('Error: ' + data.error, 'error');
                        }
                    })
                    .catch(error => {
                        console.error('Error deleting command:', error);
                        showToast('Error deleting command', 'error');
                    });
                }
            });
        }

        function reloadConfig() {
            fetch('/api/config/reload', {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    loadCommandList();
                    loadConfig();
                    showToast('Config reloaded successfully', 'success');
                } else {
                    showToast('Error: ' + data.error, 'error');
                }
            })
            .catch(error => {
                console.error('Error reloading config:', error);
                showToast('Error reloading config', 'error');
            });
        }

        function openRawConfigModal() {
            fetch('/api/config/raw')
                .then(response => response.json())
                .then(config => {
                    const editor = document.getElementById('rawConfigEditor');
                    editor.value = JSON.stringify(config, null, 2);
                    document.body.style.overflow = 'hidden';
                    document.getElementById('rawConfigModal').classList.add('active');
                })
                .catch(error => {
                    console.error('Error loading config:', error);
                    showToast('Error loading config', 'error');
                });
        }

        function closeRawConfigModal() {
            document.getElementById('rawConfigModal').classList.remove('active');
            document.body.style.overflow = '';
        }

        function saveRawConfig() {
            const editor = document.getElementById('rawConfigEditor');
            try {
                const config = JSON.parse(editor.value);
                fetch('/api/config/raw', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(config)
                })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        closeRawConfigModal();
                        loadCommandList();
                        loadConfig();
                        showToast('Config saved and reloaded successfully', 'success');
                    } else {
                        showToast('Error: ' + data.error, 'error');
                    }
                })
                .catch(error => {
                    console.error('Error saving config:', error);
                    showToast('Error saving config', 'error');
                });
            } catch (error) {
                showToast('Invalid JSON: ' + error.message, 'error');
            }
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleStatusAPI(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	serverStatus.Uptime = time.Since(serverStatus.StartTime).String()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(serverStatus)
}

func handleHealthAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func handleCommandAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmdRequest struct {
		Command string `json:"command"`
		Args    string `json:"args"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cmdRequest); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"output":  "Invalid request format",
		})
		return
	}

	var output string
	var success bool

	// Find command in config
	var cmdConfig *Command
	for i := range appConfig.Commands {
		if appConfig.Commands[i].ID == cmdRequest.Command {
			cmdConfig = &appConfig.Commands[i]
			break
		}
	}

	if cmdConfig == nil {
		output = "Unknown command: " + cmdRequest.Command
		success = false
	} else {
		// Build full command with args
		fullCommand := cmdConfig.Command
		if cmdRequest.Args != "" {
			fullCommand = fullCommand + " " + cmdRequest.Args
		}
		// Execute the configured command with environment variables
		output, success = executeShellCommand(fullCommand, cmdConfig.Env)
	}

	// Write command execution to daemon log
	logCommandExecution(cmdRequest.Command, output, success)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": success,
		"output":  output,
	})
}

func handleLogsAPI(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	
	logs, err := readLogsWithRotation()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"logs":    "",
			"error":   err.Error(),
		})
		return
	}
	
	// Apply search filter if provided
	if search != "" {
		logs = filterLogs(logs, search)
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	// Calculate line count properly
	lineCount := 0
	if logs != "" {
		lineCount = len(strings.Split(logs, "\n"))
	}
	
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"logs":    logs,
		"count":   lineCount,
	})
}

func readLogsWithRotation() (string, error) {
	// Read existing log file
	content, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	
	// Return empty string if file is empty
	if len(content) == 0 {
		return "", nil
	}
	
	logContent := string(content)
	
	// Trim trailing whitespace for cleaner output
	logContent = strings.TrimSpace(logContent)
	
	// Return empty if only whitespace
	if logContent == "" {
		return "", nil
	}
	
	lines := strings.Split(logContent, "\n")
	
	// Rotate if exceeds max lines
	if len(lines) > maxLogLines {
		// Keep only the last maxLogLines lines
		lines = lines[len(lines)-maxLogLines:]
		rotatedContent := strings.Join(lines, "\n")
		
		// Write back to file
		if err := os.WriteFile(logFile, []byte(rotatedContent), 0644); err != nil {
			return "", err
		}
		
		return rotatedContent, nil
	}
	
	return logContent, nil
}

func filterLogs(logs, search string) string {
	if search == "" {
		return logs
	}
	
	var filteredLines []string
	lines := strings.Split(logs, "\n")
	
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(search)) {
			filteredLines = append(filteredLines, line)
		}
	}
	
	return strings.Join(filteredLines, "\n")
}

func handleClearLogsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Clear the log file
	err := os.WriteFile(logFile, []byte(""), 0644)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logs cleared successfully",
	})
}

func logCommandExecution(command, output string, success bool) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	status := "✅ SUCCESS"
	if !success {
		status = "❌ FAILED"
	}
	
	logEntry := fmt.Sprintf("[%s] %s - Command: %s\n%s\n%s\n\n", timestamp, status, command, output, strings.Repeat("-", 50))
	
	// Append to log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Error opening log file: %v", err)
		return
	}
	defer file.Close()
	
	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

func handleConfigAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appConfig)
}

func handleConfigReloadAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := loadConfig(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Config reloaded successfully",
		"config":  appConfig,
	})
}

func handleCommandsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(appConfig.Commands)
}

func handleAddCommandAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmdRequest struct {
		ID              string            `json:"id"`
		Name            string            `json:"name"`
		Description     string            `json:"description"`
		Icon            string            `json:"icon"`
		Command         string            `json:"command"`
		URL             string            `json:"url"`
		Type            string            `json:"type"`
		Category        string            `json:"category"`
		SupportsArgs    bool              `json:"supports_args"`
		ArgsDescription string            `json:"args_description"`
		Env             string            `json:"env"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cmdRequest); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
		return
	}

	if cmdRequest.ID == "" || cmdRequest.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "id and name are required",
		})
		return
	}

	// Validate type-specific fields
	cmdType := cmdRequest.Type
	if cmdType == "" {
		cmdType = "command" // Default to command
	}

	if cmdType == "command" && cmdRequest.Command == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "command is required for type 'command'",
		})
		return
	}

	if cmdType == "link" && cmdRequest.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "url is required for type 'link'",
		})
		return
	}

	// Check if ID already exists
	for _, cmd := range appConfig.Commands {
		if cmd.ID == cmdRequest.ID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Command with this ID already exists",
			})
			return
		}
	}

	// Add new command
	newCommand := Command{
		ID:              cmdRequest.ID,
		Name:            cmdRequest.Name,
		Description:     cmdRequest.Description,
		Icon:            cmdRequest.Icon,
		Command:         cmdRequest.Command,
		URL:             cmdRequest.URL,
		Type:            cmdType,
		Category:        cmdRequest.Category,
		SupportsArgs:    cmdRequest.SupportsArgs,
		ArgsDescription: cmdRequest.ArgsDescription,
		Env:             parseEnvVars(cmdRequest.Env),
	}

	appConfig.Commands = append(appConfig.Commands, newCommand)

	// Save to file
	if err := saveConfigFile(&appConfig); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Command added successfully",
		"command": newCommand,
	})
}

func handleEditCommandAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmdRequest struct {
		ID              string            `json:"id"`
		Name            string            `json:"name"`
		Description     string            `json:"description"`
		Icon            string            `json:"icon"`
		Command         string            `json:"command"`
		URL             string            `json:"url"`
		Type            string            `json:"type"`
		Category        string            `json:"category"`
		SupportsArgs    bool              `json:"supports_args"`
		ArgsDescription string            `json:"args_description"`
		Env             string            `json:"env"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cmdRequest); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
		return
	}

	if cmdRequest.ID == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	// Find and update command
	found := false
	for i, cmd := range appConfig.Commands {
		if cmd.ID == cmdRequest.ID {
			found = true
			if cmdRequest.Name != "" {
				appConfig.Commands[i].Name = cmdRequest.Name
			}
			if cmdRequest.Description != "" {
				appConfig.Commands[i].Description = cmdRequest.Description
			}
			if cmdRequest.Icon != "" {
				appConfig.Commands[i].Icon = cmdRequest.Icon
			}
			if cmdRequest.Command != "" {
				appConfig.Commands[i].Command = cmdRequest.Command
			}
			if cmdRequest.URL != "" {
				appConfig.Commands[i].URL = cmdRequest.URL
			}
			if cmdRequest.Type != "" {
				appConfig.Commands[i].Type = cmdRequest.Type
			}
			if cmdRequest.Category != "" {
				appConfig.Commands[i].Category = cmdRequest.Category
			}
			appConfig.Commands[i].SupportsArgs = cmdRequest.SupportsArgs
			if cmdRequest.ArgsDescription != "" {
				appConfig.Commands[i].ArgsDescription = cmdRequest.ArgsDescription
			}
			if cmdRequest.Env != "" {
				appConfig.Commands[i].Env = parseEnvVars(cmdRequest.Env)
			}
			break
		}
	}

	if !found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Command not found",
		})
		return
	}

	// Save to file
	if err := saveConfigFile(&appConfig); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Command updated successfully",
	})
}

func handleRemoveCommandAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cmdRequest struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&cmdRequest); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request format",
		})
		return
	}

	if cmdRequest.ID == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	// Find and remove command
	found := false
	var updatedCommands []Command
	for _, cmd := range appConfig.Commands {
		if cmd.ID == cmdRequest.ID {
			found = true
		} else {
			updatedCommands = append(updatedCommands, cmd)
		}
	}

	if !found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Command not found",
		})
		return
	}

	appConfig.Commands = updatedCommands

	// Save to file
	if err := saveConfigFile(&appConfig); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Command removed successfully",
	})
}

func handleRawConfigAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Return current config as formatted JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(appConfig)
		return
	}

	if r.Method == http.MethodPost {
		// Save raw config
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Invalid JSON format: " + err.Error(),
			})
			return
		}

		// Validate config
		if newConfig.Title == "" {
			newConfig.Title = "Command Center"
		}
		if newConfig.Subtitle == "" {
			newConfig.Subtitle = "Generic command execution dashboard"
		}

		// Save to file
		if err := saveConfigFile(&newConfig); err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Reload config
		appConfig = newConfig

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Config saved and reloaded successfully",
			"config":  appConfig,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func executeShellCommand(command string, envVars map[string]string) (string, bool) {
	cmd := exec.Command("bash", "-c", command)
	
	// Set environment variables
	if len(envVars) > 0 {
		cmd.Env = append(os.Environ())
		for key, value := range envVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Sprintf("Error executing command: %v\nOutput: %s", err, string(output)), false
	}
	
	return string(output), true
}