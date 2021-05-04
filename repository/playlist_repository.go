package repository

import (
	"github.com/CORDEA/youtube-migration-tool/client"
	"google.golang.org/api/youtube/v3"
)

type PlaylistRepository struct {
	client *client.YouTubeApiClient
}

func NewPlaylistRepository(client *client.YouTubeApiClient) *PlaylistRepository {
	return &PlaylistRepository{client: client}
}

func (r *PlaylistRepository) fetchPlaylists(role client.Role, pageToken string) (*youtube.PlaylistListResponse, error) {
	call := r.client.GetPlaylistsService(role).List([]string{"id", "snippet"})
	if pageToken != "" {
		call.PageToken(pageToken)
	}
	call.Mine(true)
	call.MaxResults(50)
	return call.Do()
}

func (r *PlaylistRepository) FindAll(role client.Role) ([]*youtube.Playlist, error) {
	var lists []*youtube.Playlist
	res, err := r.fetchPlaylists(role, "")
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
		res, err = r.fetchPlaylists(role, res.NextPageToken)
		if err != nil {
			return lists, err
		}
	}
	return lists, nil
}
