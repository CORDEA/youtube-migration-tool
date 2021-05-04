package main

import (
	"context"
	"encoding/json"
	"github.com/CORDEA/youtube-migration-tool/client"
	"github.com/CORDEA/youtube-migration-tool/repository"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const cacheDir = ".cache"

type Migrator struct {
	subscriptionRepo *repository.SubscriptionRepository
	playlistRepo     *repository.PlaylistRepository
	playlistItemRepo *repository.PlaylistItemRepository
}

type Data struct {
	Subscriptions []*youtube.Subscription `json:"subscriptions"`
	Playlists     []*Playlist             `json:"playlists"`
}

type Playlist struct {
	Playlist *youtube.Playlist       `json:"playlist"`
	Items    []*youtube.PlaylistItem `json:"items"`
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
		playlists = append(playlists, &Playlist{Playlist: l, Items: items})
		break
	}

	return &Data{Subscriptions: subs, Playlists: playlists}
}

func (m *Migrator) cacheMigrationData(data *Data) error {
	path := filepath.Join(cacheDir, "data.json")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(data)
	defer f.Close()
	return err
}

func (m *Migrator) migrate() {
	data := m.fetchMigrationData()

	if err := m.cacheMigrationData(data); err != nil {
		log.Fatalln(err)
	}

	if err := m.subscriptionRepo.Insert(client.Writing, data.Subscriptions); err != nil {
		log.Fatalln(err)
	}

	for _, list := range data.Playlists {
		if err := m.playlistRepo.Insert(client.Writing, list.Playlist); err != nil {
			log.Fatalln(err)
		}
		for _, item := range list.Items {
			if err := m.playlistItemRepo.Insert(client.Writing, item); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

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

	_ = migrator.fetchMigrationData()
}
