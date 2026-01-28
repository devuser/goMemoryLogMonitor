package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"

	"goMonitor/logstore"
)

func Run(port int, store *logstore.Store) error {
	addr := fmt.Sprintf(":%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	log.Printf("TCP log server listening on %s", addr)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		go handleConn(conn, store)
	}
}

func handleConn(c net.Conn, store *logstore.Store) {
	defer c.Close()

	remote := c.RemoteAddr().String()
	log.Printf("new log connection from %s", remote)

	scanner := bufio.NewScanner(c)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r\n")
		if line == "" {
			continue
		}
		store.Append(line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("connection %s read error: %v", remote, err)
	}
}
