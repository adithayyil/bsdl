package bsdl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/vbauerster/mpb"
)

type MusicianResponse struct {
	Response struct {
		Data struct {
			Profile struct {
				UserID float64 `json:"user_id"`
			} `json:"profile"`
		} `json:"data"`
	} `json:"response"`
}

type Artwork struct {
	Sizes struct {
		Original string `json:"original"`
	} `json:"sizes"`
}

type Metadata struct {
	ArtistName string   `json:"artistName"`
	BPM        int      `json:"bpm"`
	Genres     []string `json:"genres"`
	Tags       []string `json:"tags"`
}

type Hit struct {
	Artwork          Artwork  `json:"artwork"`
	Metadata         Metadata `json:"metadata"`
	V2Id             int      `json:"v2Id"`
	ReleaseTimestamp int64    `json:"releaseTimestamp"`
	Title            string   `json:"title"`
}

type ArtistData struct {
	NbPages int   `json:"nbPages"`
	Hits    []Hit `json:"hits"`
}

func downloadArtistTracks(permalink string, streamOnly bool, threads int) {
	if threads < 1 {
		log.Fatal("Number of threads must be greater than 1")
	}

	client := &http.Client{}
	log.Println("Retrieving artist tracks...")
	tracks := getArtistTracks(permalink, client)

	if len(tracks) == 0 {
		log.Fatalf("No tracks found for artist with permalink: %s\n", permalink)
	} else {
		log.Printf("Retrieved %d tracks from %s :3\n", len(tracks), permalink)
		fmt.Println()
		errChan := make(chan error, len(tracks))
		var wg sync.WaitGroup

		p := mpb.New()

		// Create a buffered channel to limit the number of concurrent downloads
		sem := make(chan struct{}, threads)

		for _, track := range tracks {
			wg.Add(1)

			go func(track Track) {
				sem <- struct{}{}

				defer wg.Done()
				err := downloadFile(track, client, false, streamOnly, p)
				if err != nil {
					log.Printf("Failed to download track: %s by %s. Error: %v\n", track.Title, track.ArtistName, err)
					errChan <- err
				}

				<-sem
			}(track)
		}

		wg.Wait()

		fmt.Println()
		log.Printf("Downloaded all %d tracks for %s\n", len(tracks), permalink)
		close(errChan)

		for err := range errChan {
			if err != nil {
				log.Printf("Download error: %v\n", err)
			}
		}
	}
}

func getArtistTracks(permalink string, client *http.Client) []Track {
	page := 0
	var allTracks []Track

	artistIDURL := fmt.Sprintf("https://main.v2.beatstars.com/musician?permalink=%s", permalink)
	bodyTextArtistID := makeHTTPRequest(client, "GET", artistIDURL, nil)
	var musician MusicianResponse
	json.Unmarshal(bodyTextArtistID, &musician)
	userID := int(musician.Response.Data.Profile.UserID)
	memberId := fmt.Sprintf("MR%d", userID)

	queryURL := "https://nmmgzjq6qi-dsn.algolia.net/1/indexes/public_prod_inventory_track_index_bycustom/query?x-algolia-agent=Algolia%20for%20JavaScript%20(4.12.0)%3B%20Browser"

	for {
		data := fmt.Sprintf(`{
			"query": "",
			"page": %d,
			"hitsPerPage": 1000,
			"facets": ["*"],
			"analytics": false,
			"tagFilters": [],
			"facetFilters": [["profile.memberId:%s"]],
			"maxValuesPerFacet": 1000,
			"enableABTest": false,
			"userToken": null,
			"filters": "",
			"ruleContexts": []
		}`, page, memberId)
		artistDataResp := makeHTTPRequest(client, "POST", queryURL, strings.NewReader(data))

		var artistData ArtistData
		json.Unmarshal(artistDataResp, &artistData)

		nbPages := artistData.NbPages

		for _, hit := range artistData.Hits {
			trackID := int(hit.V2Id)
			streamURL := fmt.Sprintf("https://main.v2.beatstars.com/stream?id=%d&return=audio", trackID)

			track := Track{
				ArtworkOriginal:  hit.Artwork.Sizes.Original,
				StreamURL:        streamURL,
				ReleaseTimestamp: hit.ReleaseTimestamp,
				ID:               trackID,
				ArtistName:       hit.Metadata.ArtistName,
				BPM:              hit.Metadata.BPM,
				Genres:           hit.Metadata.Genres,
				Tags:             hit.Metadata.Tags,
				Title:            hit.Title,
			}

			allTracks = append(allTracks, track)
		}

		page++
		if page >= nbPages {
			break
		}
	}

	return allTracks
}
