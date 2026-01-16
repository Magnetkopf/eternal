package config

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServiceConfig represents the configuration for a service
type ServiceConfig struct {
	Exec string `yaml:"exec"`
	Dir  string `yaml:"dir"`
}

type SystemConfig struct {
	Token   string `yaml:"token"`
	APIPort int    `yaml:"api_port"`
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

// CreateServiceConfig creates a new service configuration file
func CreateServiceConfig(path string, cfg ServiceConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("service file already exists")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal service config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write service config: %w", err)
	}
	return nil
}

// DeleteServiceConfig removes a service configuration file
func DeleteServiceConfig(path string) error {
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("service file does not exist")
		}
		return fmt.Errorf("failed to remove service config: %w", err)
	}
	return nil
}

func LoadOrGenerateSystemConfig(path string) (SystemConfig, error) {
	// Try to read existing
	data, err := os.ReadFile(path)
	if err == nil {
		// config exists
		var cfg SystemConfig
		if err := yaml.Unmarshal(data, &cfg); err == nil {

			return cfg, nil
		}
	} else if !os.IsNotExist(err) {
		return SystemConfig{}, fmt.Errorf("failed to read config: %w", err)
	}

	// Generate new config
	token := generateRandomString(20)
	cfg := SystemConfig{Token: token, APIPort: 9093}

	data, err = yaml.Marshal(cfg)
	if err != nil {
		return SystemConfig{}, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return SystemConfig{}, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return SystemConfig{}, fmt.Errorf("failed to write config: %w", err)
	}

	return cfg, nil
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			ret[i] = letters[0]
			continue
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret)
}
