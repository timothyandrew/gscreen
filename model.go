package main

import "time"

type MediaMetadata struct {
	CreationTime time.Time `json:"creationTime"`
}

type MediaItem struct {
	Id       string        `json:"id"`
	BaseUrl  string        `json:"baseUrl"`
	MimeType string        `json:"mimeType"`
	Metadata MediaMetadata `json:"mediaMetadata"`
	Filename string        `json:"filename"`
}

type ListMediaItemsResponse struct {
	MediaItems    []MediaItem `json:"mediaItems"`
	NextPageToken string      `json:"nextPageToken"`
}
