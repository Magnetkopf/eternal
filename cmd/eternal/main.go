package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/Magnetkopf/Eternal/internal/config"
	"github.com/Magnetkopf/Eternal/internal/ipc"
)

const socketPath = "/tmp/eternal.sock"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: eternal [start|stop|restart|status|enable|disable|new|delete] <service_name>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	service := os.Args[2]

	var reqType ipc.RequestType
	switch cmd {
	case "start":
		reqType = ipc.RequestStart
	case "stop":
		reqType = ipc.RequestStop
	case "status":
		reqType = ipc.RequestStatus
	case "restart":
		reqType = ipc.RequestRestart
	case "enable":
		handleEnable(service)
		return
	case "disable":
		handleDisable(service)
		return
	case "new":
		handleNew(service)
		return
	case "delete":
		handleDelete(service)
		return
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}

	sendRequest(reqType, service)
}

func handleEnable(service string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home: %v\n", err)
		os.Exit(1)
	}

	serviceFile := filepath.Join(home, ".eternal", "services", service+".yaml")
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		fmt.Printf("Service definition %s not found\n", serviceFile)
		os.Exit(1)
	}

	enabledFile := filepath.Join(home, ".eternal", "enabled.yaml")

	if err := config.EnableService(enabledFile, service); err != nil {
		fmt.Printf("Failed to enable service: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Service %s enabled\n", service)
}

func handleDisable(service string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home: %v\n", err)
		os.Exit(1)
	}
	enabledFile := filepath.Join(home, ".eternal", "enabled.yaml")

	if err := config.DisableService(enabledFile, service); err != nil {
		fmt.Printf("Failed to disable service: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Service %s disabled\n", service)
}

func handleNew(service string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home: %v\n", err)
		os.Exit(1)
	}

	servicesDir := filepath.Join(home, ".eternal", "services")
	if err := os.MkdirAll(servicesDir, 0755); err != nil {
		fmt.Printf("Failed to create services directory: %v\n", err)
		os.Exit(1)
	}

	serviceFile := filepath.Join(servicesDir, service+".yaml")
	if _, err := os.Stat(serviceFile); err == nil {
		fmt.Printf("Service %s already exists at %s\n", service, serviceFile)
		os.Exit(1)
	}

	defaultContent := `# Command to execute
exec: ""
# Working directory
dir: ""
`
	if err := os.WriteFile(serviceFile, []byte(defaultContent), 0644); err != nil {
		fmt.Printf("Failed to create service file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service created. Edit config at: %s\n", serviceFile)
}

func handleDelete(service string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home: %v\n", err)
		os.Exit(1)
	}

	// 1. Check if service file exists
	serviceFile := filepath.Join(home, ".eternal", "services", service+".yaml")
	if _, err := os.Stat(serviceFile); os.IsNotExist(err) {
		fmt.Printf("Service %s does not exist\n", service)
		os.Exit(1)
	}

	// 2. Disable service if enabled
	enabledFile := filepath.Join(home, ".eternal", "enabled.yaml")
	// config.DisableService returns nil if service is not found in enabled list, which is what we want
	if err := config.DisableService(enabledFile, service); err != nil {
		fmt.Printf("Failed to disable service before deletion: %v\n", err)
		os.Exit(1)
	}

	// 3. Delete service file
	if err := os.Remove(serviceFile); err != nil {
		fmt.Printf("Failed to delete service file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Service %s deleted\n", service)
}

func sendRequest(reqType ipc.RequestType, service string) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Printf("Failed to connect to daemon: %v\n", err)
		fmt.Println("Is eternal-daemon running?")
		os.Exit(1)
	}
	defer conn.Close()

	req := ipc.Request{
		Type:    reqType,
		Service: service,
	}

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		os.Exit(1)
	}

	var resp ipc.Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		os.Exit(1)
	}

	if resp.Success {
		fmt.Println(resp.Message)
	} else {
		fmt.Printf("Error: %s\n", resp.Message)
		os.Exit(1)
	}
}
