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

	r2 := regexp.MustCompile(`\[(.*)\]`)

	songs := []string{}

	for page := 1; ; page++ {
		for _, v := range tracks.Items {
			inp2 := fmt.Sprintln(v.Track)
			inp3 := fmt.Sprintln(v.Track.Track.Artists[0].Name)
			inp4 := fmt.Sprintln(v.Track.Track.Album.Name)
			songs = append(songs, fmt.Sprintf("%s by %s in %s", r2.FindString(inp2), inp3, inp4))
		}

		for i, v := range songs {
			fmt.Printf("%d. %s\n", i+1, v)
		}

		err = client.NextPage(ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
	}

}
