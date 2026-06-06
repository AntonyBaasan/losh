package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// proxy bidirectionally copies data between two connections.
func proxy(a, b net.Conn) {
	done := make(chan struct{}, 2)
	go func() {
		io.Copy(a, b)
		a.Close()
		done <- struct{}{}
	}()
	go func() {
		io.Copy(b, a)
		b.Close()
		done <- struct{}{}
	}()
	<-done
	<-done
}

// handleControl connects to the server, establishes a session, and listens for data connection requests.
func handleControl(serverAddr, sessionID, localAddr string) (string, error) {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		return sessionID, err
	}
	defer conn.Close()

	// 1. Send CONTROL handshake
	fmt.Fprintf(conn, "CONTROL %s\n", sessionID)

	r := bufio.NewReader(conn)

	// 2. Read SESSION response
	line, err := r.ReadString('\n')
	if err != nil {
		return sessionID, fmt.Errorf("failed to read session response: %v", err)
	}
	line = strings.TrimSpace(line)
	parts := strings.Split(line, " ")
	if len(parts) != 3 || parts[0] != "SESSION" {
		return sessionID, fmt.Errorf("unexpected server response: %s", line)
	}

	assignedSessionID := parts[1]
	subdomain := parts[2]

	log.Printf("=====================================================")
	log.Printf("Tunnel established successfully!")
	log.Printf("Session ID: %s", assignedSessionID)
	log.Printf("Public URL: http://%s.localhost:8080", subdomain)
	log.Printf("Forwarding: http://%s.localhost:8080 -> %s", subdomain, localAddr)
	log.Printf("=====================================================")

	// 3. Listen for NEW connection requests
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return assignedSessionID, err
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")

		if len(parts) == 2 && parts[0] == "NEW" {
			dataConnID := parts[1]

			// Handle each new data connection request in a goroutine
			go func(id string) {
				// Dial the server again to establish the data channel
				dc, err := net.Dial("tcp", serverAddr)
				if err != nil {
					log.Println("dial data error:", err)
					return
				}

				// Identify this data connection to the server
				fmt.Fprintf(dc, "DATA %s %s\n", assignedSessionID, id)

				// Dial the local service
				local, err := net.Dial("tcp", localAddr)
				if err != nil {
					log.Println("dial local error:", err)
					dc.Close()
					return
				}

				// Proxy traffic between the local service and the server data connection
				proxy(local, dc)
			}(dataConnID)
		} else {
			log.Println("control unknown command:", line)
		}
	}
}

func main() {
	var serverAddr string
	var sessionID string
	var localAddr string

	flag.StringVar(&serverAddr, "server", "localhost:7001", "server control/data address")
	flag.StringVar(&sessionID, "session", "", "Session ID to resume a previous tunnel (optional)")
	flag.StringVar(&localAddr, "local", "localhost:8080", "local service address to forward to")
	flag.Parse()

	log.Println("Starting client...")

	for {
		var err error
		// We re-assign the sessionID in case the server generated a new one for us.
		// That way, if the connection drops, we reconnect using the newly assigned session ID
		// and retain our assigned subdomain.
		sessionID, err = handleControl(serverAddr, sessionID, localAddr)
		if err != nil {
			log.Println("control loop error:", err)
		}
		log.Println("Reconnecting in 2 seconds...")
		time.Sleep(2 * time.Second)
	}
}
