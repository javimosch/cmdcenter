#!/bin/bash
# Build Go CLI with UI - Default
go build -o boilerplate-cli-ui-go-default main.go server.go daemon.go
ls -lh boilerplate-cli-ui-go-default

# Build Go CLI with UI - Optimized
go build -ldflags "-s -w" -o boilerplate-cli-ui-go-optimized main.go server.go daemon.go
ls -lh boilerplate-cli-ui-go-optimized