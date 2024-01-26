package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyAuth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
`

func buildOAuthHTTPClient() *http.Client {
	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/youtube-go-quickstart.json
	config, err := google.ConfigFromJSON(b, youtube.YoutubepartnerScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return getClient(ctx, config)
}

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("youtube-go-quickstart.json")), err
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func handleError(err error, message string) {
	if message == "" {
		message = "Error making API call"
	}
	if err != nil {
		log.Fatalf(message+": %v", err.Error())
	}
}

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
	fmt.Println(playlistId)
	// playlistId2 := "1ckDytqUi4BUYzs6HIhcAN"

	tracks, err := client.GetPlaylistItems(ctx, spotify.ID(playlistId))
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

		// for i, v := range songs {
		// 	fmt.Printf("%d. %s\n", i+1, v)
		// }

		err = client.NextPage(ctx, tracks)
		if err == spotify.ErrNoMorePages {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	ytDevKey := os.Getenv("YOUTUBE_API_KEY")
	ytClient := &http.Client{
		Transport: &transport.APIKey{Key: ytDevKey},
	}

	service, err := youtube.New(ytClient)
	if err != nil {
		fmt.Println(err)
	}

	service2, err := youtube.New(buildOAuthHTTPClient())
	if err != nil {
		fmt.Println(err)
	}

	title := ""
	fmt.Printf("Enter the title of the playlist: ")
	fmt.Scan(&title)

	playlist := &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       title,
			Description: "Playlist created by youtube data api v3",
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	}

	songsMap := make(map[string]string)

	for _, val := range songs {
		call := service.Search.List([]string{"id", "snippet"}).Q(val).MaxResults(1)
		response, err := call.Do()
		if err != nil {
			fmt.Println(err)
		}

		for _, item := range response.Items {
			songsMap[item.Id.VideoId] = item.Snippet.Title
			// fmt.Println(item.Id.VideoId, item.Snippet.Title)
			call2 := service2.Playlists.Insert([]string{"snippet,status"}, playlist)
			_, err := call2.Do()
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}
