package ipc

// RequestType defines the type of action requested
type RequestType string

const (
	RequestStart   RequestType = "start"
	RequestStop    RequestType = "stop"
	RequestRestart RequestType = "restart" // Optional, but good to define
	RequestStatus  RequestType = "status"
)

// Request defines the structure of a command sent to the daemon
type Request struct {
	Type    RequestType `json:"type"`
	Service string      `json:"service"`
}

// Response defines the structure of the reply from the daemon
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
