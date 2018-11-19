package webflowAPI

import (
	"encoding/json"
	"time"
)

// GeneralError API response struct when there is an error.
type GeneralError struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Name string `json:"name"`
	Path string `json:"path"`
	Err  string `json:"err"`
}

// Collection API contract for CMS collection.
type Collection struct {
	ID           string    `json:"_id"`
	LastUpdated  time.Time `json:"lastUpdated"`
	CreatedOn    time.Time `json:"createdOn"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	SingularName string    `json:"singularName"`
}

// Collections List of Collection.
type Collections []Collection

// CollectionItem API contract for item(s) in a given collection.
type CollectionItem struct {
	Archived    bool   `json:"_archived"`
	Draft       bool   `json:"_draft"`
	Name        string `json:"name"`
	PostBody    string `json:"post-body"`
	PostSummary string `json:"post-summary"`
	Slug        string `json:"slug"`
	Author      string `json:"author"`
	Cid         string `json:"_cid"`
	ID          string `json:"_id"`
}

// CollectionItems API contract for retrieving collection items.
type CollectionItems struct {
	// Delay parsing until we know the type.
	Items  json.RawMessage `json:"items"`
	Count  int             `json:"count"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Total  int             `json:"total"`
}
