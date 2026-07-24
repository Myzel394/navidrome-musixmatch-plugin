# ── Shared variables ───────────────────────────────
plugin_name := "navidrome-musixmatch-plugin"
data_dir    := "navidrome-instance/data"
plugins_dir := data_dir / "plugins"

set dotenv-load := true

# ── Imports ────────────────────────────────────────
import 'just/build.just'
import 'just/dev.just'
import 'just/test.just'
import 'just/release.just'
import 'just/musixmatch-test.just'

# ── Default ────────────────────────────────────────
default:
    @just --list
