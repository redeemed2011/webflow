package webflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sethgrid/pester"
)

const (
	defaultURL     = "https://api.webflow.com"
	defaultVersion = "1.0.0"

	// List Collections.
	// http://developers.webflow.com/?shell#list-collections
	listCollectionsURL = "/sites/%s/collections"

	// Get All Items For a Collection.
	// http://developers.webflow.com/?shell#get-all-items-for-a-collection
	listCollectionItemsURL = "/collections/%s/items"
)

// API Represents a configuration struct for Webflow API object.
type API struct {
	// Client *http.Client
	Client                          *pester.Client
	Token, Version, BaseURL, SiteID string
}

// New Create a new configuration struct for the Webflow API object.
func New(token, siteID string) *API {
	client := pester.New()

	// client.Concurrency = 3
	client.MaxRetries = 5
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = true
	client.RetryOnHTTP429 = true

	return &API{
		// Client: &http.Client{},
		Client: client,
		// Client:  pester.New(),
		Token:   token,
		Version: defaultVersion,
		BaseURL: defaultURL,
		SiteID:  siteID,
	}
}

// MethodGet Execute a HTTP GET on the specified URI.
func (api *API) MethodGet(uri string, queryParams map[string]string, decodedResponse interface{}) error {
	// Form the request to make to WebFlow.
	req, err := http.NewRequest("GET", api.BaseURL+uri, nil)
	if err != nil {
		return errors.New(fmt.Sprint("Unable to create a new http request", err))
	}

	// Webflow needs to know the auth token and the version of their API to use.
	req.Header.Set("Authorization", "Bearer "+api.Token)
	req.Header.Set("Accept-Version", defaultVersion)

	// Set query parameters.
	if len(queryParams) > 0 {
		query := req.URL.Query()
		for key, val := range queryParams {
			query.Add(key, val)
		}
		req.URL.RawQuery = query.Encode()
	}

	// Make the request.
	res, err := api.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Status codes of 200 to 299 are healthy; the rest are an error, redirect, etc.
	if res.StatusCode >= 300 || res.StatusCode < 200 {
		errResp := &GeneralError{}
		if err := json.NewDecoder(res.Body).Decode(errResp); err != nil {
			return fmt.Errorf("Unknown API error; status code %d; error: %+v", res.StatusCode, err)
		}
		return errors.New(errResp.Err)
	}

	if err := json.NewDecoder(res.Body).Decode(decodedResponse); err != nil {
		return err
	}

	return nil
}

// GetAllCollections Ask the Webflow API for all the collections on a given site.
func (api *API) GetAllCollections() (*Collections, error) {
	collections := &Collections{}
	err := api.MethodGet(fmt.Sprintf(listCollectionsURL, api.SiteID), nil, collections)

	if err != nil {
		return nil, err
	}

	return collections, nil
}

// GetCollectionByName Query Webflow for all the collections then search them for the requested name, case insensitive.
func (api *API) GetCollectionByName(name string) (*Collection, error) {
	collections, err := api.GetAllCollections()
	if err != nil {
		return nil, err
	}

	lowerName := strings.ToLower(name)

	for _, collection := range *collections {
		if strings.ToLower(collection.Name) == lowerName {
			return &collection, nil
		}
	}

	// Report that no collection was found by that name.
	return nil, nil
}

// GetAllItemsInCollectionID Ask the Webflow API for all the items in a given collection, by the collection's ID.
func (api *API) GetAllItemsInCollectionID(collectionID string) (*[]CollectionItem, error) {
	items := []CollectionItem{}
	offset := 0
	safety := 10

	for {
		queryParams := map[string]string{
			"offset": strconv.Itoa(offset),
			"limit":  "100",
		}

		apiItems := &CollectionItems{}
		err := api.MethodGet(fmt.Sprintf(listCollectionItemsURL, collectionID), queryParams, apiItems)
		if err != nil {
			return nil, err
		}

		items = append(items, apiItems.Items...)

		offset = apiItems.Offset + apiItems.Count

		// Webflow API should report when the last set of items has been requested. Once this has happened, this loop should
		// be broken.
		if offset >= apiItems.Total {
			break
		}

		if safety--; safety < 0 {
			break
		}
	}

	return &items, nil
}

// GetAllItemsInCollectionName Ask the Webflow API for all the items in a given collection, by the collection's name.
// The collection name will be searched with case insensitivity.
func (api *API) GetAllItemsInCollectionName(collectionName string) (*[]CollectionItem, error) {
	// Find the collection by name.
	collection, err := api.GetCollectionByName(collectionName)
	if err != nil {
		return nil, err
	}

	// Now find the items by the collection's ID.
	return api.GetAllItemsInCollectionID(collection.ID)
}
