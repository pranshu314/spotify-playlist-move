package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyAuth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
	"os"
	"regexp"
)

func main() {

	err1 := godotenv.Load(".env")
	if err1 != nil {
		fmt.Println(err1)
	}

	ctx := context.Background()

	config := clientcredentials.Config{
		ClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		ClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		TokenURL:     spotifyAuth.TokenURL,
		Scopes:       []string{spotifyAuth.ScopeUserReadPrivate, spotifyAuth.ScopeUserLibraryRead, spotifyAuth.ScopeUserReadEmail},
	}

	token, err := config.Token(ctx)
	if err != nil {
		fmt.Printf("Unable to get the token, error: %v\n", err)
	}

	httpClient := spotifyAuth.Authenticator{}.Client(ctx, token)
	client := spotify.New(httpClient)

	r := regexp.MustCompile(`playlist\/(.*)\?`)

	inp := ""
	fmt.Printf("Enter the playlist id: ")
	fmt.Scan(&inp)
	playlistId := fmt.Sprintln(r.FindString(inp))
	playlistId = playlistId[9 : len(playlistId)-2]
	// fmt.Println(playlistId)
	playlistId2 := "1ckDytqUi4BUYzs6HIhcAN"

	tracks, err := client.GetPlaylistItems(ctx, spotify.ID(playlistId2))
	if err != nil {
		fmt.Println(err)
	}

	for page := 1; ; page++ {
		fmt.Println(tracks)

		err = client.NextPage(ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	// fmt.Println(tracks)
	// fmt.Println()
	// fmt.Printf("%T\n", tracks)
	// fmt.Printf("Playlist had total %d tracks\n", tracks.Total)
}
