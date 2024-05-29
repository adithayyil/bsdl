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
	Title            string   `json:"title"`
	ArtistName       string   `json:"artist_name"`
	StreamURL        string   `json:"stream_url"`
	ReleaseTimestamp int64    `json:"releaseTimestamp"`
	ArtworkOriginal  string   `json:"artwork_original"`
	ID               int      `json:"v2Id"`
	BPM              int      `json:"bpm"`
	Genres           []string `json:"genres"`
	Tags             []string `json:"tags"`
}

func downloadFile(track Track, client *http.Client, single bool, streamOnly bool, p *mpb.Progress) error {
	url := track.StreamURL
	artistName := track.ArtistName
	filename := track.Title

	resp, err := client.Get(url)
	checkError(err)
	defer resp.Body.Close()

	bar := p.AddBar(resp.ContentLength,
		mpb.PrependDecorators(
			decor.Name(fmt.Sprintf("Downloading %s •", track.Title)),
		),
		mpb.AppendDecorators(
			decor.OnComplete(decor.EwmaETA(decor.ET_STYLE_MMSS, 60), "Downloaded!"),
		),
	)

	proxyReader := bar.ProxyReader(resp.Body)
	bodyBytes, err := io.ReadAll(proxyReader)
	checkError(err)

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

	if streamOnly {
		err = os.WriteFile(filePath, bodyBytes, 0644)
		checkError(err)
	} else {
		err = os.WriteFile(filePath, bodyBytes, 0644)
		checkError(err)
		embedMetadata(track, filePath)
	}

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

	unixTimeMillis := track.ReleaseTimestamp
	unixTime := unixTimeMillis
	if unixTimeMillis > 10000000000 {
		unixTime = unixTimeMillis / 1000
	}
	newModTime := time.Unix(unixTime, 0)
	fileInfo, _ := os.Stat(filePath)
	accessTime := fileInfo.ModTime()
	err = os.Chtimes(filePath, accessTime, newModTime)
	checkError(err, "Error changing file modification time: ")

	return nil
}
