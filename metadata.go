package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type MetadataCache struct {
	cache []MediaItem
}

const diskCachePath = "/Users/tim/.gscreencache"

func InitializeMetadataCache(ctx context.Context, client *http.Client) *MetadataCache {
	cache := MetadataCache{}
	items := make(chan MediaItem, 10000)

	err := cache.loadDiskCache()
	if err != nil {
		log.Fatalln("Failed to read cached metadata from disk", err)
	}

	delayBeforeFirstFetch := 12 * time.Hour

	if len(cache.cache) == 0 {
		delayBeforeFirstFetch = 0
	}

	// Refresh cache periodically
	go func() {
		time.Sleep(delayBeforeFirstFetch)

	L:
		for {
			log.Println("Attempting to fetch/refresh metadata")
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

func (c *MetadataCache) Random(count int) (items []MediaItem, err error) {
	if len(c.cache) == 0 {
		return []MediaItem{}, fmt.Errorf("metadata hasn't been downloaded yet")
	}

	dedup := make(map[string]bool)

	for i := 0; i < count; i++ {
		i := rand.Intn(len(c.cache))
		item := c.cache[i]

		if _, ok := dedup[item.Id]; ok {
			continue
		}

		dedup[item.Id] = true
		items = append(items, c.cache[i])
	}

	return
}

func (c *MetadataCache) writeDiskCache() (err error) {
	file, err := os.OpenFile(diskCachePath, os.O_RDWR, 0666)
	if err != nil {
		log.Println("Failed to open metadata disk cache file", err)
		return
	}
	defer file.Close()

	//some actions happen here
	file.Truncate(0)
	file.Seek(0, 0)

	for _, item := range c.cache {
		file.Write([]byte(fmt.Sprintf("%s\n", item.Id)))
	}

	file.Sync()

	log.Printf("Wrote %d media Ids to disk\n", len(c.cache))

	return nil
}

// Assumes an empty cache
func (c *MetadataCache) loadDiskCache() (err error) {
	var lines []string

	file, err := os.Open(diskCachePath)
	if err != nil {
		log.Println("Failed to open metadata disk cache file", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		log.Println("Failed to read metadata disk cache file", err)
		return
	}

	for _, line := range lines {
		c.cache = append(c.cache, MediaItem{Id: strings.TrimSpace(line)})
	}

	log.Printf("Loaded %d media Ids from disk\n", len(lines))

	return nil
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
			continue
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
			log.Printf("Fetched all metadata pages; cache size: %d\n", len(c.cache))
			err := c.writeDiskCache()
			if err != nil {
				log.Println("Failed to cache metadata to disk", err)
				return err
			}

			return nil
		} else {
			nextPage = r.NextPageToken
		}
	}
}
