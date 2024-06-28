# bsdl

A simple BeatStars music downloader built in Go, featuring concurrent downloads and metadata tagging.

## Usage 

### Download all tracks from a artist
```shell
bsdl artist lovbug
```
### Download a track from a link
```shell
bsdl beat https://www.beatstars.com/beat/lovbug-skoolio-renter_135bpm-15063259
```

## Installation


### MacOS

Install using [Homebrew](https://brew.sh/)

```shell
brew tap adithayyil/bsdl
brew install bsdl
```

### From source

Visit this link to install [Go](https://go.dev/doc/install).

Clone this repo and build
```shell
git clone https://github.com/adithayyil/bsdl.git
cd bsdl
go build
```

After building the project, an executable named `bsdl` is created, which you can then copy to the desired location.

### Pre-compiled
Download the pre-compiled binaries from the releases page and copy them to the desired location.