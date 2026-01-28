package http

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/devuser/MemoryLogMonitor/logstore"
	"github.com/gin-gonic/gin"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// Server represents an HTTP server
type Server struct {
	port   int
	store  *logstore.LogStore
	router *gin.Engine
}

// NewServer creates a new HTTP server
func NewServer(port int, store *logstore.LogStore) *Server {
	return &Server{
		port:  port,
		store: store,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.router = gin.Default()
	s.setupRoutes()
	return s.router.Run(":" + strconv.Itoa(s.port))
}

func (s *Server) setupRoutes() {
	// API routes
	api := s.router.Group("/api")
	{
		api.GET("/logs", s.getLogs)
		api.GET("/status", s.getStatus)
	}

	// Serve frontend
	s.serveFrontend()
}

func (s *Server) getLogs(c *gin.Context) {
	// Get query parameters
	keyword := c.Query("keyword")
	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")
	sortOrder := c.DefaultQuery("sort", "desc") // desc = newest first, asc = oldest first
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "100")

	// Parse pagination parameters
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = 1
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum < 1 || pageSizeNum > 1000 {
		pageSizeNum = 100
	}

	// Parse time filters
	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = &t
		}
	}
	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	// Get filtered logs
	logs := s.store.GetFiltered(keyword, startTime, endTime)

	// Sort logs
	if sortOrder == "asc" {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Timestamp.Before(logs[j].Timestamp)
		})
	} else {
		sort.Slice(logs, func(i, j int) bool {
			return logs[i].Timestamp.After(logs[j].Timestamp)
		})
	}

	// Apply pagination
	total := len(logs)
	start := (pageNum - 1) * pageSizeNum
	end := start + pageSizeNum

	if start >= total {
		logs = []logstore.LogEntry{}
	} else {
		if end > total {
			end = total
		}
		logs = logs[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"page":     pageNum,
		"pageSize": pageSizeNum,
	})
}

func (s *Server) getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"logCount":  s.store.Count(),
		"timestamp": time.Now(),
	})
}

func (s *Server) serveFrontend() {
	// Try to serve embedded frontend
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err == nil {
		// Use NoRoute for all non-API routes
		s.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			
			// Default to index.html for root
			if path == "/" {
				path = "/index.html"
			}
			
			// Remove leading slash
			cleanPath := strings.TrimPrefix(path, "/")
			
			// Try to open the file
			file, err := distFS.Open(cleanPath)
			if err != nil {
				// File not found, serve index.html for SPA routing
				file, err = distFS.Open("index.html")
				if err != nil {
					c.String(http.StatusNotFound, "404 not found")
					return
				}
				cleanPath = "index.html"
			}
			defer file.Close()
			
			// Read file content
			content, err := io.ReadAll(file)
			if err != nil {
				c.String(http.StatusInternalServerError, "Error reading file")
				return
			}
			
			// Determine content type
			contentType := "text/html; charset=utf-8"
			if strings.HasSuffix(cleanPath, ".js") {
				contentType = "application/javascript; charset=utf-8"
			} else if strings.HasSuffix(cleanPath, ".css") {
				contentType = "text/css; charset=utf-8"
			} else if strings.HasSuffix(cleanPath, ".svg") {
				contentType = "image/svg+xml"
			}
			
			c.Data(http.StatusOK, contentType, content)
		})
	} else {
		// Fallback: serve a simple HTML page if frontend not built
		s.router.GET("/", s.serveFallbackHTML)
	}
}

func (s *Server) serveFallbackHTML(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>MemoryLogMonitor</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #333; }
        .info { background: #f0f0f0; padding: 20px; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>MemoryLogMonitor</h1>
        <div class="info">
            <p>Frontend not built. Please build the frontend first:</p>
            <pre>cd frontend && npm install && npm run build</pre>
            <p>API endpoints are available:</p>
            <ul>
                <li>GET /api/logs - Get logs with optional filters</li>
                <li>GET /api/status - Get system status</li>
            </ul>
        </div>
    </div>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
