# Musixmatch internals

## Mobile API hint

The mobile fetcher mimics Musixmatch's iOS app at
`apic-appmobile.musixmatch.com/ws/1.1` (`app_id=mac-ios-v2.0`). All Musixmatch
gateways — the official `api.musixmatch.com` (paid `apikey`), the desktop
gateway below, and this mobile one — expose the same `ws/1.1` methods and
`message.header`/`message.body` envelope; the unofficial gateways just
authenticate with an anonymous `usertoken` from `token.get` instead.

`track.search` (`1_mobile__search.go`) is Musixmatch's documented catalog
search (`q_track`, `q_artist`, `f_has_lyrics`, `page`, `page_size`). It returns
track metadata only, never lyrics. The lookup runs it first to enumerate
Musixmatch's duplicate entries per song, orders them by how close their length
is to the requested track, then fetches them by `track_id` with
`macro.subtitles.get`, stopping at the first synced lyrics and only using plain
lyrics if no entry has synced ones — the same search-then-fetch pattern used
by lyrics clients such as syncedlyrics. Only when the search yields no usable
track ID, or every track lookup fails, does the plugin fall back to a
query-based `macro.subtitles.get` call that lets the server-side matcher pick
the track.

## Desktop API hint

The desktop fetcher mimics Musixmatch's
unofficial desktop client API at `apic-desktop.musixmatch.com/ws/1.1`:

1. Fetch an anonymous token with `token.get`, passing
   `app_id=web-desktop-app-v1.0` and `user_language=en`. The plugin caches this
   `user_token` for 10 minutes.
2. Fetch lyrics with `macro.subtitles.get`, passing `format=json`,
   `namespace=lyrics_richsynced`, `optional_calls=track.richsync`,
   `subtitle_format=lrc`, `q_artist`, `q_track`, and the cached `usertoken`.
   When Navidrome provides a duration, the plugin also sends a rounded
   `f_subtitle_length` plus `f_subtitle_length_max_deviation=3` to narrow the
   match.
3. Parse the macro response's `macro_calls` in quality order:
   `track.richsync.get` first, converting each richsync `{ts, x}` line to LRC;
   then `track.subtitles.get`, which can already contain LRC text; then
   `track.lyrics.get` as plain-text fallback.
4. If the desktop API returns `401`, the cached token is removed so a later
   lookup can fetch a fresh one. If this path fails or has no lyrics, the caller
   can fall back to Musixmatch website scraping.

This mirrors a known public GitHub recipe used by lyrics clients such as
Spicetify, Cider, and Lyricify Lyrics Helper: bootstrap `token.get`, call
`macro.subtitles.get` with `web-desktop-app-v1.0`, request richsync/LRC data,
and prefer synced lyrics over plain text. It is an unofficial endpoint, so the
shape and blocking behavior can change without warning.
