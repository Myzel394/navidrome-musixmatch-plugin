# Musixmatch Lyrics Plugin for Navidrome

Fetches lyrics from Musixmatch, the official lyrics provider for Spotify. The plugin first queries Musixmatch's app API, which does not require a user-provided token. If that does not return lyrics, the plugin can optionally fall back to the Musixmatch website. This plugin supports both synced and plain lyrics.

This is an independent project. It is not affiliated with, endorsed by, or sponsored by Musixmatch.

## Installation

1. Download `navidrome-musixmatch-plugin.ndp` from the [latest release](https://github.com/Myzel394/navidrome-musixmatch-plugin/releases/latest).
2. Copy it to your Navidrome plugins folder (default: `<navidrome-data-directory>/plugins/`).
3. Add `navidrome-musixmatch-plugin` to the lyrics priority list (e.g. using envs: `ND_LYRICSPRIORITY=other-lyric-provider,navidrome-musixmatch-plugin`)
4. In Navidrome, go to **Settings > Plugins > Navidrome Plugin** and toggle it on.

No Musixmatch settings are required for the app API path.

### Which release should I use?

You can download this plugin from [GitHub](https://github.com/Myzel394/navidrome-musixmatch-plugin) or from the [Forgejo mirror](https://git.nerdvpn.de/Myzel394/navidrome-musixmatch-plugin).

Most development happens on GitHub, but both repositories are kept available so you can use whichever platform you prefer and report issues there.
GitHub releases are built with GitHub Actions. Forgejo releases are built by my own self-hosted Forgejo runner.

If you are unsure which one to choose, I recommend the [Forgejo releases](https://git.nerdvpn.de/Myzel394/navidrome-musixmatch-plugin/releases), as they are built on my own runner and are therefore fully under my control.

## Automatically save fetched lyrics

`navidrome-musixmatch-plugin` can automatically save fetched lyrics to your library. It will write lyric files next to the audio files using the extension `.lrc` for synced lyrics and `.txt` for plain lyrics. Example: `Billie Eilish - Happier Than Ever.flac` will have its lyrics saved as `Billie Eilish - Happier Than Ever.lrc` or `Billie Eilish - Happier Than Ever.txt`.
Lyrics writing is off by default and can be enabled in the Navidrome plugin settings.

For this to work, make sure your Music library is writable by Navidrome (if you use Docker for example, mount it _without_ `ro` at the end of the volume mount), and in the Navidrome plugin settings under "Library Permissions", enable "Allow write access".

**IMPORTANT**: `navidrome-musixmatch-plugin` will NOT check for existing lyric files, it will always fetch lyrics from Musixmatch and overwrite any existing lyric files. If you want to prioritize existing lyric files, [make sure your lyrics priority list has extensions set _before_ `navidrome-musixmatch-plugin`](https://www.navidrome.org/docs/usage/configuration/options/#opt-lyricspriority), like this:

```env
ND_LYRICSPRIORITY=.lrc,.txt,embedded,navidrome-musixmatch-plugin
```

## Optional Website Fallback

If the app API returns no lyrics or the request fails, the plugin can fall back to the Musixmatch website. This fallback requires a Musixmatch website token. If no token is configured, the fallback is skipped.

To enable the website fallback, log in to Musixmatch and copy your user token using a Chromium-based browser:

1. Log in to [Musixmatch](https://www.musixmatch.com/).
2. Open your browser's developer tools (F12) and go to the "Application" tab.
3. Click on Cookies >> https://www.musixmatch.com.
4. Find the cookie named `musixmatchUserToken` and click on it.
5. In the bottom pane, copy the value of the cookie. Make sure "Show URL-decoded" is enabled.
6. Paste that token into "Musixmatch User Token" in the plugin settings.
7. If Musixmatch asks you to complete a CAPTCHA, complete it in your browser, then repeat the same steps for the `captcha_id` cookie and paste its value into "Musixmatch Captcha ID".


It's recommended to set this plugin's priority to the lowest position, as Musixmatch may limit or decline requests from third-party clients.

**This plugin relies on interfaces that are not part of Musixmatch's documented developer API. Musixmatch can change or restrict them at any time, so the plugin may stop working without notice. Please always download the latest version to check if your issue is already fixed.**

## This plugin shares analytics

To detect failed lookups, this plugin shares some analytics. Those analytics are opt-out, anonymous, and cannot be traced back to you. I only collect this information to improve the plugin. Analytics are stored on my self-hosted [OpenObserve](https://github.com/openobserve/openobserve) instance. The server is full-disk encrypted and only accessible to me. Analytics are not shared with any third party and only stored for 30 days.

I'd kindly ask you to keep analytics enabled, as it helps both me, in improving the plugin, and ultimately you, to get more lyrics. If you want to disable analytics, you can do so in the plugin settings.

## Reporting Issues

Before opening an [issue](https://github.com/Myzel394/navidrome-musixmatch-plugin/issues), grep your Navidrome logs and attach the matching lines:

```sh
grep navidrome-musixmatch-plugin /path/to/navidrome.log
```

## Related Plugins

I can highly recommend to install my other lyrics plugins as well to get the best lyrics coverage as possible:

* [navidrome-shazam-plugin](https://github.com/Myzel394/navidrome-shazam-plugin)
* [navidrome-lyricstranslate-plugin](https://github.com/Myzel394/navidrome-lyricstranslate-plugin)

