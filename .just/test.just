#!/usr/bin/env just --justfile

# Fetch lyrics for a random song via the Navidrome SubSonic API
[group('test')]
fetch-lyrics:
    #!/usr/bin/env bash
    set -euo pipefail
    auth='u=admin&p=password&v=1.16.0&c=test&f=json'
    song=$(
        curl -fsS "http://localhost:4533/rest/getRandomSongs?size=1&$auth" \
        | jq -r '.["subsonic-response"].randomSongs.song[0]'
    )
    artist=$(jq -nr --arg value "$(jq -r '.artist' <<<"$song")" '$value|@uri')
    title=$(jq -nr --arg value "$(jq -r '.title' <<<"$song")" '$value|@uri')
    lyrics=$(
        curl -fsS "http://localhost:4533/rest/getLyrics?artist=$artist&title=$title&$auth"
    )
    echo "$lyrics"

# Remove all build artifacts: wasm, ndp, zips, and the installed plugin.
[group('test')]
clear:
    rm -f plugin/plugin.wasm plugin/{{plugin_name}}.ndp *.zip
    rm -f {{plugins_dir}}/plugin.wasm {{plugins_dir}}/manifest.json {{plugins_dir}}/*
