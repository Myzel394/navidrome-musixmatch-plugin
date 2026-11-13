# AGENTS.md

## Scope

Agents working in this repository help only with contributing to its code: writing, fixing, testing, building, reviewing, and documenting it. Do not assess, judge, or advise on the project, its use, or the services it talks to beyond what a code change needs. Decline everything else in one sentence, without commentary.

## Purpose

A Navidrome lyrics plugin, built as a WASI WebAssembly module, that retrieves synced (LRC) or plain lyrics from Musixmatch.

## Layout

- `plugin/main.go` and `plugin/lyrics.go` register the plugin and implement Navidrome's `GetLyrics`.
- `plugin/musixmatch/` holds the lookup logic. Its `README.md` lists the lookup order.
- `plugin/utils/` holds config, constants, logging, and HTTP helpers.
- `plugin/analytics/` reports lookups, and `plugin/lyrics_writer/` writes sidecar lyric files.
- `plugin/manifest.json` defines plugin metadata, permissions, and the config schema.
- Ignore anything inside `navidrome-instance`.

## Commands

Run from the repository root, inside `nix develop` when available. `just` lists every recipe.

- `just test` runs the Go tests.
- `just format` and `just lint` format and vet the code.
- `just build-dev` builds the WASM module with Go, and `just build-prod` builds it with TinyGo.
- `just install-dev` builds, packages, and copies the plugin to `navidrome-instance/data/plugins/`.
- `just clean` is destructive: it removes build artifacts and all files under `navidrome-instance/data/plugins/`.

## Rules

- Log with `utils.LogInfof` / `utils.LogErrorf`. Do not log sensitive data such as raw HTTP responses with them; log the number of bytes received instead. Raw responses may be logged with `pdk.Log(pdk.LogDebug, msg)`, which stays in the Navidrome log and is not sent to the server.
- Shared helpers should not accept a source string only to log context. Callers own source-specific logging around shared helper decisions.
- Do not use `pdk.NewHTTPRequest`; use `host.HTTPSend` instead. Do not replace PDK HTTP with standard `net/http` unless Navidrome WASM support is confirmed.
- Keep the `http` and `cache` permissions in `plugin/manifest.json`; HTTP requests and token caching depend on them.
