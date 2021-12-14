package main

type MediaItem struct {
	Id       string `json:id`
	BaseUrl  string `json:baseUrl`
	MimeType string `json:mimeType`
	Filename string `json:filename`
}

type ListMediaItemsResponse struct {
	MediaItems    []MediaItem `json:mediaItems`
	NextPageToken string      `json:nextPageToken`
}
