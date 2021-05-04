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

func (m *Migrator) cacheMigrationData(data *Data) error {
	for _, s := range data.subscriptions {
		path := filepath.Join(cacheDir, "s"+s.Id+".json")
		if err := m.storeData(path, s); err != nil {
			return err
		}
	}

	for _, l := range data.playlists {
		id := "p" + l.playlist.Id
		path := filepath.Join(cacheDir, id+".json")
		if err := m.storeData(path, l.playlist); err != nil {
			return err
		}
		if err := os.Mkdir(id, 0700); err != nil {
			return err
		}
		for _, i := range l.items {
			path = filepath.Join(cacheDir, id, "i"+i.Id+".json")
			if err := m.storeData(path, i); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Migrator) storeData(path string, v interface{}) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	err = json.NewEncoder(f).Encode(v)
	defer f.Close()
	return err
}

func (m *Migrator) migrate() {
	data := m.fetchMigrationData()

	if err := m.cacheMigrationData(data); err != nil {
		log.Fatalln(err)
	}

	if err := m.subscriptionRepo.Insert(client.Writing, data.subscriptions); err != nil {
		log.Fatalln(err)
	}

	for _, list := range data.playlists {
		if err := m.playlistRepo.Insert(client.Writing, list.playlist); err != nil {
			log.Fatalln(err)
		}
		for _, item := range list.items {
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
