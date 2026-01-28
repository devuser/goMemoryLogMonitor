package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9090", "server TCP address")
	flag.Parse()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup
	wg.Add(1)

	// Start sending logs in background with retry logic
	go func() {
		defer wg.Done()
		runWithRetry(ctx, *addr)
	}()

	// Wait for interrupt signal
	<-sigChan
	fmt.Println("\nReceived shutdown signal, stopping gracefully...")

	// Cancel context to stop sending logs
	cancel()

	// Wait for goroutine to finish (with timeout)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("Gracefully stopped")
	case <-time.After(5 * time.Second):
		fmt.Println("Timeout waiting for graceful shutdown")
	}
}

func runWithRetry(ctx context.Context, addr string) {
	totalSent := 0
	firstConnect := true

	for {
		// Try to connect
		conn, err := connectWithRetry(ctx, addr, firstConnect)
		if err != nil {
			// Context cancelled, exit
			if ctx.Err() != nil {
				fmt.Printf("\nTotal logs sent: %d\n", totalSent)
				return
			}
			// Connection failed, will retry in the loop
			continue
		}

		firstConnect = false
		fmt.Printf("Connected to %s, sending logs (50 logs every 5s)...\n", addr)
		if totalSent == 0 {
			fmt.Println("Press Ctrl+C to stop gracefully")
		}

		// Send logs with this connection
		sent := sendLogs(ctx, conn, &totalSent)
		conn.Close()

		// If context cancelled, exit
		if ctx.Err() != nil {
			fmt.Printf("\nTotal logs sent: %d\n", totalSent)
			return
		}

		// Connection lost, will retry
		if sent > 0 {
			fmt.Printf("Connection lost, will retry in 5s...\n")
		}
	}
}

func connectWithRetry(ctx context.Context, addr string, firstConnect bool) (net.Conn, error) {
	retryInterval := 5 * time.Second

	for {
		// Try to connect
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			return conn, nil
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Print error and wait before retry
		if firstConnect {
			fmt.Fprintf(os.Stderr, "Failed to connect to %s: %v\n", addr, err)
			fmt.Printf("Will retry in 5s...\n")
		} else {
			fmt.Fprintf(os.Stderr, "Reconnection to %s failed: %v\n", addr, err)
			fmt.Printf("Will retry in 5s...\n")
		}

		// Wait for retry interval or context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
			// Continue to retry
		}
	}
}

func sendLogs(ctx context.Context, conn net.Conn, totalSent *int) int {
	w := bufio.NewWriter(conn)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	initialTotal := *totalSent

	// Send initial batch immediately
	if !sendBatch(w, totalSent) {
		return *totalSent - initialTotal
	}

	for {
		select {
		case <-ctx.Done():
			// Flush any remaining data before exit
			w.Flush()
			return *totalSent - initialTotal
		case <-ticker.C:
			if !sendBatch(w, totalSent) {
				// Write error, connection may be lost
				return *totalSent - initialTotal
			}
		}
	}
}

func sendBatch(w *bufio.Writer, totalSent *int) bool {
	batchSize := 50
	startTime := time.Now()

	for i := 0; i < batchSize; i++ {
		*totalSent++
		line := fmt.Sprintf("[%s] mock log #%d level=%s msg=\"test message %d\"",
			time.Now().Format(time.RFC3339),
			*totalSent,
			randomLevel(),
			*totalSent,
		)
		if _, err := w.WriteString(line + "\n"); err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			return false
		}
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush error: %v\n", err)
		return false
	}

	elapsed := time.Since(startTime)
	fmt.Printf("[%s] Sent batch of %d logs (total: %d, elapsed: %v)\n",
		time.Now().Format("15:04:05"), batchSize, *totalSent, elapsed)
	return true
}

func randomLevel() string {
	switch rand.Intn(4) {
	case 0:
		return "INFO"
	case 1:
		return "WARN"
	case 2:
		return "ERROR"
	default:
		return "DEBUG"
	}
}
