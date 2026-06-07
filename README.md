# Musixmatch Lyrics Plugin for Navidrome

Scrapes lyrics from Musixmatch, the official lyrics provider for Spotify. This plugin uses your auth token to access the lyrics and provide them to your Navidrome instance. This plugin supports both synced & plain lyrics.

## Installation

1. Download `navidrome-musixmatch-plugin.ndp` from the [latest release](https://github.com/Myzel394/navidrome-musixmatch-plugin/releases/latest).
2. Copy it to your Navidrome plugins folder (default: `<navidrome-data-directory>/plugins/`).
3. Add `navidrome-musixmatch-plugin` to the lyrics priority list (e.g. using envs: `ND_LYRICSPRIORITY=other-lyric-provider,navidrome-musixmatch-plugin`)
4. In Navidrome, go to **Settings > Plugins > Navidrome Plugin** and toggle it on.
5. Keep this page open for the next steps

Now you need to log in to Musixmatch and extract your auth token (using Chromium-based browsers):

6. Log in to [Musixmatch](https://www.musixmatch.com/)
7. Open your browser's developer tools (F12) and go to the "Application" tab (visible at the top)
8. Click on Cookies >> https://www.musixmatch.com
9. Find the cookie named `musixmatchUserToken` and click on it
10. In the bottom pane, copy the value of the cookie and paste it into the plugin's settings in Navidrome (make sure "Show URL-decoded" is enabled)
11. Paste that token into "Musixmatch User Token"
12. Repeat the same steps for the `captcha_id` cookie, and paste that into the "Musixmatch Captcha ID" field


It's recommended to set this plugin's priority to the lowest position, as scraping is less reliable than using an API.

**Musixmatch is known to aggressively block scrapers, so this plugin may break often. Please always download the latest version to check if your issue is already fixed.**

## Reporting Issues

Before opening an [issue](https://github.com/Myzel394/navidrome-musixmatch-plugin/issues), grep your Navidrome logs and attach the matching lines:

```sh
grep navidrome-musixmatch-plugin /path/to/navidrome.log
```

