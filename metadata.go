package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type MetadataCache struct {
	cache []MediaItem
}

func InitializeMetadataCache(ctx context.Context, client *http.Client) *MetadataCache {
	cache := MetadataCache{}
	items := make(chan MediaItem, 10000)

	// Refresh cache periodically
	go func() {
	L:
		for {
			log.Println("Attempting to fetch/refresh metadata")
			cache.cache = []MediaItem{}
			err := cache.fetch(client, items)
			if err != nil {
				log.Println("Failed to fetch metadata", err)
			}

			select {
			case <-ctx.Done():
				break L
			case <-time.After(12 * time.Hour):
				// repeat
			}
		}
	}()

	go func() {
	L:
		for {
			select {
			case <-ctx.Done():
				break L
			case item := <-items:
				cache.cache = append(cache.cache, item)
			}
		}
	}()

	return &cache
}

func (c *MetadataCache) Random() MediaItem {
	i := rand.Intn(len(c.cache))
	return c.cache[i]
}

func (c *MetadataCache) fetch(client *http.Client, out chan MediaItem) error {
	var nextPage string
	i := 1
	url := "https://photoslibrary.googleapis.com/v1/mediaItems?pageSize=100"

	dedup := make(map[string]bool)
	for _, v := range c.cache {
		dedup[v.Id] = true
	}

	for {
		if nextPage != "" {
			url = fmt.Sprintf("https://photoslibrary.googleapis.com/v1/mediaItems?pageSize=100&pageToken=%s", nextPage)
		}

		resp, err := client.Get(url)
		if err != nil {
			log.Println("Failed to fetch metadata", err)
			return err
		}

		if resp.StatusCode != 200 {
			log.Println("Received non-200 response while fetching metadata", resp.StatusCode)
			return err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Failed to read metadata response", err)
			return err
		}

		r := &ListMediaItemsResponse{}
		err = json.Unmarshal(body, &r)
		if err != nil {
			log.Println("Failed to parse metadata response", err)
			return err
		}

		go func() {
			for _, item := range r.MediaItems {
				if _, ok := dedup[item.Id]; !ok {
					out <- item
				}
			}
		}()

		log.Printf("Fetched metadata page #%d\n", i)
		i++

		if r.NextPageToken == "" {
			return nil
		} else {
			nextPage = r.NextPageToken
		}
	}
}
