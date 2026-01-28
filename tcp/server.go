package tcp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/devuser/MemoryLogMonitor/logstore"
)

// Server represents a TCP log receiver server
type Server struct {
	port     int
	store    *logstore.LogStore
	listener net.Listener
}

// NewServer creates a new TCP server
func NewServer(port int, store *logstore.LogStore) *Server {
	return &Server{
		port:  port,
		store: store,
	}
}

// Start starts the TCP server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}
	s.listener = listener

	log.Printf("TCP server listening on port %d", s.port)

	go s.acceptConnections()
	return nil
}

// Stop stops the TCP server
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := scanner.Text()
		if message != "" {
			entry := logstore.LogEntry{
				Timestamp: time.Now(),
				Message:   message,
			}
			s.store.Add(entry)
			log.Printf("Received log: %s", message)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from connection: %v", err)
	}
}
