#!/usr/bin/env just --justfile

# Install the packed .ndp plugin into the local Navidrome instance.
[group('prod')]
install: 
    @just build
    @just pack
    mkdir -p {{plugins_dir}}
    cp plugin/{{plugin_name}}.ndp {{plugins_dir}}/

# Build & Install a full & ready prod build
[group('prod')]
full-prod-install:
    nix build .#default
    cp result/bin/*.ndp {{plugins_dir}}/
