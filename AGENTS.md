# AGENTS.md

## Purpose

This repository contains a Navidrome lyrics plugin that scrapes lyrics from Musixmatch. Navidrome loads the plugin as a WASI WebAssembly module and calls its lyrics extension point. The plugin searches Musixmatch for a requested artist/title, fetches the matched lyrics page, parses embedded Next.js data, and returns synced LRC lyrics when available or plain lyrics as a fallback.

Treat this file as the first source of project context. Search the codebase only when this file is incomplete, stale, or contradicted by the current task.

## Repository Layout

- `plugin/` is the Go module and plugin source root.
- `plugin/main.go` registers the Navidrome lyrics plugin and exports `nd_on_init`.
- `plugin/lyrics.go` implements Navidrome's `GetLyrics` method and delegates to `plugin/musixmatch`.
- `plugin/musixmatch/` contains Musixmatch-specific desktop API, website search/fetch fallback, normalization, and LRC conversion logic.
- `plugin/utils/` contains shared config, logging, constants, and PDK HTTP helpers.
- `plugin/manifest.json` defines Navidrome plugin metadata, HTTP permissions for `apic-desktop.musixmatch.com` and `www.musixmatch.com`, cache permission for the desktop API token, and plugin config schema.
- `just/` and `justfile` define local build, test, install, and cleanup commands.
- `flake.nix` defines the Nix dev shell and reproducible TinyGo/Nix package build.

## Runtime Flow

1. `plugin/main.go` calls `lyrics.Register(&plugin{})` in `init`.
2. Navidrome calls `(*plugin).GetLyrics` in `plugin/lyrics.go`.
3. `GetLyrics` calls `musixmatch.FetchLyrics` and prefixes returned errors with `navidrome-musixmatch-plugin: `.
4. `musixmatch.FetchLyrics` logs the requested track and first tries the free, unofficial Musixmatch desktop API.
5. `fetchLyricsFromDesktopAPI` gets an anonymous desktop `user_token` via `token.get`, caches it with Navidrome's host cache for 10 minutes, then calls `macro.subtitles.get`.
6. Desktop API lyrics are returned in this order: richsync converted to LRC, subtitle LRC, then plain lyrics.
7. If the desktop API fails or returns no lyrics, `FetchLyrics` falls back to website scraping.
8. `searchForTrack` normalizes artist/title, fetches the Musixmatch website search page, extracts `<script id="__NEXT_DATA__">`, unmarshals embedded search JSON, then chooses `bestMatch` when it is a `track`, otherwise `tracks[0]`.
9. `scrapeWebsiteLyricsForTrack` fetches the Musixmatch lyrics page for the selected `commontrack_vanity_id`, extracts `<script id="__NEXT_DATA__">`, and unmarshals the JSON payload.
10. Website fallback lyrics are returned in this order: synced LRC from `trackStructureList`, synced LRC from `subtitle`, then plain `lyrics.body`.

## Key Files

- `plugin/musixmatch/1_desktop.go` implements the unofficial desktop API path at `apic-desktop.musixmatch.com/ws/1.1`, including 10-minute token caching and `macro.subtitles.get` parsing.
- `plugin/musixmatch/2_search.go` uses `MusixmatchSearchPageURL`, extracts the search page's `__NEXT_DATA__`, and parses `pageProps.data.openSearch.data.opensearchTrackSearch.body`.
- `plugin/musixmatch/3_website_scraper.go` uses `nextDataRe` to parse the Musixmatch page and reads `props.pageProps.data.trackInfo.data`.
- `plugin/musixmatch/4_website_scraper__lyrics_parser.go` converts Musixmatch timestamp totals into LRC tags like `[mm:ss.hh]`.
- `plugin/musixmatch/9_utils.go` normalizes input by lowercasing, stripping diacritics, removing bracketed text, removing common dash suffixes, and collapsing whitespace.
- `plugin/utils/http.go` sends GET requests with PDK HTTP, `Accept-Language: en`, configured `Accept`, configured `User-Agent`, and optional Musixmatch cookies.
- `plugin/utils/constants.go` contains the hardcoded Musixmatch URLs and default headers.

## Configuration

Config keys come from `plugin/manifest.json` and `plugin/utils/constants.go`.

- `musixmatch_user_token` is sent as the `musixmatchUserToken` cookie when set.
- `musixmatch_captcha_id` is sent as the `captcha_id` cookie only when `musixmatch_user_token` is also set.
- `musixmatch_user_agent` defaults to mobile Safari.
- `musixmatch_http_accept` defaults to a browser-like HTML accept header.

The desktop API path does not require user-entered Musixmatch cookies or an official API key. It automatically fetches an anonymous desktop `user_token` and caches it with Navidrome's `host.Cache*` service for 10 minutes. The manifest must keep the `cache` permission so Navidrome exports the required cache host functions.

Note: `musixmatch_captcha_id` exists in the manifest schema, but it is not currently included in `uiSchema.elements`, so verify the Navidrome plugin settings UI before assuming users can edit it there.

## Build And Validation

Run commands from the repository root unless a command states otherwise.

- Enter the Nix dev shell when available: `nix develop`.
- Run tests: `go test ./...` from `plugin/`, or `just test` from the repo root. The Musixmatch desktop API end-to-end test makes real network requests.
- Format Go and repository files: `just format`.
- Vet Go packages: `just lint`.
- Dev WASM build: `just build-dev`.
- Production WASM build: `just build-prod`.
- Package plugin archive: `just package`, which creates `plugin/navidrome-musixmatch-plugin.ndp` from `plugin/manifest.json` and `plugin/plugin.wasm`.
- Dev release package: `just create-dev-release` builds dev WASM and packages it.
- Local dev install: `just install-dev` builds, packages, and copies the `.ndp` to `navidrome-instance/data/plugins/`.
- Production release package: `just create-prod-release` builds the Nix package.
- Fetch lyrics from a local Navidrome test instance: `just fetch-lyrics`.
- Destructive cleanup: `just clean` removes build artifacts and all files under `navidrome-instance/data/plugins/`.
- Reproducible package build: `nix build .#default`.

Validated during repository onboarding: `go test ./...` passes from `plugin/`, but there are currently no test files.

## Fragility And Gotchas

- This uses an unofficial desktop API first and website scraping as fallback. Musixmatch can block requests, require captcha cookies on the website path, or change API/page/search JSON shapes without warning.
- `MusixmatchDesktopAPIURL` uses the unofficial desktop endpoint with `app_id=web-desktop-app-v1.0`; if free lookups start failing, check this endpoint and token flow first.
- Website search parses embedded `__NEXT_DATA__` from `/search?query=...`; it requires a valid `musixmatchUserToken` cookie value and may redirect to auth without one.
- Search result selection is permissive and does not verify returned artist/title equality after normalization.
- `slugify` in `plugin/musixmatch/9_utils.go` appears unused.
- The plugin returns an empty successful lyrics response when search has no results, but returns errors for fetch/parse failures.
- HTTP requests require Navidrome plugin HTTP permission from `manifest.json`; desktop token caching requires cache permission. Do not replace PDK HTTP with standard `net/http` unless Navidrome WASM support is confirmed.

## Dependency Notes

- Main runtime dependency is `github.com/navidrome/navidrome/plugins/pdk/go`.
- `golang.org/x/text` is used for Unicode normalization and diacritic stripping.
- TinyGo is required for WASI plugin builds.
- Nix dev shell includes Go, TinyGo, gopls, gofumpt, zip, just, binaryen, nixfmt-tree, and pinact.

## Best-Practice Source

This file follows the `AGENTS.md` convention: a root Markdown file that gives coding agents project overview, build/test commands, code layout, and known gotchas. The convention is documented at `https://agents.md/` and is also recognized by GitHub Copilot repository custom-instructions documentation as an agent-instructions file.
