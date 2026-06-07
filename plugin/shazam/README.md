# Architecture

Here is the actual code for scraping Shazam.

* `main.go`: is the entry file; this first searches for the song uses the search API, then scrapes the lyrics from the song page
* `search.go`: searches for the song using the Shazam web search API, and return the track ID
* `fetch.go`: scrapes the lyrics from the song page

The other files are mostly helpers
