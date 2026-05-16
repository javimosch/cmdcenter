#!/bin/bash
# Build cmdcenter - Default
go build -o cmdcenter-default main.go server.go daemon.go config.go
ls -lh cmdcenter-default

# Build cmdcenter - Optimized
go build -ldflags "-s -w" -o cmdcenter-optimized main.go server.go daemon.go config.go
ls -lh cmdcenter-optimized