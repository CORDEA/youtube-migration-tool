package youtube

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

func ProvideService(ctx context.Context, secret []byte) (*youtube.Service, error) {
	client, err := getClient(ctx, secret, youtube.YoutubeReadonlyScope)
	if err != nil {
		return nil, err
	}
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return service, nil
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
		state += fmt.Sprintf("%c", rand.Intn(26) + 97)
	}
	url := config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println(url)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	return config.Exchange(ctx, code)
}
