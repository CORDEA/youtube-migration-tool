package main

import (
	"encoding/json"
	"github.com/CORDEA/youtube-migration-tool/client"
	"github.com/CORDEA/youtube-migration-tool/repository"
	"google.golang.org/api/youtube/v3"
	"log"
	"os"
	"path/filepath"
)

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

type errorType int

const (
	subscription errorType = iota + 1
	playlist
	playlistItem
)

type migrationError struct {
	errType errorType
	errId   string
}

func (m *Migrator) migrate() {
	data, err := m.restoreMigrationData()
	if err != nil {
		data, err = m.fetchMigrationData()
		if err != nil {
			log.Fatalln(err)
		}
	}

	if err := m.cacheMigrationData(data); err != nil {
		log.Fatalln(err)
	}

	var errId string
	var errType errorType
	// TODO: Create request from response.
	for _, sub := range data.Subscriptions {
		if e := m.subscriptionRepo.Insert(client.Writing, sub); e != nil {
			errId = sub.Id
			errType = subscription
			err = e
			break
		}
	}

	for _, list := range data.Playlists {
		if e := m.playlistRepo.Insert(client.Writing, list.Playlist); e != nil {
			errId = list.Playlist.Id
			errType = playlist
			err = e
			break
		}
		for _, item := range list.Items {
			if e := m.playlistItemRepo.Insert(client.Writing, item); e != nil {
				errId = item.Id
				errType = playlistItem
				err = e
				break
			}
		}
	}

	if err == nil {
		return
	}

	if err := m.cacheRemainingData(data, migrationError{errType: errType, errId: errId}); err != nil {
		log.Println(err)
	}

	log.Fatalln(err)
}

func (m *Migrator) fetchMigrationData() (*Data, error) {
	subs, err := m.subscriptionRepo.FindAll(client.Reading)
	if err != nil {
		return nil, err
	}

	lists, err := m.playlistRepo.FindAll(client.Reading)
	if err != nil {
		return nil, err
	}

	var playlists []*Playlist
	for _, l := range lists {
		items, err := m.playlistItemRepo.Find(client.Reading, l.Id)
		if err != nil {
			return nil, err
		}
		playlists = append(playlists, &Playlist{Playlist: l, Items: items})
	}

	return &Data{Subscriptions: subs, Playlists: playlists}, nil
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

func (m *Migrator) restoreMigrationData() (*Data, error) {
	path := filepath.Join(cacheDir, "data.json")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data := &Data{}
	err = json.NewDecoder(f).Decode(data)
	defer f.Close()
	return data, err
}

func (m *Migrator) cacheRemainingData(data *Data, err migrationError) error {
	var newSubs []*youtube.Subscription
	var newList []*Playlist
	ignore := true
	if err.errType == subscription {
		for _, sub := range data.Subscriptions {
			if sub.Id == err.errId {
				ignore = false
			}
			if ignore {
				continue
			}
			newSubs = append(newSubs, sub)
		}

		return m.cacheMigrationData(&Data{Subscriptions: newSubs, Playlists: newList})
	}

	for _, list := range data.Playlists {
		if err.errType == playlist && list.Playlist.Id == err.errId {
			ignore = false
		}
		if !ignore {
			newList = append(newList, list)
			continue
		}
		if err.errType != playlistItem {
			continue
		}
		var newItems []*youtube.PlaylistItem
		for _, item := range list.Items {
			if item.Id == err.errId {
				ignore = false
			}
			if !ignore {
				newItems = append(newItems, item)
			}
		}
		newList = append(newList, &Playlist{Playlist: list.Playlist, Items: newItems})
	}

	return m.cacheMigrationData(&Data{Subscriptions: newSubs, Playlists: newList})
}
