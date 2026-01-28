package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpserver "github.com/devuser/MemoryLogMonitor/http"
	"github.com/devuser/MemoryLogMonitor/logstore"
	tcpserver "github.com/devuser/MemoryLogMonitor/tcp"
)

const (
	// MaxLogEntries is the maximum number of log entries to store
	MaxLogEntries = 10000
	// TCPPort is the port for the TCP log receiver
	TCPPort = 9090
	// HTTPPort is the port for the HTTP server
	HTTPPort = 8080
)

func main() {
	log.Println("Starting MemoryLogMonitor...")

	// Create log store
	store := logstore.NewLogStore(MaxLogEntries)

	// Start TCP server
	tcpServer := tcpserver.NewServer(TCPPort, store)
	if err := tcpServer.Start(); err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}

	// Start HTTP server in a goroutine
	httpServer := httpserver.NewServer(HTTPPort, store)
	
	// Create HTTP server with proper shutdown support
	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpServer.GetRouter(),
	}
	
	go func() {
		log.Printf("HTTP server starting on port %d", HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	
	// Gracefully shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	
	// Stop TCP server
	tcpServer.Stop()
	
	log.Println("Shutdown complete")
}
