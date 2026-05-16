package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	server *http.Server
	mu     sync.Mutex
)

type Status struct {
	Status    string    `json:"status"`
	Port      int       `json:"port"`
	Uptime    string    `json:"uptime"`
	StartTime time.Time `json:"start_time"`
}

var serverStatus Status

func startServer(port int) {
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

func handleHome(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Boilerplate CLI UI</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 40px;
            max-width: 500px;
            width: 100%;
        }
        h1 { 
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 14px;
        }
        .status-card {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .status-item {
            display: flex;
            justify-content: space-between;
            padding: 10px 0;
            border-bottom: 1px solid #e9ecef;
        }
        .status-item:last-child {
            border-bottom: none;
        }
        .label { color: #666; font-size: 14px; }
        .value { color: #333; font-weight: 600; font-size: 14px; }
        .status-running { color: #28a745; }
        .status-stopped { color: #dc3545; }
        .btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 16px;
            font-weight: 600;
            width: 100%;
            transition: transform 0.2s;
        }
        .btn:hover { transform: translateY(-2px); }
        .btn:active { transform: translateY(0); }
        .api-section {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e9ecef;
        }
        .api-section h3 {
            color: #333;
            margin-bottom: 15px;
            font-size: 16px;
        }
        .api-endpoint {
            background: #f8f9fa;
            padding: 10px;
            border-radius: 4px;
            margin-bottom: 8px;
            font-family: monospace;
            font-size: 12px;
            color: #495057;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Boilerplate CLI UI</h1>
        <p class="subtitle">Go CLI with HTTP server and simple UI</p>
        
        <div class="status-card">
            <div class="status-item">
                <span class="label">Status</span>
                <span class="value status-running">Running</span>
            </div>
            <div class="status-item">
                <span class="label">Port</span>
                <span class="value">8080</span>
            </div>
            <div class="status-item">
                <span class="label">Version</span>
                <span class="value">1.0.0</span>
            </div>
        </div>
        
        <button class="btn" onclick="refreshStatus()">Refresh Status</button>
        
        <div class="api-section">
            <h3>API Endpoints</h3>
            <div class="api-endpoint">GET /api/status - Server status</div>
            <div class="api-endpoint">GET /api/health - Health check</div>
        </div>
    </div>
    
    <script>
        function refreshStatus() {
            fetch('/api/status')
                .then(r => r.json())
                .then(data => {
                    console.log('Status:', data);
                })
                .catch(err => console.error('Error:', err));
        }
        
        // Auto-refresh every 5 seconds
        setInterval(refreshStatus, 5000);
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