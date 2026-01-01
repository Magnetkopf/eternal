package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"eternal/pkg/config"
	"eternal/pkg/ipc"
)

const socketPath = "/tmp/eternal.sock"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: eternal [start|stop|status|enable|disable] <service_name>")
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
	case "enable":
		handleEnable(service)
		return
	case "disable":
		handleDisable(service)
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
