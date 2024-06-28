package bsdl

import (
	"log"
	"net/http"
	"strings"
)

func checkError(err error, info ...string) {
	if err != nil {
		log.Fatal(info, err)
	}
}

func checkStatusCode(statusCode int, info ...string) {
	if statusCode == 404 {
		log.Fatal("404 Not found ", info)
	}

	if statusCode != 200 {
		log.Fatal(statusCode, info)
	}
}

func makeHTTPRequest(client *http.Client, method, url string, data *strings.Reader) *http.Response {
	var request *http.Request
	var err error

	if data != nil {
		request, err = http.NewRequest(method, url, data)
	} else {
		request, err = http.NewRequest(method, url, nil)
	}
	checkError(err)

	request.Header = getDefaultHeaders()
	response, err := client.Do(request)
	checkError(err)
	checkStatusCode(response.StatusCode, "Trouble accessing: ", url)

	return response
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
