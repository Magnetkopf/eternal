# Eternal Docs

## Configuration

Eternal uses YAML files for configuration. All configuration files are stored in the `~/.eternal/` directory.

## Directory Structure

```text
~/.eternal/
├── config.yaml          # System-wide configuration (daemon settings)
├── enabled.yaml         # List of services that should start on boot
└── services/            # Directory containing individual service configurations
    ├── web-server.yaml
    ├── worker.yaml
    └── ...
```

## System Configuration

The system configuration is stored in `~/.eternal/config.yaml`. This file controls the behavior of the `eternal-daemon`.

If this file does not exist, it will be automatically generated with a random authentication token and default port when the daemon starts.

### Fields

| Field      | Type   | Description                                                      | Default |
|------------|--------|------------------------------------------------------------------|---------|
| `token`    | string | Authentication token for API access. Auto-generated on first run.| (Random)|
| `api_port` | int    | The TCP port where the Eternal API server listens.               | `9093`  |

### Example `config.yaml`

```yaml
token: "a1b2c3d4e5f6g7h8i9j0"
api_port: 9093
```

## Service Configuration

Each process managed by Eternal is defined by a service configuration file located in `~/.eternal/services/<service-name>.yaml`. The filename (without extension) determines the service name.

### Fields

| Field  | Type   | Required | Description |
|--------|--------|----------|-------------|
| `exec` | string | **Yes**  | The command string to execute. Arguments should be space-separated. |
| `dir`  | string | No       | The working directory for the process. If omitted, it defaults to the directory where the daemon was started (or system default). |

### Example `my-service.yaml`

```yaml
# ~/.eternal/services/my-service.yaml
exec: "/usr/bin/python3 app.py --port 8080"
dir: "/home/user/projects/my-app"
```

## Enabled Services

The `~/.eternal/enabled.yaml` file maintains a list of services that are automatically started when the `eternal-daemon` launches.

This file is managed automatically by the `eternal enable <name>` and `eternal disable <name>` commands. You typically do not need to edit this file manually.

### Example `enabled.yaml`

```yaml
- web-server
- database-worker
```
