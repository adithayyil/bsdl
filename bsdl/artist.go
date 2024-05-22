package bsdl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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

type Track struct {
	Title           string   `json:"title"`
	ArtistName      string   `json:"artist_name"`
	StreamURL       string   `json:"stream_url"`
	ReleaseDate     string   `json:"release_date"`
	ArtworkOriginal string   `json:"artwork_original"`
	ID              string   `json:"id"`
	BPM             int      `json:"bpm"`
	Genres          []string `json:"genres"`
	Tags            []string `json:"tags"`
}

func DownloadArtistMusic(permalink string) {
	client := &http.Client{}
	tracks := getArtists(permalink, client)

	fmt.Println("Total number of tracks: ", len(tracks))
}

func getArtists(permalink string, client *http.Client) []Track {
	page := 0
	var allTracks []Track

	artistIDURL := fmt.Sprintf("https://main.v2.beatstars.com/musician?permalink=%s", permalink)
	bodyTextArtistID := makeHTTPRequest(client, "GET", artistIDURL, nil)
	var musician MusicianResponse
	json.Unmarshal(bodyTextArtistID, &musician)
	userID := int(musician.Response.Data.Profile.UserID)
	memberId := fmt.Sprintf("MR%d", userID)

	reqURL := "https://nmmgzjq6qi-dsn.algolia.net/1/indexes/public_prod_inventory_track_index_bycustom/query?x-algolia-agent=Algolia%20for%20JavaScript%20(4.12.0)%3B%20Browser"

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
		bodyText := makeHTTPRequest(client, "POST", reqURL, strings.NewReader(data))

		var result map[string]interface{}
		json.Unmarshal(bodyText, &result)

		nbPages := int(result["nbPages"].(float64))

		hits := result["hits"].([]interface{})
		for _, hit := range hits {
			hitMap, ok := hit.(map[string]interface{})
			if !ok {
				log.Fatalf("Could not convert hit to map[string]interface{}")
			}

			artwork := hitMap["artwork"].(map[string]interface{})
			bundle := hitMap["bundle"].(map[string]interface{})
			metadata := hitMap["metadata"].(map[string]interface{})

			var genres, tags []string
			if metadata["genres"] != nil {
				genres = convertInterfaceSliceToStringSlice(metadata["genres"].([]interface{}))
			}
			if metadata["tags"] != nil {
				tags = convertInterfaceSliceToStringSlice(metadata["tags"].([]interface{}))
			}

			track := Track{
				ArtworkOriginal: artwork["sizes"].(map[string]interface{})["original"].(string),
				StreamURL:       bundle["stream"].(map[string]interface{})["url"].(string),
				ReleaseDate:     hitMap["releaseDate"].(string),
				ID:              hitMap["id"].(string),
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

func makeHTTPRequest(client *http.Client, method, url string, data *strings.Reader) []byte {
	var req *http.Request
	var err error

	if data != nil {
		req, err = http.NewRequest(method, url, data)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		log.Fatal(err)
	}
	req.Header = getDefaultHeaders()
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return bodyText
}

func getDefaultHeaders() http.Header {
	headers := http.Header{}
	headers.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:125.0) Gecko/20100101 Firefox/125.0")
	headers.Set("Accept", "*/*")
	headers.Set("Accept-Language", "en-CA,en-US;q=0.7,en;q=0.3")
	headers.Set("x-algolia-api-key", "b3513eb709fe8f444b4d5c191b63ea47") // this is a public api key lol. not a secret.
	headers.Set("x-algolia-application-id", "NMMGZJQ6QI")
	headers.Set("content-type", "application/x-www-form-urlencoded")
	headers.Set("Origin", "https://www.beatstars.com")
	headers.Set("Connection", "keep-alive")
	headers.Set("Referer", "https://www.beatstars.com/")
	headers.Set("Sec-Fetch-Dest", "empty")
	headers.Set("Sec-Fetch-Mode", "cors")
	headers.Set("Sec-Fetch-Site", "cross-site")
	headers.Set("Sec-GPC", "1")
	headers.Set("Pragma", "no-cache")
	headers.Set("Cache-Control", "no-cache")

	return headers
}
