package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Magnetkopf/Eternal/internal/config"
	"github.com/Magnetkopf/Eternal/internal/process"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ProcessData struct {
	Name   string `json:"name"`
	PID    int    `json:"pid,omitempty"`
	Status string `json:"status"`
}

type ServiceListEntry struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Enabled bool   `json:"enabled"`
}

type CreateServiceRequest struct {
	Exec string `json:"exec"`
	Dir  string `json:"dir"`
}

func StartServer(pm *process.Manager, port int, servicesDir, enabledFile, authToken string) {
	mux := http.NewServeMux()

	// wrapper to inject dependencies
	h := &handler{
		pm:          pm,
		servicesDir: servicesDir,
		enabledFile: enabledFile,
	}

	// GET /v1/processes
	mux.HandleFunc("GET /v1/processes", h.handleList)

	// GET /v1/processes/:name
	mux.HandleFunc("GET /v1/processes/{name}", h.handleGet)

	// DELETE /v1/processes/:name
	mux.HandleFunc("DELETE /v1/processes/{name}", h.handleDelete)

	// PUT /v1/processes/:name
	mux.HandleFunc("PUT /v1/processes/{name}", h.handleCreate)

	// POST /v1/processes/:name/:action
	mux.HandleFunc("POST /v1/processes/{name}/{action}", h.handleAction)

	// Auth Middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("access-token")
			if token != authToken {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Printf("API Server listening on %s\n", addr)

	if err := http.ListenAndServe(addr, authMiddleware(mux)); err != nil {
		fmt.Printf("API Server failed: %v\n", err)
	}
}

type handler struct {
	pm          *process.Manager
	servicesDir string
	enabledFile string
}

func (h *handler) respondJSON(w http.ResponseWriter, code int, message string, data interface{}, errStr string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := Response{
		Code: code,
	}
	if errStr != "" {
		resp.Error = errStr
	} else {
		resp.Message = message
		resp.Data = data
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *handler) respondError(w http.ResponseWriter, code int, errStr string) {
	h.respondJSON(w, code, "", nil, errStr)
}

func (h *handler) respondSuccess(w http.ResponseWriter, message string, data interface{}) {
	h.respondJSON(w, 200, message, data, "")
}

func (h *handler) handleList(w http.ResponseWriter, r *http.Request) {
	// Get runtime status
	statuses := h.pm.ListServices()

	// Get enabled status
	enabledList, err := config.LoadEnabledServices(h.enabledFile)
	if err != nil {
		h.respondError(w, 500, fmt.Sprintf("failed to load enabled services: %v", err))
		return
	}

	enabledMap := make(map[string]bool)
	for _, name := range enabledList {
		enabledMap[name] = true
	}

	allNames := make(map[string]struct{})
	for k := range statuses {
		allNames[k] = struct{}{}
	}
	for k := range enabledMap {
		allNames[k] = struct{}{} // Though 'enabled' is just names
	}

	var list []ServiceListEntry
	for name := range allNames {
		st, ok := statuses[name]
		statusStr := "unknown"
		if ok {
			statusStr = string(st)
		} else {
			statusStr = "not-found"
		}

		list = append(list, ServiceListEntry{
			Name:    name,
			Status:  statusStr,
			Enabled: enabledMap[name],
		})
	}

	h.respondSuccess(w, "success", list)
}

func (h *handler) handleGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	status, err := h.pm.GetStatus(name)
	if err != nil {
		h.respondError(w, 404, fmt.Sprintf("service '%s' not found", name))
		return
	}

	data := ProcessData{
		Name:   name,
		Status: string(status),
	}
	h.respondSuccess(w, "success", data)
}

func (h *handler) handleAction(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	action := r.PathValue("action")

	var err error
	var msg string

	switch action {
	case "start":
		err = h.pm.StartService(name)
		msg = "process started successfully"
	case "stop":
		err = h.pm.StopService(name)
		msg = "process stopped successfully"
	case "restart":
		err = h.pm.RestartService(name)
		msg = "process restarted successfully"
	case "enable":
		err = config.EnableService(h.enabledFile, name)
		msg = "service enabled"
	case "disable":
		err = config.DisableService(h.enabledFile, name)
		msg = "service disabled"
	default:
		h.respondError(w, 400, fmt.Sprintf("unknown action: %s", action))
		return
	}

	if err != nil {
		h.respondError(w, 500, err.Error())
		return
	}

	// For start/stop/restart, returning status is good.
	status, _ := h.pm.GetStatus(name)

	data := ProcessData{
		Name:   name,
		Status: string(status),
	}

	h.respondSuccess(w, msg, data)
}

func (h *handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var req CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, 400, "invalid json body")
		return
	}

	if req.Exec == "" {
		h.respondError(w, 400, "exec is required")
		return
	}

	cfg := config.ServiceConfig{
		Exec: req.Exec,
		Dir:  req.Dir,
	}

	serviceFile := filepath.Join(h.servicesDir, name+".yaml")
	if err := config.CreateServiceConfig(serviceFile, cfg); err != nil {
		h.respondError(w, 500, err.Error())
		return
	}

	h.pm.LoadServices()

	h.respondSuccess(w, "service created", nil)
}

func (h *handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	// Stop if running
	status, _ := h.pm.GetStatus(name)
	if status == process.StatusRunning {
		h.pm.StopService(name)
	}

	// Remove from manager memory
	h.pm.RemoveService(name)

	// Disable
	config.DisableService(h.enabledFile, name)

	// Delete file
	serviceFile := filepath.Join(h.servicesDir, name+".yaml")
	if err := config.DeleteServiceConfig(serviceFile); err != nil {
		h.respondError(w, 500, err.Error())
		return
	}

	h.respondSuccess(w, "service deleted", nil)
}
