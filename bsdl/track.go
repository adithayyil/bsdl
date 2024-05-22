package bsdl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/vbauerster/mpb"
)

type TrackDetails struct {
	Title    string `json:"title"`
	Musician struct {
		ArtistName string `json:"display_name"`
	} `json:"musician"`
	StreamURL string `json:"stream_url"`
	Artwork   struct {
		Original string `json:"original"`
	} `json:"artwork"`
	ID               int      `json:"track_id"`
	ReleaseTimestamp int64    `json:"release_date_time"`
	BPM              int      `json:"bpm"`
	Genres           []string `json:"genre"`
	Tags             []string `json:"tags"`
}

type TrackData struct {
	Response struct {
		Data struct {
			Details TrackDetails `json:"details"`
		} `json:"data"`
	} `json:"response"`
}

func downloadTrack(link string) {
	regex := regexp.MustCompile(`/beat/.*?-(\d+)$`)
	match := regex.FindStringSubmatch(link)

	if len(match) > 1 {
		client := &http.Client{}
		trackID := match[1]
		track := getTrack(trackID, client)

		if track.ID == 0 {
			log.Fatalf("Track from link %s not found\n", link)
		} else {
			p := mpb.New()
			downloadFile(track, client, true, p)

			fmt.Println()
			log.Printf("Downloaded %s by %s\n", track.Title, track.ArtistName)
		}
	} else {
		log.Println("Invalid link")
	}
}

func getTrack(trackID string, client *http.Client) Track {
	trackDataURL := fmt.Sprintf("https://main.v2.beatstars.com/beat?id=%s&fields=details", trackID)
	trackDataReq := makeHTTPRequest(client, "GET", trackDataURL, nil)

	var trackData TrackData
	json.Unmarshal(trackDataReq, &trackData)

	trackDetails := trackData.Response.Data.Details

	track := Track{
		ArtworkOriginal:  trackDetails.Artwork.Original,
		StreamURL:        trackDetails.StreamURL,
		ID:               trackDetails.ID,
		Title:            trackDetails.Title,
		ArtistName:       trackDetails.Musician.ArtistName,
		ReleaseTimestamp: trackDetails.ReleaseTimestamp,
		BPM:              trackDetails.BPM,
		Genres:           trackDetails.Genres,
		Tags:             trackDetails.Tags,
	}

	return track
}
