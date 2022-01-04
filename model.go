package main

import (
	"strings"
	"time"
)

type VideoMetadata struct {
	Status string `json:"status"`
}

type MediaMetadata struct {
	CreationTime time.Time     `json:"creationTime"`
	Video        VideoMetadata `json:"video"`
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

type Status struct {
	Code    int    `json:"code"`
	Message string `json:"string"`
}

type MediaItemResult struct {
	Status    Status    `json:"status"`
	MediaItem MediaItem `json:"mediaItem"`
}

type BatchGetMediaItemsResponse struct {
	MediaItemResults []MediaItemResult `json:"mediaItemResults"`
}

func (m *MediaItem) isImage() bool {
	return strings.Contains(m.MimeType, "image")
}

func (m *MediaItem) isVideo() bool {
	return strings.Contains(m.MimeType, "video") && m.Metadata.Video.Status == "READY"
}
