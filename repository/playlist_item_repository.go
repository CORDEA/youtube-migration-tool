package repository

import (
	"github.com/CORDEA/youtube-migration-tool/client"
	"google.golang.org/api/youtube/v3"
)

type PlaylistItemRepository struct {
	client *client.YouTubeApiClient
}

func NewPlaylistItemRepository(client *client.YouTubeApiClient) *PlaylistItemRepository {
	return &PlaylistItemRepository{client: client}
}

func (r *PlaylistItemRepository) fetchPlaylistItems(role client.Role, id string, pageToken string) (*youtube.PlaylistItemListResponse, error) {
	call := r.client.GetPlaylistItemsService(role).List([]string{"id", "snippet"})
	if pageToken != "" {
		call.PageToken(pageToken)
	}
	call.PlaylistId(id)
	call.MaxResults(50)
	return call.Do()
}

func (r *PlaylistItemRepository) Find(role client.Role, id string) ([]*youtube.PlaylistItem, error) {
	var lists []*youtube.PlaylistItem
	res, err := r.fetchPlaylistItems(role, id, "")
	if err != nil {
		return lists, err
	}
	num := res.PageInfo.TotalResults
	for ; ; {
		for _, p := range res.Items {
			lists = append(lists, p)
		}
		num -= res.PageInfo.ResultsPerPage
		if num <= 0 {
			break
		}
		res, err = r.fetchPlaylistItems(role, id, res.NextPageToken)
		if err != nil {
			return lists, err
		}
	}
	return lists, nil
}
