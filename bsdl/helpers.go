package bsdl

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func downloadFile(url, filename string, client *http.Client, artistName string) error {
	resp, err := client.Get(url)
	checkError(err)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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

	dirPath := filepath.Join(".", artistName)
	checkError(err)
	out, err := os.Create(filepath.Join(dirPath, filename+extension))
	checkError(err)

	defer out.Close()

	_, err = out.Write(bodyBytes)
	checkError(err)

	fmt.Printf("Downloaded: %s\n", filename+extension)
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
