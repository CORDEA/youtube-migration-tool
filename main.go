package main

import (
	"context"
	"fmt"
	"github.com/CORDEA/youtube-migration-tool/youtube"
	youtube2 "google.golang.org/api/youtube/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

func fetchSubscriptions(service *youtube2.Service, pageToken string) (*youtube2.SubscriptionListResponse, error) {
	call := service.Subscriptions.List([]string{"snippet"})
	if pageToken != "" {
		call.PageToken(pageToken)
	}
	call.MaxResults(50)
	call.Mine(true)
	return call.Do()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	secret, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalln(err)
	}
	service, err := youtube.ProvideService(ctx, secret)
	if err != nil {
		log.Fatalln(err)
	}

	var subs []*youtube2.Subscription
	r, err := fetchSubscriptions(service, "")
	if err != nil {
		log.Fatalln(err)
	}
	num := r.PageInfo.TotalResults
	for ;; {
		for _, s := range r.Items {
			subs = append(subs, s)
		}
		num -= r.PageInfo.ResultsPerPage
		if num <= 0 {
			break
		}
		r, err = fetchSubscriptions(service, r.NextPageToken)
		if err != nil {
			log.Fatalln(err)
		}
	}

	for _, s := range subs {
		fmt.Println(s.Snippet.Title)
	}
}
