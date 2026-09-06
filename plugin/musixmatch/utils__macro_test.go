package musixmatch

import (
	"encoding/json"
	"testing"

	"github.com/Myzel394/navidrome-musixmatch-plugin/plugin/utils"
)

func TestParseMatchedTrackMetadataValid(t *testing.T) {
	call := macroResponse{}
	call.Message.Header.StatusCode = utils.HTTPStatusOK
	call.Message.Body = json.RawMessage(`{"track":{"track_name":"Song","artist_name":"Artist","album_name":"Album"}}`)

	meta, err := parseResponseToTrackMetadata(call)
	if err != nil {
		t.Fatalf("parseMatchedTrackMetadata returned error: %v", err)
	}
	if meta.Artist != "Artist" || meta.Title != "Song" || meta.Album != "Album" {
		t.Fatalf("unexpected metadata: %#v", meta)
	}
}

func TestParseMatchedTrackMetadataUnavailable(t *testing.T) {
	for _, call := range []macroResponse{{}, func() macroResponse {
		c := macroResponse{}
		c.Message.Header.StatusCode = 404
		c.Message.Body = json.RawMessage(`{"track":{"track_name":"Song"}}`)
		return c
	}()} {
		meta, err := parseResponseToTrackMetadata(call)
		if err != nil {
			t.Fatalf("parseMatchedTrackMetadata returned error for unavailable metadata: %v", err)
		}
		if meta != (trackMetadata{}) {
			t.Fatalf("expected zero metadata, got %#v", meta)
		}
	}
}

func TestParseMatchedTrackMetadataMalformedJSON(t *testing.T) {
	call := macroResponse{}
	call.Message.Header.StatusCode = utils.HTTPStatusOK
	call.Message.Body = json.RawMessage(`not-json`)

	meta, err := parseResponseToTrackMetadata(call)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if meta != (trackMetadata{}) {
		t.Fatalf("expected zero metadata on parse error, got %#v", meta)
	}
}
