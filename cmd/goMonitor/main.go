package main

import (
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"goMonitor/config"
	"goMonitor/logstore"
	"goMonitor/tcpserver"
	"goMonitor/web"
)

func main() {
	configPath := flag.String("config", "config.yml", "path to config YAML file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize in-memory log store
	store := logstore.NewStore(int64(cfg.CacheSizeMB) * 1024 * 1024)

	// Start TCP log receiver
	go func() {
		if err := tcpserver.Run(cfg.TCPPort, store); err != nil {
			log.Fatalf("tcp server error: %v", err)
		}
	}()

	// Setup Gin HTTP server
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		api.GET("/status", func(c *gin.Context) {
			// Get memory stats
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			// Get log store stats
			entries, cacheBytes := store.Snapshot()
			logCount := len(entries)

			c.JSON(http.StatusOK, gin.H{
				"memoryMB":     float64(m.Alloc) / 1024 / 1024,
				"logCount":      logCount,
				"httpPort":      cfg.HTTPPort,
				"tcpPort":       cfg.TCPPort,
				"cacheSizeMB":   float64(cacheBytes) / 1024 / 1024,
			})
		})

		api.DELETE("/logs", func(c *gin.Context) {
			store.Clear()
			c.JSON(http.StatusOK, gin.H{
				"message": "Logs cleared successfully",
			})
		})

		api.GET("/logs", func(c *gin.Context) {
			// Query parameters
			page := parseIntQuery(c, "page", 1)
			pageSize := parseIntQuery(c, "pageSize", 50)
			if page < 1 {
				page = 1
			}
			if pageSize <= 0 || pageSize > 1000 {
				pageSize = 50
			}

			startDateStr := c.Query("startDate")
			endDateStr := c.Query("endDate")
			q := c.Query("q")
			topN := parseIntQuery(c, "topN", 0)
			sortBy := c.Query("sortBy")
			sortOrder := c.Query("sortOrder")

			var (
				startTime *time.Time
				endTime   *time.Time
			)

			const dateLayout = "2006-01-02"
			if startDateStr != "" {
				if t, err := time.Parse(dateLayout, startDateStr); err == nil {
					startTime = &t
				}
			}
			if endDateStr != "" {
				if t, err := time.Parse(dateLayout, endDateStr); err == nil {
					// include the whole end day
					t = t.Add(24 * time.Hour)
					endTime = &t
				}
			}

			// Validate sortBy and sortOrder
			if sortBy != "time" && sortBy != "content" {
				sortBy = "" // Use default sorting
			}
			if sortOrder != "asc" && sortOrder != "desc" {
				sortOrder = "desc" // Default to descending
			}

			result := store.Query(logstore.QueryOptions{
				Page:      page,
				PageSize:  pageSize,
				StartTime: startTime,
				EndTime:   endTime,
				Text:      q,
				TopN:      topN,
				SortBy:    sortBy,
				SortOrder: sortOrder,
			})

			c.JSON(http.StatusOK, gin.H{
				"items":          result.Items,
				"total":          result.Total,
				"page":           page,
				"pageSize":       pageSize,
				"cacheCount":     result.CacheCount,
				"cacheSizeBytes": result.CacheSizeBytes,
			})
		})
	}

	// Serve embedded frontend
	web.RegisterStatic(r)

	addr := cfg.HTTPAddr()
	log.Printf("HTTP server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}

func parseIntQuery(c *gin.Context, key string, def int) int {
	if v := c.Query(key); v != "" {
		var parsed int
		if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil {
			return parsed
		}
	}
	return def
}
