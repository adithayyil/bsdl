package bsdl

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "bsdl",
	Version: "0.0.0",
	Short:   `BeatStars Music Downloader`,
}

var artist = &cobra.Command{
	Use:     "artist [permalink]",
	Aliases: []string{"a"},
	Short:   "Download all tracks from an artist",
	Long:    `Download all tracks from an artist on BeatStars.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		artistPermalink := args[0]
		downloadArtistTracks(artistPermalink)
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
		downloadTrack(link)
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(artist)
	rootCmd.AddCommand(track)
}
