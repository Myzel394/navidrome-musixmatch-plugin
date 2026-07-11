#!/usr/bin/env just --justfile

# Format Go and repository files.
[group('dev')]
format:
    gofumpt -w plugin
    treefmt .

# Vet Go packages.
[group('dev'), working-directory: 'plugin']
lint:
    go vet ./...

# Update Go module dependencies.
[group('dev'), working-directory: 'plugin']
tidy:
    go mod tidy

# Update pinned GitHub Action versions.
[group('dev')]
update-pinact:
    pinact run

# Build and package the PR dev artifact.
[group('dev')]
create-dev-release:
    @just build-dev
    @just package

# Build & install dev version of the plugin to Navidrome
[group('install')]
install-dev:
    @just create-dev-release
    mkdir -p {{plugins_dir}}
    cp plugin/{{plugin_name}}.ndp {{plugins_dir}}/

# Destructive: remove build artifacts and all files under navidrome-instance/data/plugins/.
[group('install')]
clean:
    rm -f plugin/plugin.wasm plugin/{{plugin_name}}.ndp *.zip
    rm -f {{plugins_dir}}/*
