package main

import (
	"context"
	"github.com/CORDEA/youtube-migration-tool/client"
	"github.com/CORDEA/youtube-migration-tool/repository"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

type Migrator struct {
	subscriptionRepo *repository.SubscriptionRepository
	playlistRepo     *repository.PlaylistRepository
	playlistItemRepo *repository.PlaylistItemRepository
}

type Data struct {
	subscriptions []*youtube.Subscription
	playlists     []*Playlist
}

type Playlist struct {
	playlist *youtube.Playlist
	items    []*youtube.PlaylistItem
}

func (m *Migrator) fetchMigrationData() *Data {
	subs, err := m.subscriptionRepo.FindAll(client.Reading)
	if err != nil {
		log.Println(err)
	}

	lists, err := m.playlistRepo.FindAll(client.Reading)
	if err != nil {
		log.Println(err)
	}

	var playlists []*Playlist
	for _, l := range lists {
		items, err := m.playlistItemRepo.Find(client.Reading, l.Id)
		if err != nil {
			log.Println(err)
		}
		playlists = append(playlists, &Playlist{playlist: l, items: items})
		break
	}

	return &Data{subscriptions: subs, playlists: playlists}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	secret, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalln(err)
	}
	apiClient, err := client.NewYouTubeApiClient(ctx, secret)
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

	_ = migrator.fetchMigrationData()
}
