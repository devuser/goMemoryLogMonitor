package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	go func() {
		log.Printf("HTTP server starting on port %d", HTTPPort)
		if err := httpServer.Start(); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	tcpServer.Stop()
}
