#!/usr/bin/env just --justfile

GO_REPLACEMENTS := '-X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAuthToken=${OPENOBSERVE_AUTH_TOKEN:-} -X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAttributePluginSource=${OPENOBSERVE_PLUGIN_SOURCE:-github} -X github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils.OpenObserveAttributeVersion=$(git describe --abbrev=0 --tags)'

# Fast dev wasm build using the standard Go compiler.
[group('build'), working-directory: 'plugin']
build-dev:
    GOOS=wasip1 GOARCH=wasm \
        go build -buildvcs=false -buildmode=c-shared \
        -ldflags="{{GO_REPLACEMENTS}}" \
        -o plugin.wasm .

# Optimized production wasm build.
[group('build'), working-directory: 'plugin']
build-prod:
    tinygo build -ldflags="{{GO_REPLACEMENTS}}" -o plugin.wasm -target wasip1 -buildmode=c-shared .

# Bundle, for optimization use `advzip`
[group('build'), working-directory: 'plugin']
package:
    zip -0 -j {{plugin_name}}.ndp manifest.json plugin.wasm
