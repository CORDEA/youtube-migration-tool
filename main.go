package main

import (
	"context"
	"github.com/CORDEA/youtube-migration-tool/client"
	"github.com/CORDEA/youtube-migration-tool/repository"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const cacheDir = ".cache"

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	secret, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalln(err)
	}
	apiClient, err := client.NewYouTubeApiClient(ctx, secret, cacheDir)
	if err != nil {
		log.Fatalln(err)
	}
	subscriptionRepo := repository.NewSubscriptionRepository(apiClient)
	playlistRepo := repository.NewPlaylistRepository(apiClient)
	playlistItemRepo := repository.NewPlaylistItemRepository(apiClient)
	migrator := Migrator{
		subscriptionRepo: subscriptionRepo,
		playlistRepo:     playlistRepo,
		playlistItemRepo: playlistItemRepo,
	}

	migrator.migrate()
}
