package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServiceConfig represents the configuration for a service
type ServiceConfig struct {
	Exec string `yaml:"exec"`
	Dir  string `yaml:"dir"`
}

// LoadConfig loads a service configuration from a YAML file
func LoadConfig(path string) (*ServiceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ServiceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Basic validation
	if cfg.Exec == "" {
		return nil, fmt.Errorf("exec field is required")
	}

	return &cfg, nil
}

// LoadEnabledServices loads the list of enabled services from the given file
func LoadEnabledServices(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read enabled services: %w", err)
	}

	var services []string
	if err := yaml.Unmarshal(data, &services); err != nil {
		return nil, fmt.Errorf("failed to parse enabled services: %w", err)
	}

	return services, nil
}

// EnableService adds a service to the enabled list
func EnableService(path, name string) error {
	services, err := LoadEnabledServices(path)
	if err != nil {
		return err
	}

	for _, s := range services {
		if s == name {
			return nil // Already enabled
		}
	}

	services = append(services, name)
	return saveEnabledServices(path, services)
}

// DisableService removes a service from the enabled list
func DisableService(path, name string) error {
	services, err := LoadEnabledServices(path)
	if err != nil {
		return err
	}

	newServices := make([]string, 0, len(services))
	for _, s := range services {
		if s != name {
			newServices = append(newServices, s)
		}
	}

	if len(newServices) == len(services) {
		return nil // Not found, nothing to do
	}

	return saveEnabledServices(path, newServices)
}

func saveEnabledServices(path string, services []string) error {
	data, err := yaml.Marshal(services)
	if err != nil {
		return fmt.Errorf("failed to marshal enabled services: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write enabled services: %w", err)
	}
	return nil
}
