#!/usr/bin/env just --justfile

# Run Go tests.
[group('test'), working-directory: 'plugin']
test:
    go test ./...

# Fetch lyrics for a random song via Navidrome SubSonic.
[group('test')]
fetch-plain-lyrics:
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

# Fetch structured synchronized lyrics for a random song via Navidrome SubSonic.
[group('test')]
fetch-synced-lyrics:
    #!/usr/bin/env bash
    set -euo pipefail
    auth='u=admin&p=password&v=1.16.1&c=test&f=json'
    song_id=$(
        curl -fsS "http://localhost:4533/rest/getRandomSongs?size=1&$auth" \
        | jq -r '."subsonic-response".randomSongs.song[0].id'
    )
    lyrics=$(
        curl -fsS "http://localhost:4533/rest/getLyricsBySongId?id=$song_id&enhanced=true&$auth"
    )
    echo "$lyrics"
