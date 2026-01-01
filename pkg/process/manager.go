package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"eternal/pkg/config"
)

// ProcessStatus represents the state of a process
type ProcessStatus string

const (
	StatusRunning ProcessStatus = "running"
	StatusStopped ProcessStatus = "stopped"
	StatusError   ProcessStatus = "error"
)

// ManagedProcess holds the state of a single service
type ManagedProcess struct {
	Config *config.ServiceConfig
	Cmd    *exec.Cmd
	Status ProcessStatus
	Err    error
}

// Manager handles multiple services
type Manager struct {
	processes   map[string]*ManagedProcess
	mu          sync.RWMutex
	servicesDir string
}

// NewManager creates a new process manager
func NewManager(servicesDir string) *Manager {
	return &Manager{
		processes:   make(map[string]*ManagedProcess),
		servicesDir: servicesDir,
	}
}

// LoadServices scans the services directory and loads configurations
func (m *Manager) LoadServices() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	entries, err := os.ReadDir(m.servicesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No services dir yet, that's fine
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		cfgPath := filepath.Join(m.servicesDir, entry.Name())
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			fmt.Printf("Failed to load service %s: %v\n", name, err)
			continue
		}

		// Only add if not already running/exists, or update?
		// For now, simpler: just add.
		if _, exists := m.processes[name]; !exists {
			m.processes[name] = &ManagedProcess{
				Config: cfg,
				Status: StatusStopped,
			}
		}
	}
	return nil
}

// StartService starts a service by name
func (m *Manager) StartService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	if proc.Status == StatusRunning {
		return fmt.Errorf("service %s is already running", name)
	}

	// Parse command line
	parts := strings.Fields(proc.Config.Exec)
	if len(parts) == 0 {
		return fmt.Errorf("empty exec command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if proc.Config.Dir != "" {
		cmd.Dir = proc.Config.Dir
	}

	// Optional: Set stdout/stderr to something useful or /dev/null
	// For now, let's inherit or ignore. Daemon usually logs to file.
	// We'll leave it attached to nil (os.DevNull) for now.

	if err := cmd.Start(); err != nil {
		proc.Status = StatusError
		proc.Err = err
		return fmt.Errorf("failed to start: %w", err)
	}

	proc.Cmd = cmd
	proc.Status = StatusRunning
	proc.Err = nil

	// Defunct process handling:
	// In a real system, we'd want to Wait() for the process in a goroutine
	// to update status when it dies.
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()
		// Check if it's still the same process (it might have been restarted)
		if m.processes[name] == proc {
			proc.Status = StatusStopped
			if err != nil {
				proc.Err = err
				proc.Status = StatusError
			}
		}
	}()

	return nil
}

// StopService stops a service
func (m *Manager) StopService(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	proc, exists := m.processes[name]
	if !exists {
		return fmt.Errorf("service %s not found", name)
	}

	if proc.Status != StatusRunning || proc.Cmd == nil || proc.Cmd.Process == nil {
		return fmt.Errorf("service %s is not running", name)
	}

	// Try graceful stop (SIGTERM)
	// Windows doesn't support Signal, but we are on Linux.
	if runtime.GOOS != "windows" {
		if err := proc.Cmd.Process.Signal(os.Interrupt); err != nil {
			// Fallback to Kill
			proc.Cmd.Process.Kill()
		}
	} else {
		proc.Cmd.Process.Kill()
	}

	// process status update happens in the Wait() goroutine
	return nil
}

// GetStatus returns the status of a service
func (m *Manager) GetStatus(name string) (ProcessStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	proc, exists := m.processes[name]
	if !exists {
		return "", fmt.Errorf("service %s not found", name)
	}
	return proc.Status, nil
}
