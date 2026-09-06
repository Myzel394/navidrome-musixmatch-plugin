package musixmatch

import "regexp"

var WEBSITE_NEXT_DATA_REGEX = regexp.MustCompile(`(?s)<script id="__NEXT_DATA__" type="application/json">(.*?)</script>`)

type subtitleLine struct {
	Text string `json:"text"`
	Time struct {
		Total      float64 `json:"total"`
		Minutes    int     `json:"minutes"`
		Seconds    int     `json:"seconds"`
		Hundredths int     `json:"hundredths"`
	} `json:"time"`
	Type string `json:"type"`
}

type sectionLine struct {
	Title string         `json:"title"`
	Lines []subtitleLine `json:"lines"`
}

type nextDataResponse struct {
	Props struct {
		PageProps struct {
			Data struct {
				TrackInfo struct {
					Data struct {
						Track struct {
							AlbumName  string `json:"albumName"`
							ArtistName string `json:"artistName"`
						} `json:"track"`
						Lyrics struct {
							Body     string `json:"body"`
							Language string `json:"language"`
						} `json:"lyrics"`
						TrackStructureList []sectionLine  `json:"trackStructureList"`
						Subtitle           []subtitleLine `json:"subtitle"`
					} `json:"data"`
				} `json:"trackInfo"`
			} `json:"data"`
		} `json:"pageProps"`
	} `json:"props"`
}
