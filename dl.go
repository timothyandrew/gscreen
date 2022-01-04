package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type CachedImage struct {
	mediaItem MediaItem
	data      []byte
}

type DownloadCache = chan CachedImage

// Can't go over 50 here: https://developers.google.com/photos/library/reference/rest/v1/mediaItems/batchGet
const batchSize = 50

func InitializeDownloadCache(ctx context.Context, client *http.Client, metadata *MetadataCache) DownloadCache {
	cache := make(chan CachedImage, batchSize)

	go func() {
		for {
			items, err := metadata.Random(batchSize)
			if err != nil {
				log.Println("Image metadata hasn't loaded yet, going to wait...", err)
				time.Sleep(2 * time.Second)
				continue
			}

			items, err = RefreshBaseUrls(client, items)
			if err != nil {
				log.Println("Failed to get base URLs; continuing...", err)
				continue
			}

			imageChan := make(chan CachedImage, batchSize)
			var wg sync.WaitGroup

			for _, item := range items {
				wg.Add(1)

				// TODO: Worker pool to avoid 50 downloads at once
				go func(out chan CachedImage, item MediaItem) {
					defer wg.Done()

					url := fmt.Sprintf("%s=w4000", item.BaseUrl)
					if item.isVideo() {
						url = fmt.Sprintf("%s=dv", item.BaseUrl)
					}

					image, err := GetMediaBytes(url)
					if err != nil {
						log.Println("GetImage failed; continuing...", err)
						return
					}

					out <- CachedImage{item, image}
				}(imageChan, item)
			}

			go func() {
				wg.Wait()
				close(imageChan)
			}()

			for image := range imageChan {
				select {
				case cache <- image:
					// nothing
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return cache
}

func GetMediaBytes(url string) ([]byte, error) {
	client := http.Client{Timeout: 30 * time.Second}

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
func RefreshBaseUrls(client *http.Client, items []MediaItem) (urls []MediaItem, err error) {
	url := "https://photoslibrary.googleapis.com/v1/mediaItems:batchGet?mediaItemIds="
	url += items[0].Id

	for i := 1; i < len(items); i++ {
		url += fmt.Sprintf("&mediaItemIds=%s", items[i].Id)
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to fetch metadata", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("Received non-200 response while fetching metadata", resp.StatusCode)
		return nil, fmt.Errorf("non-200 response while fetching metadata: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Failed to read metadata response", err)
		return nil, err
	}

	r := &BatchGetMediaItemsResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println("Failed to parse metadata response", err)
		return nil, err
	}

	for _, v := range r.MediaItemResults {
		// https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
		if v.Status.Code == 0 {
			urls = append(urls, v.MediaItem)
		} else {
			log.Printf("BatchGet failed on Id %d with error code %d and error %s\n", v.Status.Code, v.Status.Message)
		}
	}

	return
}
