package lyrics_writer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
	"github.com/navidrome/navidrome/plugins/pdk/go/host"
	"github.com/navidrome/navidrome/plugins/pdk/go/lyrics"
)

func WriteLyrics(
	track lyrics.TrackInfo,
	lyrics lyrics.GetLyricsResponse,
	format utils.LyricsFormat,
) error {
	utils.LogInfof("writing lyrics for track: %s - %s", track.Artist, track.Title)

	library, err := host.LibraryGetLibrary(track.LibraryID)
	if err != nil || library == nil {
		return fmt.Errorf("could not access library %d: %w", track.LibraryID, err)
	}

	trackPath := filepath.Join(library.MountPoint, track.Path)
	extension := getExtensionForLyricsFormat(format)
	basename := filepath.Base(trackPath)
	newFilename := basename[:len(basename)-len(filepath.Ext(basename))] + extension
	lyricsPath := filepath.Join(filepath.Dir(trackPath), newFilename)
	lyricsFile, err := os.Create(lyricsPath)
	if err != nil {
		utils.LogErrorf("failed to create lyrics file: %v", err)
		return err
	}

	_, err = lyricsFile.WriteString(lyrics.Lyrics[0].Text)
	if err != nil {
		utils.LogErrorf("Failed to write lyrics to file %s: %v", lyricsPath, err)
		return err
	}

	utils.LogInfof("successfully wrote lyrics to file %s", lyricsPath)

	return nil
}
