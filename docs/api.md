# Eternal Docs

## API

### RESTful API

#### Base URL
`http://127.0.0.1:9093`

#### Authentication
The `access-token` header is required for all requests. You can find your token in `~/.eternal/config.yaml`.

#### Standard Response Format
All responses follow this JSON structure:
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

#### Endpoints

##### 1. List Processes
**GET** `/v1/processes`

Returns a list of all services, their current running status, and whether they are enabled to start on boot.

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": [
    {
      "name": "test-service",
      "status": "running",
      "enabled": true
    }
  ]
}
```

##### 2. Get Process Status
**GET** `/v1/processes/:name`

Returns the status of a specific service.

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "name": "test-service",
    "status": "running"
  }
}
```

##### 3. Create Service
**PUT** `/v1/processes/:name`

Creates a new service configuration.

**Request Body:**
```json
{
  "exec": "sleep 100",  // Required: Command to execute
  "dir": "/tmp"         // Optional: Working directory
}
```

**Response:**
```json
{
  "code": 200,
  "message": "service created"
}
```

##### 4. Start Service
**POST** `/v1/processes/:name/start`

Starts a stopped service.

**Response:**
```json
{
  "code": 200,
  "message": "process started successfully",
  "data": {
      "name": "test_service",
      "status": "running"
  }
}
```

##### 5. Stop Service
**POST** `/v1/processes/:name/stop`

Stops a running service.

**Response:**
```json
{
  "code": 200,
  "message": "process stopped successfully",
  "data": {
      "name": "test_service",
      "status": "stopped" 
  }
}
```

##### 6. Restart Service
**POST** `/v1/processes/:name/restart`

Restarts a service.

**Response:**
```json
{
  "code": 200,
  "message": "process restarted successfully",
  "data": {
      "name": "test_service",
      "status": "running" 
  }
}
```

##### 7. Enable Service
**POST** `/v1/processes/:name/enable`

Enables a service to start automatically on daemon startup.

**Response:**
```json
{
  "code": 200,
  "message": "service enabled",
  "data": {
      "name": "test_service",
      "status": "stopped" 
  }
}
```

##### 8. Disable Service
**POST** `/v1/processes/:name/disable`

Disables a service from starting automatically.

**Response:**
```json
{
  "code": 200,
  "message": "service disabled",
  "data": {
      "name": "test_service",
      "status": "stopped" 
  }
}
```

##### 9. Delete Service
**DELETE** `/v1/processes/:name`

Stops the service (if running), disables it, and deletes the configuration file.

**Response:**
```json
{
  "code": 200,
  "message": "service deleted"
}
```
