package client

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
)

type YouTubeApiClient struct {
	readingService *youtube.Service
	writingService *youtube.Service
}

type Role int

const (
	Reading Role = iota + 1
	Writing
)

func NewYouTubeApiClient(ctx context.Context, secret []byte, cacheDir string) (*YouTubeApiClient, error) {
	rClient, err := getClient(ctx, cacheDir, secret, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, err
	}
	rService, err := youtube.NewService(ctx, option.WithHTTPClient(rClient))
	if err != nil {
		return nil, err
	}
	wClient, err := getClient(ctx, cacheDir, secret, youtube.YoutubeScope)
	if err != nil {
		return nil, err
	}
	wService, err := youtube.NewService(ctx, option.WithHTTPClient(wClient))
	if err != nil {
		return nil, err
	}
	return &YouTubeApiClient{readingService: rService, writingService: wService}, nil
}

func (c *YouTubeApiClient) GetSubscriptionsService(role Role) *youtube.SubscriptionsService {
	if role == Writing {
		return c.writingService.Subscriptions
	}
	return c.readingService.Subscriptions
}

func (c *YouTubeApiClient) GetPlaylistsService(role Role) *youtube.PlaylistsService {
	if role == Writing {
		return c.writingService.Playlists
	}
	return c.readingService.Playlists
}

func (c *YouTubeApiClient) GetPlaylistItemsService(role Role) *youtube.PlaylistItemsService {
	if role == Writing {
		return c.writingService.PlaylistItems
	}
	return c.readingService.PlaylistItems
}

func getClient(ctx context.Context, cacheDir string, secret []byte, scope ...string) (*http.Client, error) {
	config, err := google.ConfigFromJSON(secret, scope...)
	if err != nil {
		return nil, err
	}

	if err := os.Mkdir(cacheDir, 0700); err != nil {
		return nil, err
	}
	path := filepath.Join(cacheDir, "credentials.json")
	token, err := restoreToken(path)
	if err == nil {
		return config.Client(ctx, token), nil
	}

	token, err = requestToken(ctx, config)
	if err != nil {
		return nil, err
	}
	client := config.Client(ctx, token)
	err = storeToken(path, token)
	return client, err
}

func requestToken(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	state := ""
	for i := 0; i < 3; i++ {
		state += fmt.Sprintf("%c", rand.Intn(26)+97)
	}
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println(url)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	return config.Exchange(ctx, code)
}

func restoreToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

func storeToken(path string, token *oauth2.Token) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
