package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Magnetkopf/Eternal/internal/api"
	"github.com/Magnetkopf/Eternal/internal/config"
	"github.com/Magnetkopf/Eternal/internal/ipc"
	"github.com/Magnetkopf/Eternal/internal/process"
)

const socketPath = "/tmp/eternal.sock"
const servicesDir = "./services"

func main() {
	// 1. Initialize Process Manager
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get user home directory: %v", err)
	}
	baseDir := filepath.Join(home, ".eternal")
	servicesDir := filepath.Join(baseDir, "services")
	enabledFile := filepath.Join(baseDir, "enabled.yaml")

	// Ensure directories exist
	if err := os.MkdirAll(servicesDir, 0755); err != nil {
		log.Fatalf("Failed to create services directory: %v", err)
	}

	pm := process.NewManager(servicesDir)
	if err := pm.LoadServices(); err != nil {
		log.Printf("Warning: Failed to load some services: %v", err)
	}

	// Auto-start enabled services
	enabledServices, err := config.LoadEnabledServices(enabledFile)
	if err != nil {
		log.Printf("Warning: Failed to load enabled services: %v", err)
	} else {
		for _, name := range enabledServices {
			if err := pm.StartService(name); err != nil {
				log.Printf("Failed to auto-start service %s: %v", name, err)
			} else {
				log.Printf("Auto-started service %s", name)
			}
		}
	}

	// 2. Setup Socket
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to remove old socket: %v", err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on socket: %v", err)
	}
	defer listener.Close()

	// 3. Handle Signals for Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down...")
		// Stop all services? For now, we trust the OS or Process Manager if we added StopAll method.
		// Since we didn't add StopAll, we might leave them running or kill them.
		// Systemd leaves them running usually, but eternal-daemon maybe should kill them if it's the parent.
		// Let's just exit for now.
		// Ideally: pm.StopAll()
		os.Exit(0)
	}()

	log.Println("Eternal Daemon started, listening on", socketPath)

	// Load Auth Token
	configFile := filepath.Join(baseDir, "config.yaml")
	authToken, err := config.LoadOrGenerateToken(configFile)
	if err != nil {
		log.Fatalf("Failed to load auth token: %v", err)
	}
	log.Printf("Auth token: %s", authToken)

	// Start API Server
	go api.StartServer(pm, 9093, servicesDir, enabledFile, authToken)

	// 4. Accept Connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn, pm)
	}
}

func handleConnection(conn net.Conn, pm *process.Manager) {
	defer conn.Close()

	var req ipc.Request
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	if err := decoder.Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		return
	}

	var resp ipc.Response

	switch req.Type {
	case ipc.RequestStart:
		err := pm.StartService(req.Service)
		if err != nil {
			resp.Success = false
			resp.Message = err.Error()
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("Service %s started", req.Service)
		}
	case ipc.RequestStop:
		err := pm.StopService(req.Service)
		if err != nil {
			resp.Success = false
			resp.Message = err.Error()
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("Service %s stopped", req.Service)
		}
	case ipc.RequestStatus:
		status, err := pm.GetStatus(req.Service)
		if err != nil {
			resp.Success = false
			resp.Message = err.Error()
		} else {
			resp.Success = true
			resp.Message = string(status)
		}
	case ipc.RequestRestart:
		err := pm.RestartService(req.Service)
		if err != nil {
			resp.Success = false
			resp.Message = err.Error()
		} else {
			resp.Success = true
			resp.Message = fmt.Sprintf("Service %s restarted", req.Service)
		}
	default:
		resp.Success = false
		resp.Message = "Unknown request type"
	}

	if err := encoder.Encode(resp); err != nil {
		log.Printf("Failed to send response: %v", err)
	}
}
