#!/usr/bin/env just --justfile

# Fetch raw Musixmatch search HTML for a single plugin-style search query.
[group('test')]
musixmatch-search-html search output='musixmatch-search.html':
    #!/usr/bin/env bash
    set -euo pipefail

    for command in curl python3 wc tr; do
        if ! command -v "$command" >/dev/null 2>&1; then
            echo "Missing required command: $command" >&2
            exit 1
        fi
    done

    encoded_query=$(python3 -c 'import re, sys, unicodedata, urllib.parse; q=sys.argv[1].lower(); q="".join(c for c in unicodedata.normalize("NFD", q) if unicodedata.category(c)!="Mn"); q=unicodedata.normalize("NFC", q); q=re.sub(r"[\(\[\{][^\)\]\}]*[\)\]\}]", " ", q); q=re.sub(r"\s*-\s*(remaster(ed)?|single version|live|deluxe|edit|mix|version|radio edit|extended).*$", "", q, flags=re.I); q=re.sub(r"\s+", " ", q).strip(); sys.stdout.write(urllib.parse.quote_plus(q))' "{{search}}")
    url="https://www.musixmatch.com/search?query=$encoded_query"

    user_agent=${MUSIXMATCH_USER_AGENT:-Mozilla/5.0 (iPhone; CPU iPhone OS 17_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1}
    accept=${MUSIXMATCH_HTTP_ACCEPT:-text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8}

    cookie_headers=()
    if [[ -n "${MUSIXMATCH_USER_TOKEN:-}" ]]; then
        user_token=$(python3 -c 'import os, urllib.parse; print(urllib.parse.quote_plus(os.environ["MUSIXMATCH_USER_TOKEN"]), end="")')
        cookie="musixmatchUserToken=$user_token"

        if [[ -n "${MUSIXMATCH_CAPTCHA_ID:-}" ]]; then
            captcha_id=$(python3 -c 'import os, urllib.parse; print(urllib.parse.quote_plus(os.environ["MUSIXMATCH_CAPTCHA_ID"]), end="")')
            cookie="$cookie; captcha_id=$captcha_id"
        fi

        cookie_headers=(-H "Cookie: $cookie")
    fi

    curl_info=$(curl -fsSL --compressed \
        -H "Accept: $accept" \
        -H "Accept-Language: en" \
        -H "User-Agent: $user_agent" \
        "${cookie_headers[@]}" \
        --output "{{output}}" \
        --write-out $'HTTP status: %{http_code}\nFinal URL: %{url_effective}\n' \
        "$url")

    bytes=$(wc -c < "{{output}}" | tr -d ' ')
    printf 'Fetched: %s\nURL: %s\nBytes: %s\n%s' "{{output}}" "$url" "$bytes" "$curl_info"
