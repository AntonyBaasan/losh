package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// Session represents an active client connection and its associated routing information.
type Session struct {
	ID        string
	Subdomain string
	Control   net.Conn
}

// Server handles control connections, data connections, and HTTP routing.
type Server struct {
	mu         sync.Mutex
	sessions   map[string]*Session     // Maps session_id -> Session
	subdomains map[string]string       // Maps subdomain -> session_id
	pending    map[string]chan net.Conn // Maps data_conn_id -> channel for pairing
}

// NewServer initializes a new Server instance.
func NewServer() *Server {
	return &Server{
		sessions:   make(map[string]*Session),
		subdomains: make(map[string]string),
		pending:    make(map[string]chan net.Conn),
	}
}

// generateUUID creates a simple pseudo-random UUID.
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// generateSubdomain creates a short random hex string for the subdomain.
func generateSubdomain() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Run starts the listener for control and data connections.
func (s *Server) Run(controlAddr string) error {
	ln, err := net.Listen("tcp", controlAddr)
	if err != nil {
		return err
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("accept control/data error:", err)
			continue
		}
		go s.handleConn(c)
	}
}

// handleConn determines whether a new connection is a control or data connection.
func (s *Server) handleConn(c net.Conn) {
	r := bufio.NewReader(c)
	line, err := r.ReadString('\n')
	if err != nil {
		log.Println("read handshake error:", err)
		c.Close()
		return
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")

	// CONTROL <session_id>
	// If session_id is empty or unknown, a new one is assigned.
	if len(parts) >= 1 && parts[0] == "CONTROL" {
		sessionID := ""
		if len(parts) >= 2 {
			sessionID = parts[1]
		}
		s.handleControlConn(sessionID, c, r)
		return
	}

	// DATA <session_id> <data_conn_id>
	if len(parts) >= 3 && parts[0] == "DATA" {
		sessionID := parts[1]
		dataConnID := parts[2]
		s.handleDataConn(sessionID, dataConnID, c)
		return
	}

	log.Println("unknown handshake line:", line)
	c.Close()
}

// handleControlConn registers the control connection for a session.
func (s *Server) handleControlConn(sessionID string, c net.Conn, r *bufio.Reader) {
	s.mu.Lock()
	sess, exists := s.sessions[sessionID]
	
	// Create new session if none exists or session ID was empty
	if !exists || sessionID == "" {
		sess = &Session{
			ID:        generateUUID(),
			Subdomain: generateSubdomain(),
			Control:   c,
		}
		s.sessions[sess.ID] = sess
		s.subdomains[sess.Subdomain] = sess.ID
		log.Printf("Created new session %s with subdomain %s", sess.ID, sess.Subdomain)
	} else {
		// Update existing session's control connection
		sess.Control = c
		log.Printf("Resumed session %s with subdomain %s", sess.ID, sess.Subdomain)
	}
	s.mu.Unlock()

	// Send the assigned session and subdomain back to the client
	fmt.Fprintf(c, "SESSION %s %s\n", sess.ID, sess.Subdomain)

	// Keep reading to detect disconnects
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		// We don't expect the client to send anything else on the control connection,
		// but we read to keep the connection alive and detect when it drops.
	}

	// Cleanup on disconnect
	s.mu.Lock()
	if currentSess, ok := s.sessions[sess.ID]; ok && currentSess.Control == c {
		currentSess.Control = nil // Mark as disconnected, but keep session alive for resume
		log.Printf("Control connection for session %s disconnected", sess.ID)
	}
	s.mu.Unlock()
	c.Close()
}

// handleDataConn pairs an incoming data connection with a pending request.
func (s *Server) handleDataConn(sessionID, dataConnID string, c net.Conn) {
	s.mu.Lock()
	ch, ok := s.pending[dataConnID]
	s.mu.Unlock()

	if ok {
		select {
		case ch <- c:
			// Successfully paired
		default:
			// Channel was full or closed
			c.Close()
		}
	} else {
		// No pending request for this data connection
		log.Printf("No pending request for data conn ID %s", dataConnID)
		c.Close()
	}
}

// requestDataConn sends a NEW command to the client to establish a fresh data connection.
func (s *Server) requestDataConn(sessionID string) (net.Conn, error) {
	dataConnID := generateUUID()
	ch := make(chan net.Conn, 1)
	
	s.mu.Lock()
	s.pending[dataConnID] = ch
	sess, ok := s.sessions[sessionID]
	var ctrl net.Conn
	if ok && sess != nil {
		ctrl = sess.Control
	}
	s.mu.Unlock()

	if ctrl == nil {
		s.cleanupPending(dataConnID)
		return nil, fmt.Errorf("client not connected")
	}

	// Ask the client to dial back with a data connection
	_, err := fmt.Fprintf(ctrl, "NEW %s\n", dataConnID)
	if err != nil {
		s.cleanupPending(dataConnID)
		return nil, fmt.Errorf("failed to notify control: %v", err)
	}

	// Wait for the data connection
	select {
	case dataConn := <-ch:
		s.cleanupPending(dataConnID)
		return dataConn, nil
	case <-time.After(30 * time.Second):
		s.cleanupPending(dataConnID)
		return nil, fmt.Errorf("timeout waiting for data connection")
	}
}

// cleanupPending removes the pending data connection channel.
func (s *Server) cleanupPending(dataConnID string) {
	s.mu.Lock()
	delete(s.pending, dataConnID)
	s.mu.Unlock()
}

// proxyConnections bidirectionally copies data between two connections.
func (s *Server) proxyConnections(client net.Conn, backend net.Conn, clientReader *bufio.Reader) {
	// If we read some bytes from the client to parse headers, we need to replay them
	// to the backend first.
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		io.Copy(pw, clientReader)
	}()
	
	// Create a MultiReader that first reads the buffered bytes, then from the raw client connection.
	mr := io.MultiReader(pr, client)

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(backend, mr)
		backend.Close()
		done <- struct{}{}
	}()
	go func() {
		io.Copy(client, backend)
		client.Close()
		done <- struct{}{}
	}()
	<-done
	<-done
}

// listenHTTP starts the HTTP router listener.
func (s *Server) listenHTTP(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("http listener error: %v", err)
		return
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("accept http error:", err)
			continue
		}
		go s.handleHTTPConn(c)
	}
}

// handleHTTPConn routes an incoming HTTP request to the appropriate client session.
func (s *Server) handleHTTPConn(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	
	// Read headers to extract the Host
	var buf []byte
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			log.Println("read http headers error:", err)
			return
		}
		buf = append(buf, line...)
		if bytes.HasSuffix(buf, []byte("\r\n\r\n")) || bytes.HasSuffix(buf, []byte("\n\n")) {
			break
		}
		if len(buf) > 64*1024 {
			log.Println("http headers too large")
			return
		}
	}
	
	headers := string(buf)
	host := ""
	for _, line := range strings.Split(headers, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(strings.ToLower(line), "host:") {
			host = strings.TrimSpace(line[5:])
			break
		}
	}

	if host == "" {
		fmt.Fprintf(c, "HTTP/1.1 400 Bad Request\r\n\r\nNo Host header found.\n")
		return
	}

	// Extract subdomain from the host
	subdomain := host
	if strings.Contains(subdomain, ":") {
		subdomain = strings.Split(subdomain, ":")[0]
	}
	if strings.Contains(subdomain, ".") {
		subdomain = strings.Split(subdomain, ".")[0]
	}

	// Look up the session ID
	s.mu.Lock()
	sessionID, ok := s.subdomains[subdomain]
	s.mu.Unlock()

	if !ok {
		fmt.Fprintf(c, "HTTP/1.1 404 Not Found\r\n\r\nTunnel %s not found.\n", host)
		return
	}

	// Request a data connection from the client
	dataConn, err := s.requestDataConn(sessionID)
	if err != nil {
		fmt.Fprintf(c, "HTTP/1.1 502 Bad Gateway\r\n\r\nFailed to establish backend connection: %v\n", err)
		return
	}

	// Replay the headers and proxy the rest of the connection directly to the client
	// We pass a new reader containing only the headers we buffered
	headerReader := bufio.NewReader(bytes.NewReader(buf))
	s.proxyConnections(c, dataConn, headerReader)
}
