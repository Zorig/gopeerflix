#!/bin/bash

APP_NAME="gopeerflix"

echo "Building $APP_NAME for multiple platforms...🚀"

# Linux build
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/${APP_NAME}-linux ./cmd

# macOS build
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/${APP_NAME}-mac ./cmd

# Windows build
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/${APP_NAME}-windows.exe ./cmd

echo "Build completed! Binaries are in ./bin/"
