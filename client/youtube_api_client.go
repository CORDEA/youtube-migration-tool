package client

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"math/rand"
	"net/http"
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

func NewYouTubeApiClient(ctx context.Context, secret []byte) (*YouTubeApiClient, error) {
	rClient, err := getClient(ctx, secret, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, err
	}
	rService, err := youtube.NewService(ctx, option.WithHTTPClient(rClient))
	if err != nil {
		return nil, err
	}
	wClient, err := getClient(ctx, secret, youtube.YoutubeScope)
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

func getClient(ctx context.Context, secret []byte, scope ...string) (*http.Client, error) {
	config, err := google.ConfigFromJSON(secret, scope...)
	if err != nil {
		return nil, err
	}
	token, err := requestToken(ctx, config)
	if err != nil {
		return nil, err
	}
	return config.Client(ctx, token), nil
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
