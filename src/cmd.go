package bsdl

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "bsdl",
	Version: "0.6",
	Short:   `BeatStars Music Downloader`,
}

var streamOnly bool
var threads int

var artist = &cobra.Command{
	Use:     "artist [permalink]",
	Aliases: []string{"a"},
	Short:   "Download all tracks from an artist",
	Long:    `Download all tracks from an artist on BeatStars.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		artistPermalink := args[0]
		downloadArtistTracks(artistPermalink, streamOnly, threads)
	},
}

var track = &cobra.Command{
	Use:     "track [link]",
	Aliases: []string{"t"},
	Short:   "Download a track from a link",
	Long:    `Download a track from a link on BeatStars.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		link := args[0]
		downloadTrack(link, streamOnly)
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	artist.PersistentFlags().BoolVar(&streamOnly, "stream-only", false, "Get streams only and don't embed metadata")
	artist.PersistentFlags().IntVar(&threads, "threads", 6, "Number of concurrent downloads")
	track.PersistentFlags().BoolVar(&streamOnly, "stream-only", false, "Get streams only and don't embed metadata")
	rootCmd.AddCommand(artist)
	rootCmd.AddCommand(track)
}
