package webflowAPI

// Tell `go generate` to generate the mock for us.
//go:generate moq -pkg mock -out mock/webflowAPI_moq.go . Interface

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sethgrid/pester"
	"github.com/tidwall/gjson"
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

// Interface Interface for this package's method. Created primarily for testing your code that depends on this package.
type Interface interface {
	MethodGet(uri string, queryParams map[string]string, decodedResponse interface{}) error
	GetAllCollections() (*Collections, error)
	GetCollectionByName(name string) (*Collection, error)
	GetCollectionBySlug(slug string) (*Collection, error)
	GetAllItemsInCollectionByID(ID string, maxPages int) ([][]byte, error)
	GetAllItemsInCollectionByName(name string, maxPages int) ([][]byte, error)
	GetAllItemsInCollectionBySlug(slug string, maxPages int) ([][]byte, error)
	GetItem(cName, cSlug, cID, iName, iID string) ([]byte, error)
}

// apiConfig Represents a configuration struct for Webflow apiConfig object.
type apiConfig struct {
	Client                          *pester.Client
	Token, Version, BaseURL, SiteID string
	// The following methods are overrides for the public methods. Use only for internal testing of the pkg.
	methodGet                     func(uri string, queryParams map[string]string, decodedResponse interface{}) error
	getAllCollections             func() (*Collections, error)
	getCollectionByName           func(name string) (*Collection, error)
	getCollectionBySlug           func(slug string) (*Collection, error)
	getAllItemsInCollectionByID   func(ID string, maxPages int) ([][]byte, error)
	getAllItemsInCollectionByName func(name string, maxPages int) ([][]byte, error)
	getAllItemsInCollectionBySlug func(slug string, maxPages int) ([][]byte, error)
	getItem                       func(cName, cSlug, cID, iName, iID string) ([]byte, error)
}

// New Create a new configuration struct for the Webflow API object.
func New(token, siteID string, hc *http.Client) *apiConfig {
	var client *pester.Client

	// If a http client is passed in, use it.
	if hc != nil {
		client = pester.NewExtendedClient(hc)
	} else {
		client = pester.New()
	}

	// client.Concurrency = 3
	client.MaxRetries = 10
	client.Backoff = pester.ExponentialBackoff
	client.KeepLog = true
	client.RetryOnHTTP429 = true

	return &apiConfig{
		Client:  client,
		Token:   token,
		Version: defaultVersion,
		BaseURL: defaultURL,
		SiteID:  siteID,
	}
}

// MethodGet Execute a HTTP GET on the specified URI.
func (api *apiConfig) MethodGet(uri string, queryParams map[string]string, decodedResponse interface{}) error {
	// If an override was configured, use it instead.
	if api.methodGet != nil {
		return api.methodGet(uri, queryParams, decodedResponse)
	}

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
	// TODO: read docs for ReaderCloser.Close() to determine what to do when it errors.
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
func (api *apiConfig) GetAllCollections() (*Collections, error) {
	// If an override was configured, use it instead.
	if api.getAllCollections != nil {
		return api.getAllCollections()
	}

	collections := &Collections{}
	err := api.MethodGet(fmt.Sprintf(listCollectionsURL, api.SiteID), nil, collections)

	if err != nil {
		return nil, err
	}

	return collections, nil
}

// GetCollectionByName Query Webflow for all the collections then search them for the requested name, case insensitive.
func (api *apiConfig) GetCollectionByName(name string) (*Collection, error) {
	// If an override was configured, use it instead.
	if api.getCollectionByName != nil {
		return api.getCollectionByName(name)
	}

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

// GetCollectionBySlug Query Webflow for all the collections then search them for the requested slug, case insensitive.
func (api *apiConfig) GetCollectionBySlug(slug string) (*Collection, error) {
	// If an override was configured, use it instead.
	if api.getCollectionBySlug != nil {
		return api.getCollectionBySlug(slug)
	}

	collections, err := api.GetAllCollections()
	if err != nil {
		return nil, err
	}

	lowerSlug := strings.ToLower(slug)

	for _, collection := range *collections {
		if strings.ToLower(collection.Slug) == lowerSlug {
			return &collection, nil
		}
	}

	// Report that no collection was found by that slug.
	return nil, nil
}

// GetAllItemsInCollectionByID Ask the Webflow API for all the items in a given collection, by the collection's ID.
func (api *apiConfig) GetAllItemsInCollectionByID(id string, maxPages int) ([][]byte, error) {
	// If an override was configured, use it instead.
	if api.getAllItemsInCollectionByID != nil {
		return api.getAllItemsInCollectionByID(id, maxPages)
	}

	offset := 0
	items := [][]byte{}

	for {
		queryParams := map[string]string{
			"offset": strconv.Itoa(offset),
			"limit":  "100",
		}

		collectionItems := &CollectionItems{}
		err := api.MethodGet(fmt.Sprintf(listCollectionItemsURL, id), queryParams, collectionItems)
		if err != nil {
			return nil, err
		}

		// Iteratae over all the collection's items.
		jsonItems := gjson.Parse(string(collectionItems.Items))
		jsonItems.ForEach(func(key, value gjson.Result) bool {
			// Add each json item to the slice.
			items = append(items, []byte(value.Raw))
			// Keep iterating.
			return true
		})

		offset = collectionItems.Offset + collectionItems.Count

		// Webflow API should report when the last set of items has been requested. Once this has happened, this loop should
		// be broken.
		if offset >= collectionItems.Total {
			break
		}

		// Safety feature to keep the code from infinite looping or asking the API for far too many items.
		if maxPages--; maxPages < 0 {
			break
		}
	}

	return items, nil
}

// GetAllItemsInCollectionByName Ask the Webflow API for all the items in a given collection, by the collection's name.
// The collection name will be searched with case insensitivity.
func (api *apiConfig) GetAllItemsInCollectionByName(name string, maxPages int) ([][]byte, error) {
	// If an override was configured, use it instead.
	if api.getAllItemsInCollectionByName != nil {
		return api.getAllItemsInCollectionByName(name, maxPages)
	}

	// Find the collection by name.
	collection, err := api.GetCollectionByName(name)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, nil
	}

	// Now find the items by the collection's ID.
	return api.GetAllItemsInCollectionByID(collection.ID, maxPages)
}

// GetAllItemsInCollectionBySlug Ask the Webflow API for all the items in a given collection, by the collection's slug.
// The collection slug will be searched with case insensitivity.
func (api *apiConfig) GetAllItemsInCollectionBySlug(slug string, maxPages int) ([][]byte, error) {
	// If an override was configured, use it instead.
	if api.getAllItemsInCollectionBySlug != nil {
		return api.getAllItemsInCollectionBySlug(slug, maxPages)
	}

	// Find the collection by slug.
	collection, err := api.GetCollectionBySlug(slug)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, nil
	}

	// Now find the items by the collection's ID.
	return api.GetAllItemsInCollectionByID(collection.ID, maxPages)
}

// GetItem Searches all the items in a given collection for the desired item name or ID.
// cName Case insensitive search for collection by name. Not necessary if `cSlug` or `cID` is provided.
// cSlug Case insensitive search for collection by slug. Not necessary if `cName` or `cID` is provided.
// cID ID of collection to find. Not necessary if `cName` is provided.
// iName Case insensitive search for item by name. Not necessary if `iID` is provided.
// iID ID of item to find. Not necessary if `iName` is provided.
func (api *apiConfig) GetItem(cName, cSlug, cID, iName, iID string) ([]byte, error) {
	// If an override was configured, use it instead.
	if api.getItem != nil {
		return api.getItem(cName, cSlug, cID, iName, iID)
	}

	var items [][]byte
	var err error

	// Just quietly return nothing since a collection name & slug & ID were not provided.
	if cName == "" && cSlug == "" && cID == "" {
		return nil, nil
	}

	// Just quietly return nothing since neither an item name nor an ID was provided.
	if iName == "" && iID == "" {
		return nil, nil
	}

	if cName != "" {
		items, err = api.GetAllItemsInCollectionByName(cName, 10)
		if err != nil {
			return nil, fmt.Errorf("unable to get all items in collection by collection name; error: %+v", err)
		}
	} else if cSlug != "" {
		items, err = api.GetAllItemsInCollectionBySlug(cSlug, 10)
		if err != nil {
			return nil, fmt.Errorf("unable to get all items in collection by collection slug; error: %+v", err)
		}
	} else {
		items, err = api.GetAllItemsInCollectionByID(cID, 10)
		if err != nil {
			return nil, fmt.Errorf("unable to get all items in collection by collection ID; error: %+v", err)
		}
	}

	for _, rawItem := range items {
		item := &CollectionItem{}
		if err2 := json.Unmarshal(rawItem, item); err2 != nil {
			return nil, fmt.Errorf(
				"GetItem() did not receive the proper collection item type: %+v",
				err2,
			)
		}

		if iName != "" && item.Name != iName {
			continue
		}

		if iID != "" && item.ID != iID {
			continue
		}

		return rawItem, nil
	}

	return nil, nil
}
