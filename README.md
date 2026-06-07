# Shazam Lyrics Plugin for Navidrome

Scrapes lyrics from Shazam. No API key required.

Since this plugin scrapes Shazam's website, this plugin may break often, please always download the latest version to check if your issue is already fixed.

## Installation

1. Download `navidrome-shazam-plugin.ndp` from the [latest release](https://github.com/Myzel394/navidrome-shazam-plugin/releases/latest).
2. Copy it to your Navidrome plugins folder (default: `<navidrome-data-directory>/plugins/`).
3. Add `navidrome-shazam-plugin` to the lyrics priority list (e.g. using envs: `ND_LYRICSPRIORITY=other-lyric-provider,navidrome-shazam-plugin`)
4. In Navidrome, go to **Settings > Plugins > Shazam Plugin** and toggle it on.

It's recommended to set this plugin's priority to the lowest position, as scraping is less reliable than using an API.

## Configuration

All settings are optional. The most relevant one — **Search Country** — defaults to `US`; set it to your country's ISO code for better results. Other settings (User-Agent, language, match threshold, etc.) live under **Advanced**.

## Reporting Issues

Before opening an [issue](https://github.com/Myzel394/navidrome-shazam-plugin/issues), grep your Navidrome logs and attach the matching lines:

```sh
grep navidrome-shazam-plugin /path/to/navidrome.log
```

Include the track (artist + title) that failed and your plugin version.

