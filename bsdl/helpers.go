package bsdl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

type Track struct {
	Title           string   `json:"title"`
	ArtistName      string   `json:"artist_name"`
	StreamURL       string   `json:"stream_url"`
	ReleaseDate     string   `json:"release_date"`
	ArtworkOriginal string   `json:"artwork_original"`
	ID              int      `json:"v2Id"`
	BPM             int      `json:"bpm"`
	Genres          []string `json:"genres"`
	Tags            []string `json:"tags"`
}

func checkError(err error, info ...string) {
	if err != nil {
		log.Fatal(info, err)
	}
}

func downloadFile(track Track, client *http.Client, single bool, p *mpb.Progress) error {
	url := track.StreamURL
	artistName := track.ArtistName
	filename := track.Title

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := p.AddBar(resp.ContentLength,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Downloading %s: ", track.Title)),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 60), "Downloaded!"),
			decor.Name(" | "),
			decor.AverageSpeed(decor.UnitKB, "% .2f"),
		),
	)

	proxyReader := bar.ProxyReader(resp.Body)

	bodyBytes, err := io.ReadAll(proxyReader)
	if err != nil {
		return err
	}

	contentType := http.DetectContentType(bodyBytes[:512])
	var extension string
	if strings.Contains(contentType, "mpeg") {
		extension = ".mp3"
	} else if strings.Contains(contentType, "wav") {
		extension = ".wav"
	} else {
		// TODO - handle other content types
		extension = ".mp3"
	}

	var dirPath string
	if single {
		dirPath = "."
	} else {
		artistFolder := artistName
		if _, err := os.Stat(artistFolder); os.IsNotExist(err) {
			os.MkdirAll(artistFolder, 0755)
		}
		dirPath = artistFolder
	}

	filePath := dirPath + "/" + filename + extension
	err = os.WriteFile(filePath, bodyBytes, 0644)
	if err != nil {
		return err
	}

	embedMetadata(track, filePath)

	return nil
}

func embedMetadata(track Track, filePath string) error {
	file, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	checkError(err, "Error while opening audio file for tagging: ")
	defer file.Close()

	file.SetArtist(track.ArtistName)
	file.SetTitle(track.Title)
	file.SetGenre(strings.Join(track.Genres, ", "))
	file.AddFrame(file.CommonID("BPM"), id3v2.TextFrame{Encoding: id3v2.EncodingUTF8, Text: string(rune(track.BPM))})

	coverResp, _ := http.Get(track.ArtworkOriginal)
	if coverResp.StatusCode != http.StatusOK {
		fmt.Printf("Error: non-OK HTTP status: %v\n", coverResp.StatusCode)
	}
	defer coverResp.Body.Close()

	coverArt, err := io.ReadAll(coverResp.Body)
	checkError(err)

	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Description: "Cover",
		Picture:     coverArt,
	}
	file.AddAttachedPicture(pic)

	if err = file.Save(); err != nil {
		log.Fatal("Error while saving a tag: ", err)
	}

	dateString := track.ReleaseDate
	newModTime, err := time.Parse(time.RFC3339, dateString)
	checkError(err, "Error parsing date: ")
	fileInfo, _ := os.Stat(filePath)
	accessTime := fileInfo.ModTime()
	err = os.Chtimes(filePath, accessTime, newModTime)
	checkError(err, "Error changing file modification time: ")

	return nil
}

func makeHTTPRequest(client *http.Client, method, url string, data *strings.Reader) []byte {
	var req *http.Request
	var err error

	if data != nil {
		req, err = http.NewRequest(method, url, data)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	checkError(err)

	req.Header = getDefaultHeaders()
	resp, err := client.Do(req)
	checkError(err)
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	checkError(err)

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
