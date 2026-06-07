# ── Shared variables ───────────────────────────────
plugin_name := "navidrome-shazam-plugin"
data_dir    := "navidrome-instance/data"
plugins_dir := data_dir / "plugins"

username := "admin"
password := "password"

# ── Imports ────────────────────────────────────────
import '.just/plugin.just'
import '.just/dev.just'
import '.just/prod.just'
import '.just/test.just'
import '.just/cicd.just'

# ── Default ────────────────────────────────────────
default:
    @just --list

lint:
    @just lint-plugin
    treefmt .
