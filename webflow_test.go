package webflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const (
	siteID = "mysiteid"
)

var (
	exampleDogCollection = &Collection{
		ID:   "1",
		Name: "dogs",
	}
	exampleCatCollection = &Collection{
		ID:   "2",
		Name: "cats",
	}
	exampleCollections = &Collections{
		*exampleDogCollection,
		*exampleCatCollection,
	}
	exampleItemDog1 = &CollectionItem{
		ID:   "1",
		Name: "blue",
	}
	exampleItemDog2 = &CollectionItem{
		ID:   "2",
		Name: "blue",
	}
	exampleItemDog3 = &CollectionItem{
		ID:   "2",
		Name: "blue",
	}
	exampleItemDog4 = &CollectionItem{
		ID:   "2",
		Name: "blue",
	}
	apiResponseItemsDogs1 = &CollectionItems{
		Items: []CollectionItem{
			*exampleItemDog1,
			*exampleItemDog2,
		},
		Offset: 0,
		Count:  2,
		Total:  4,
	}
	apiResponseItemsDogs2 = &CollectionItems{
		Items: []CollectionItem{
			*exampleItemDog3,
			*exampleItemDog4,
		},
		Offset: 2,
		Count:  2,
		Total:  4,
	}
	// Must be the combination of items from apiResponseItemsDogs1 & apiResponseItemsDogs2.
	exampleItemsDogs = &[]CollectionItem{
		*exampleItemDog1,
		*exampleItemDog2,
		*exampleItemDog3,
		*exampleItemDog4,
	}
	apiResponseItemsDogs = &CollectionItems{
		Items:  *exampleItemsDogs,
		Offset: 0,
		Count:  4,
		Total:  4,
	}
)

func TestApiGetUnknownResponseFormat(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`Unexpected`))
	}))
	defer server.Close()

	res := &CollectionItem{}
	api := New("mytoken", "mysiteid")
	api.BaseURL = server.URL
	err := api.MethodGet("/", nil, res)
	if err == nil {
		t.Error("API Get() is expected to return an error when receiving an unknown response format.")
	}

	if !strings.Contains(err.Error(), "invalid character") {
		t.Error("API Get() should error stating the unknown response format has an invalid character.")
	}
}

func TestApiGetErrorResponseFormat(t *testing.T) {
	// Keep track of number of times the API is touched.
	tries := 0
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Count this attempt.
		tries++
		// Simple error to return.
		errResp := &GeneralError{
			Code: http.StatusTooManyRequests,
			Err:  "rate limiting you!",
		}
		data, _ := json.Marshal(errResp)
		rw.WriteHeader(http.StatusTooManyRequests)
		rw.Write(data)
	}))
	defer server.Close()

	res := &CollectionItem{}
	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	// Setup a backoff that is always just 1 millisecond.
	api.Client.Backoff = func(retry int) time.Duration {
		return 1 * time.Millisecond
	}
	err := api.MethodGet("/", nil, res)

	if err == nil {
		t.Error("API Get() is expected to return an error when an API error response is received.")
	}

	if tries < 2 {
		t.Errorf("API Get() is expected to retry when rate limiting errors are encountered! It tried the request %d times.", tries)
	}
}

func TestApiGet(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// fmt.Printf("%+v", req.URL.Query())
		reqURI := req.URL.String()
		expectedURI := "/?limit=1&offset=2"
		if reqURI != expectedURI {
			t.Errorf(
				"MethodGet() did not request the proper URI! requested '%s'; expected '%s'.",
				reqURI,
				expectedURI,
			)
		}
		data, _ := json.Marshal(exampleItemDog1)
		rw.Write(data)
	}))
	defer server.Close()

	res := &CollectionItem{}
	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	query := map[string]string{
		"offset": "2",
		"limit":  "1",
	}
	err := api.MethodGet("/", query, res)

	if err != nil {
		t.Error("API MethodGet() is expected to return no error when no error is encountered.")
	}

	if !reflect.DeepEqual(exampleItemDog1, res) {
		t.Errorf("API MethodGet() did not return the expected values! Got %+v.", res)
	}
}

func TestApiGetAllCollections(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		expectedURI := fmt.Sprintf(listCollectionsURL, siteID)
		if reqURI != expectedURI {
			t.Errorf(
				"GetAllCollections() did not request the proper URI! requested '%s'; expected '%s'.",
				reqURI,
				expectedURI,
			)
		}
		data, _ := json.Marshal(exampleCollections)
		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	res, err := api.GetAllCollections()

	if err != nil {
		t.Error("API GetAllCollections() is expected to return no func error when receiving a properly formatted response.")
	}

	if len(*res) != len(*exampleCollections) {
		t.Errorf("API GetAllCollections() is expected to return %d collections! Got %d.", len(*exampleCollections), len(*res))
	}
}

func TestApiGetAllCollectionsUnknownError(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusRequestTimeout)
		rw.Write([]byte{})
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	_, err := api.GetAllCollections()

	if err == nil {
		t.Error("API GetAllCollections() is expected to return a func error when receiving an unknown error.")
	}
}

func TestApiGetAllCollectionsErrorResponse(t *testing.T) {
	errMsg := "item not found!"

	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		errResp := &GeneralError{
			Code: http.StatusNotFound,
			Err:  errMsg,
		}
		data, _ := json.Marshal(errResp)
		rw.WriteHeader(http.StatusNotFound)
		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	_, err := api.GetAllCollections()

	if err == nil {
		t.Error("API GetAllCollections() is expected to return a func error when receiving an unknown error.")
	}

	if err.Error() != errMsg {
		t.Errorf("API GetAllCollections() returned an incorrect error message! Got '%s'; expected '%s'.", err.Error(), errMsg)
	}
}

func TestApiGetCollectionByName(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		data, _ := json.Marshal(exampleCollections)
		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL

	// Test searching for a collection that exists.
	{
		collection, err := api.GetCollectionByName("dogs")

		if err != nil {
			t.Error("API GetCollectionByName() is expected to return no func error when receiving a properly formatted response.")
		}

		if !reflect.DeepEqual(collection, exampleDogCollection) {
			t.Errorf("API GetCollectionByName() is expected to return exampleDogCollection! Got %+v.", collection)
		}
	}

	// Test searching for a collection that does not exist.
	{
		collection, _ := api.GetCollectionByName("birds")

		if collection != nil {
			t.Errorf("API GetCollectionByName() is expected to return nil whenever the collection is not found! Got %+v.", collection)
		}
	}
}

func TestApiGetAllItemsInCollectionID(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		expectedURI := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)
		if !strings.HasPrefix(reqURI, expectedURI) {
			t.Errorf(
				"GetAllItemsInCollectionID() did not request the proper URI! requested '%s'; expected '%s'.",
				reqURI,
				expectedURI,
			)
		}

		// Look at the query params. Return one set of results if the "offset" query param is set to 2; and another set of
		// results otherwise.
		queryParams := req.URL.Query()
		if queryParams.Get("offset") == "2" {
			data, _ := json.Marshal(apiResponseItemsDogs2)
			rw.Write(data)
			return
		}
		data, _ := json.Marshal(apiResponseItemsDogs1)
		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	res, err := api.GetAllItemsInCollectionID(exampleDogCollection.ID)

	if err != nil {
		t.Error("API GetAllItemsInCollectionID() is expected to return no error when receiving a properly formatted response.")
	}

	if !reflect.DeepEqual(res, exampleItemsDogs) {
		t.Errorf("API GetAllItemsInCollectionID() is expected to return exampleItemsDogs! Got %+v.", res)
	}
}

func TestApiGetAllItemsInCollectionName(t *testing.T) {
	collectionURI := fmt.Sprintf(listCollectionsURL, siteID)
	itemsURI := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)

	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		switch {
		case strings.HasPrefix(reqURI, collectionURI):
			data, _ := json.Marshal(exampleCollections)
			rw.Write(data)
		case strings.HasPrefix(reqURI, itemsURI):
			data, _ := json.Marshal(apiResponseItemsDogs)
			rw.Write(data)
		default:
			t.Errorf("GetAllItemsInCollectionName() did not request the proper URI! requested '%s'.", reqURI)
		}
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	res, err := api.GetAllItemsInCollectionName(exampleDogCollection.Name)

	if err != nil {
		t.Error("API GetAllItemsInCollectionName() is expected to return no error when receiving a properly formatted response.")
	}

	if !reflect.DeepEqual(res, exampleItemsDogs) {
		t.Errorf("API GetAllItemsInCollectionName() is expected to return exampleItemsDogs! Got %+v.", res)
	}
}
