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

func downloadArtistTracks(permalink string) {
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
		sem := make(chan struct{}, 10)

		for _, track := range tracks {
			wg.Add(1)

			go func(track Track) {
				sem <- struct{}{}

				defer wg.Done()
				err := downloadFile(track, client, false, p)
				if err != nil {
					log.Printf("Failed to download track: %s by %s. Error: %v\n", track.Title, track.ArtistName, err)
					errChan <- err
				}

				<-sem
			}(track)
		}

		wg.Wait()

		fmt.Println()
		log.Printf("- Downloaded all tracks for %s\n", permalink)
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

		var artistData map[string]interface{}
		json.Unmarshal(artistDataResp, &artistData)

		nbPages := int(artistData["nbPages"].(float64))

		hits := artistData["hits"].([]interface{})
		for _, hit := range hits {
			hitMap, ok := hit.(map[string]interface{})
			if !ok {
				log.Fatalf("Could not convert hit to map[string]interface{}")
			}

			artwork := hitMap["artwork"].(map[string]interface{})
			metadata := hitMap["metadata"].(map[string]interface{})

			var genres, tags []string
			if metadata["genres"] != nil {
				genres = convertInterfaceSliceToStringSlice(metadata["genres"].([]interface{}))
			}
			if metadata["tags"] != nil {
				tags = convertInterfaceSliceToStringSlice(metadata["tags"].([]interface{}))
			}

			trackID := int(hitMap["v2Id"].(float64))
			streamURL := fmt.Sprintf("https://main.v2.beatstars.com/stream?id=%d&return=audio", trackID)

			track := Track{
				ArtworkOriginal: artwork["sizes"].(map[string]interface{})["original"].(string),
				StreamURL:       streamURL,
				ReleaseDate:     hitMap["releaseDate"].(string),
				ID:              trackID,
				ArtistName:      metadata["artistName"].(string),
				BPM:             int(metadata["bpm"].(float64)),
				Genres:          genres,
				Tags:            tags,
				Title:           hitMap["title"].(string),
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

func convertInterfaceSliceToStringSlice(slice []interface{}) []string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = v.(string)
	}
	return strSlice
}
