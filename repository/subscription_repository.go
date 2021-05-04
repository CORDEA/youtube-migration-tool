package repository

import (
	"github.com/CORDEA/youtube-migration-tool/client"
	"google.golang.org/api/youtube/v3"
)

type SubscriptionRepository struct {
	client *client.YouTubeApiClient
}

func NewSubscriptionRepository(client *client.YouTubeApiClient) *SubscriptionRepository {
	return &SubscriptionRepository{client: client}
}

func (r *SubscriptionRepository) fetchSubscriptions(role client.Role, pageToken string) (*youtube.SubscriptionListResponse, error) {
	call := r.client.GetSubscriptionsService(role).List([]string{"snippet"})
	if pageToken != "" {
		call.PageToken(pageToken)
	}
	call.MaxResults(50)
	call.Mine(true)
	return call.Do()
}

func (r *SubscriptionRepository) FindAll(role client.Role) ([]*youtube.Subscription, error) {
	var subs []*youtube.Subscription
	res, err := r.fetchSubscriptions(role, "")
	if err != nil {
		return subs, err
	}
	num := res.PageInfo.TotalResults
	for ; ; {
		for _, s := range res.Items {
			subs = append(subs, s)
		}
		num -= res.PageInfo.ResultsPerPage
		if num <= 0 {
			break
		}
		res, err = r.fetchSubscriptions(role, res.NextPageToken)
		if err != nil {
			return subs, err
		}
	}
	return subs, nil
}
