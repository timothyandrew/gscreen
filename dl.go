package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CachedImage struct {
	mediaItem MediaItem
	data      []byte
}

type DownloadCache = chan CachedImage

func InitializeDownloadCache(ctx context.Context, client *http.Client, metadata *MetadataCache) DownloadCache {
	cache := make(chan CachedImage, 10)

	go func() {
		for {
			item, err := metadata.Random()
			if err != nil {
				log.Println("Couldn't get a random image's metadata; going to wait and continue", err)
				time.Sleep(2 * time.Second)
				continue
			}

			url, err := GetImageUrl(client, item)
			if err != nil {
				log.Println("GetImageUrl failed; continuing...", err)
				continue
			}

			image, err := GetImage(fmt.Sprintf("%s=w4000", url))
			if err != nil {
				log.Println("GetImage failed; continuing...", err)
				continue
			}

			select {
			case cache <- CachedImage{item, image}:
				// nothing
			case <-ctx.Done():
				return
			}
		}
	}()

	return cache
}

func GetImage(url string) ([]byte, error) {
	client := http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to download image", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Received non-200 response while fetching image", resp.StatusCode)
		return nil, err
	}

	image, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read image response", err)
		return nil, err
	}

	return image, nil
}

// Can't rely on initial `baseUrl` because it only lasts 60 minutes
func GetImageUrl(client *http.Client, item MediaItem) (string, error) {
	url := fmt.Sprintf("https://photoslibrary.googleapis.com/v1/mediaItems/%s", item.Id)

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to fetch metadata", err)
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Received non-200 response while fetching metadata", resp.StatusCode)
		return "", fmt.Errorf("non-200 response while fetching metadata: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read metadata response", err)
		return "", err
	}

	r := &MediaItem{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println("Failed to parse metadata response", err)
		return "", err
	}

	return r.BaseUrl, nil
}
