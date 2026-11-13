# Musixmatch internals

`FetchLyrics` in `0_main.go` tries these sources in order:

1. **Mobile app gateway** (`1_mobile*.go`) at `apic-appmobile.musixmatch.com/ws/1.1`. It obtains an anonymous `usertoken` with `token.get`, lists matching catalog entries with `track.search`, and fetches lyrics per `track_id` with `macro.subtitles.get`. If no entry can be fetched, a query-based `macro.subtitles.get` call lets the server pick the track.
2. **Website fallback** (`3_search*.go`, `4_website_scraper*.go`, `5_lyrics_scraper.go`), used only when `musixmatch_user_token` is set. It reads the `__NEXT_DATA__` JSON embedded in the search page and the lyrics page.

The desktop client gateway (`2_desktop*.go`, `apic-desktop.musixmatch.com/ws/1.1`) uses the same `token.get` and `macro.subtitles.get` sequence, but is currently disabled in `0_main.go`.

Every source prefers synced lyrics over plain text. Neither gateway is part of Musixmatch's documented developer API, so request and response shapes can change without notice.
