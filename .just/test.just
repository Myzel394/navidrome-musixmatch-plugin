#!/usr/bin/env just --justfile

# Fetch lyrics for a random song via the Navidrome SubSonic API
[group('test')]
fetch-lyrics:
    #!/usr/bin/env bash
    set -euo pipefail
    auth='u={{username}}&p={{password}}&v=1.16.0&c=test&f=json'
    song_id=$(
        curl -s "http://localhost:4533/rest/getRandomSongs?$auth" \
        | jq -r '.["subsonic-response"].randomSongs.song[0].id'
    )
    lyrics=$(
        curl -s "http://localhost:4533/rest/getLyricsBySongId?id=$song_id&$auth" \
        
    )
    echo "$lyrics"

# Remove all build artifacts: wasm, ndp, zips, and the installed plugin.
[group('test')]
clear:
    rm -f plugin/plugin.wasm plugin/{{plugin_name}}.ndp *.zip
    rm -f {{plugins_dir}}/plugin.wasm {{plugins_dir}}/manifest.json {{plugins_dir}}/*
