# Eternal

Eternal is a simple process manager for Linux.

## Building

```bash
bash build.sh

mv eternal eternal-daemon ~/.local/bin/
# or
sudo mv eternal eternal-daemon /usr/local/bin/
```

## Usage

Eternal supports CLI and API access.

### CLI

```bash
cat <<EOF > ~/.eternal/services/example.yaml
exec: /bin/sleep 100
dir: /tmp
EOF

# create service
eternal new example
# delete service
eternal delete example

# enable auto start
eternal enable example
# start now
eternal start example
# disable auto start
eternal disable example
# stop now
eternal stop example
```

### API

Check [api.md](docs/api.md) for API usage and testing.

## Configuration

Services are stored in `~/.eternal/services/`, add a YAML file for each service.

- `exec`: The command to run
- `dir`: The directory to run the command in

Example:

```yaml
exec: /bin/sleep 100
dir: /tmp
```
