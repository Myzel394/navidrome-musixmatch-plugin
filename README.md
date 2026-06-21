# Musixmatch Lyrics Plugin for Navidrome

Fetches lyrics from Musixmatch, the official lyrics provider for Spotify. The plugin always tries Musixmatch's free desktop API first, which does not require a user-provided token. If that does not work, the plugin can optionally fall back to website scraping. This plugin supports both synced and plain lyrics.

## Installation

1. Download `navidrome-musixmatch-plugin.ndp` from the [latest release](https://github.com/Myzel394/navidrome-musixmatch-plugin/releases/latest).
2. Copy it to your Navidrome plugins folder (default: `<navidrome-data-directory>/plugins/`).
3. Add `navidrome-musixmatch-plugin` to the lyrics priority list (e.g. using envs: `ND_LYRICSPRIORITY=other-lyric-provider,navidrome-musixmatch-plugin`)
4. In Navidrome, go to **Settings > Plugins > Navidrome Plugin** and toggle it on.

No Musixmatch settings are required for the desktop API path.

## Optional Website Fallback

If the desktop API cannot find lyrics or is blocked, website scraping is supported as a fallback. This fallback requires a Musixmatch website token. If no token is configured, the fallback is skipped.

To enable website fallback, log in to Musixmatch and extract your auth token using a Chromium-based browser:

1. Log in to [Musixmatch](https://www.musixmatch.com/).
2. Open your browser's developer tools (F12) and go to the "Application" tab.
3. Click on Cookies >> https://www.musixmatch.com.
4. Find the cookie named `musixmatchUserToken` and click on it.
5. In the bottom pane, copy the value of the cookie. Make sure "Show URL-decoded" is enabled.
6. Paste that token into "Musixmatch User Token" in the plugin settings.
7. If you encounter captchas, repeat the same steps for the `captcha_id` cookie and paste that into "Musixmatch Captcha ID".


It's recommended to set this plugin's priority to the lowest position, as Musixmatch can block unofficial clients and scrapers.

**Musixmatch is known to aggressively block scrapers, so this plugin may break often. Please always download the latest version to check if your issue is already fixed.**

## Reporting Issues

Before opening an [issue](https://github.com/Myzel394/navidrome-musixmatch-plugin/issues), grep your Navidrome logs and attach the matching lines:

```sh
grep navidrome-musixmatch-plugin /path/to/navidrome.log
```
