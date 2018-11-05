package webflow

import (
	"time"
)

type GeneralError struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Name string `json:"name"`
	Path string `json:"path"`
	Err  string `json:"err"`
}

type Collection struct {
	ID           string    `json:"_id"`
	LastUpdated  time.Time `json:"lastUpdated"`
	CreatedOn    time.Time `json:"createdOn"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	SingularName string    `json:"singularName"`
}

type Collections []Collection

type CollectionItem struct {
	Archived       bool   `json:"_archived"`
	Draft          bool   `json:"_draft"`
	Color          string `json:"color"`
	Featured       bool   `json:"featured"`
	Name           string `json:"name"`
	PostBody       string `json:"post-body"`
	PostSummary    string `json:"post-summary"`
	ThumbnailImage struct {
		FileID string `json:"fileId"`
		URL    string `json:"url"`
	} `json:"thumbnail-image"`
	MainImage struct {
		FileID string `json:"fileId"`
		URL    string `json:"url"`
	} `json:"main-image"`
	Slug        string    `json:"slug"`
	UpdatedOn   time.Time `json:"updated-on"`
	UpdatedBy   string    `json:"updated-by"`
	CreatedOn   time.Time `json:"created-on"`
	CreatedBy   string    `json:"created-by"`
	PublishedOn time.Time `json:"published-on"`
	PublishedBy string    `json:"published-by"`
	Author      string    `json:"author"`
	Cid         string    `json:"_cid"`
	ID          string    `json:"_id"`
}

type CollectionItems struct {
	Items  []CollectionItem `json:"items"`
	Count  int              `json:"count"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
	Total  int              `json:"total"`
}
