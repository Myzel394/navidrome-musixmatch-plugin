#!/usr/bin/env just --justfile

GO_REPLACEMENTS := '-X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAuthToken=${OPENOBSERVE_AUTH_TOKEN:-} -X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAttributePluginSource=${OPENOBSERVE_PLUGIN_SOURCE:-github} -X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAttributeVersion=$(git describe --abbrev=0 --tags)'

# Update go module dependencies
[group('plugin'), working-directory: "plugin"]
tidy:
    go mod tidy

# Format and vet all Go source
[group('plugin'), working-directory: "plugin"]
lint-plugin:
    gofumpt -w .

# Run project tests
[group('plugin'), working-directory: "plugin"]
test:
    go test ./...

# Optimized production wasm build (small binary).
[group('plugin'), working-directory: "plugin"]
build:
    tinygo build -ldflags="{{GO_REPLACEMENTS}}" -o plugin.wasm -target wasip1 -buildmode=c-shared .

# Fast dev wasm build using the standard Go compiler.
[group('plugin'), working-directory: "plugin"]
build-dev:
    GOOS=wasip1 GOARCH=wasm \
        go build -buildvcs=false -buildmode=c-shared \
        -ldflags="{{GO_REPLACEMENTS}}" \
        -o plugin.wasm .

# Bundle wasm + manifest into a distributable .ndp archive.
[group('plugin'), working-directory: "plugin"]
pack:
    zip -0 -j {{plugin_name}}.ndp manifest.json plugin.wasm
