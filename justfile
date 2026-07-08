# ── Shared variables ───────────────────────────────
plugin_name := "navidrome-musixmatch-plugin"
data_dir    := "navidrome-instance/data"
plugins_dir := data_dir / "plugins"

username := "admin"
password := "password"

set dotenv-load := true

# ── Imports ────────────────────────────────────────
import '.just/plugin.just'
import '.just/dev.just'
import '.just/prod.just'
import '.just/test.just'
import '.just/cicd.just'
import '.just/release.just'

# ── Default ────────────────────────────────────────
default:
    @just --list

lint:
    @just lint-plugin
    treefmt .
