package main

import (
	"context"
	"github.com/CORDEA/youtube-migration-tool/youtube"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	secret, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalln()
	}
	_, err = youtube.ProvideService(ctx, secret)
	if err != nil {
		log.Fatalln()
	}
}
