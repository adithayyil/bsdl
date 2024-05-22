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
	Short:   "Download music from an artist",
	Long:    `Download music from an artist on BeatStars.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		artistPermalink := args[0]
		DownloadArtistMusic(artistPermalink)
	},
}

func Execute() {
	rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(artist)
}
