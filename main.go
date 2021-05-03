package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"net/http"
)

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
	// TODO
	url := config.AuthCodeURL("", oauth2.AccessTypeOffline)
	fmt.Println(url)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	return config.Exchange(ctx, code)
}

func main() {
	ctx := context.Background()
	secret, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalln()
	}

	_, err = getClient(ctx, secret, youtube.YoutubeReadonlyScope)
	if err != nil {
		log.Fatalln()
	}
}
