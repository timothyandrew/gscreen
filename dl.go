package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Can't rely on initial `baseUrl` because it only lasts 60 minutes
func GetImageUrl(client *http.Client, item MediaItem) (string, error) {
	url := fmt.Sprintf("https://photoslibrary.googleapis.com/v1/mediaItems/%s", item.Id)

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to fetch metadata", err)
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Println("Received non-200 response while fetching metadata", resp.StatusCode)
		return "", err
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
