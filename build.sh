#!/bin/sh

set -e

GOOS=linux GOARCH=amd64 go build -o eternal ./cmd/eternal
GOOS=linux GOARCH=amd64 go build -o eternal-daemon ./cmd/eternal-daemon