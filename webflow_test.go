package webflowAPI

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

type mockItem struct {
	Name      string `json:"name"`
	Body      string `json:"body"`
	Something string `json:"something"`
	ID        string `json:"_id"`
}

var (
	exampleDogCollection = &Collection{
		ID:   "1",
		Name: "dogs",
	}
	exampleItemDog1 = &mockItem{
		ID:   "1",
		Name: "blue",
	}
	exampleItemDog2 = &mockItem{
		ID:   "2",
		Name: "blue",
	}
	exampleItemDog3 = &mockItem{
		ID:   "2",
		Name: "blue",
	}
	exampleItemDog4 = &mockItem{
		ID:   "2",
		Name: "blue",
	}
	apiResponseItemsDogs1SubItemsJSON, _ = json.Marshal([]mockItem{
		*exampleItemDog1,
		*exampleItemDog2,
	})
	apiResponseItemsDogs1 = &CollectionItems{
		Items:  apiResponseItemsDogs1SubItemsJSON,
		Offset: 0,
		Count:  2,
		Total:  4,
	}
	apiResponseItemsDogs2SubItemsJSON, _ = json.Marshal([]mockItem{
		*exampleItemDog3,
		*exampleItemDog4,
	})
	apiResponseItemsDogs2 = &CollectionItems{
		Items:  apiResponseItemsDogs2SubItemsJSON,
		Offset: 2,
		Count:  2,
		Total:  4,
	}
	// Must be the combination of items from apiResponseItemsDogs1 & apiResponseItemsDogs2.
	exampleItemsDogs = &[]mockItem{
		*exampleItemDog1,
		*exampleItemDog2,
		*exampleItemDog3,
		*exampleItemDog4,
	}
	exampleItemsDogsJSON, _ = json.Marshal(*exampleItemsDogs)
	apiResponseItemsDogs    = &CollectionItems{
		Items:  exampleItemsDogsJSON,
		Offset: 0,
		Count:  4,
		Total:  4,
	}
	exampleItemCat1 = &mockItem{
		ID:   "11",
		Name: "blue",
	}
	exampleItemCat2 = &mockItem{
		ID:   "12",
		Name: "blue",
	}
	exampleItemCat3 = &mockItem{
		ID:   "12",
		Name: "blue",
	}
	exampleItemCat4 = &mockItem{
		ID:   "12",
		Name: "blue",
	}
	apiResponseItemsCats1SubItemsJSON, _ = json.Marshal([]mockItem{
		*exampleItemCat1,
		*exampleItemCat2,
	})
	apiResponseItemsCats1 = &CollectionItems{
		Items:  apiResponseItemsCats1SubItemsJSON,
		Offset: 0,
		Count:  2,
		Total:  4,
	}
	apiResponseItemsCats2SubItemsJSON, _ = json.Marshal([]mockItem{
		*exampleItemCat3,
		*exampleItemCat4,
	})
	apiResponseItemsCats2 = &CollectionItems{
		Items:  apiResponseItemsCats2SubItemsJSON,
		Offset: 2,
		Count:  2,
		Total:  4,
	}
	// Must be the combination of items from apiResponseItemsCats1 & apiResponseItemsCats2.
	exampleItemsCats = &[]mockItem{
		*exampleItemCat1,
		*exampleItemCat2,
		*exampleItemCat3,
		*exampleItemCat4,
	}
	exampleItemsCatsJSON, _ = json.Marshal(*exampleItemsCats)
	apiResponseItemsCats    = &CollectionItems{
		Items:  exampleItemsCatsJSON,
		Offset: 0,
		Count:  4,
		Total:  4,
	}
	exampleCatCollection = &Collection{
		ID:   "2",
		Name: "cats",
	}
	exampleCollections = &Collections{
		*exampleDogCollection,
		*exampleCatCollection,
	}
)

func TestApiGetUnknownResponseFormat(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`Unexpected`))
	}))
	defer server.Close()

	res := &mockItem{}
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

	res := &mockItem{}
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

	res := &mockItem{}
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

// func TestApiGetAll(t *testing.T) {
// 	// Start a special, local HTTP server.
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		// Look at the query params. Return one set of results if the "offset" query param is set to 2; and another set of
// 		// results otherwise.
// 		queryParams := req.URL.Query()
// 		if queryParams.Get("offset") == "2" {
// 			data, _ := json.Marshal(apiResponseItemsDogs2)
// 			rw.Write(data)
// 			return
// 		}
// 		data, _ := json.Marshal(apiResponseItemsDogs1)
// 		rw.Write(data)
// 	}))
// 	defer server.Close()

// 	api := New("mytoken", siteID)
// 	api.BaseURL = server.URL
// 	res, err := api.GetAllItemsInCollectionID(exampleDogCollection.ID)

// 	if err != nil {
// 		t.Error("API GetAllItemsInCollectionID() is expected to return no error when receiving a properly formatted response.")
// 	}

// 	if !reflect.DeepEqual(res, exampleItemsDogs) {
// 		t.Errorf("API GetAllItemsInCollectionID() is expected to return exampleItemsDogs! Got %+v.", res)
// 	}
// }

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

// TestApiGetAllItemsInCollectionID1 Test the GetAllItemsInCollectionID func for one collection type (dogs).
func TestApiGetAllItemsInCollectionID1(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		queryParams := req.URL.Query()
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
	items := []mockItem{}
	err := api.GetAllItemsInCollectionID(exampleDogCollection.ID, 10, func(jsonItems json.RawMessage) error {
		tempItems := &[]mockItem{}
		if err2 := json.Unmarshal(jsonItems, tempItems); err2 != nil {
			return fmt.Errorf("API GetAllItemsInCollectionID() did not return the proper collection items types. Error %+v", err2)
		}
		items = append(items, *tempItems...)
		return nil
	})

	if err != nil {
		t.Errorf("API GetAllItemsInCollectionID() is expected to return no error when receiving a properly formatted response. got: %+v", err)
	}

	if !reflect.DeepEqual(items, *exampleItemsDogs) {
		t.Errorf("API GetAllItemsInCollectionID() is expected to return exampleItemsDogs! Got %+v.", items)
	}
}

// TestApiGetAllItemsInCollectionID2 Test the GetAllItemsInCollectionID func for another collection type (cats).
func TestApiGetAllItemsInCollectionID2(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		queryParams := req.URL.Query()
		expectedURI := fmt.Sprintf(listCollectionItemsURL, exampleCatCollection.ID)
		if !strings.HasPrefix(reqURI, expectedURI) {
			t.Errorf(
				"GetAllItemsInCollectionID() did not request the proper URI! requested '%s'; expected '%s'.",
				reqURI,
				expectedURI,
			)
		}

		// Look at the query params. Return one set of results if the "offset" query param is set to 2; and another set of
		// results otherwise.
		if queryParams.Get("offset") == "2" {
			data, _ := json.Marshal(apiResponseItemsCats2)
			rw.Write(data)
			return
		}
		data, _ := json.Marshal(apiResponseItemsCats1)
		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID)
	api.BaseURL = server.URL
	items := []mockItem{}
	err := api.GetAllItemsInCollectionID(exampleCatCollection.ID, 10, func(jsonItems json.RawMessage) error {
		tempItems := &[]mockItem{}
		if err2 := json.Unmarshal(jsonItems, tempItems); err2 != nil {
			return fmt.Errorf("API GetAllItemsInCollectionID() did not return the proper collection items types. Error %+v", err2)
		}
		items = append(items, *tempItems...)
		return nil
	})

	if err != nil {
		t.Errorf("API GetAllItemsInCollectionID() is expected to return no error when receiving a properly formatted response. got: %+v", err)
	}

	if !reflect.DeepEqual(items, *exampleItemsCats) {
		t.Errorf("API GetAllItemsInCollectionID() is expected to return exampleItemsCats! Got %+v.", items)
	}
}

// func TestApiGetAllItemsInCollectionName(t *testing.T) {
// 	collectionURI := fmt.Sprintf(listCollectionsURL, siteID)
// 	itemsURI := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)

// 	// Start a special, local HTTP server.
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		reqURI := req.URL.String()
// 		switch {
// 		case strings.HasPrefix(reqURI, collectionURI):
// 			data, _ := json.Marshal(exampleCollections)
// 			rw.Write(data)
// 		case strings.HasPrefix(reqURI, itemsURI):
// 			data, _ := json.Marshal(apiResponseItemsDogs)
// 			rw.Write(data)
// 		default:
// 			t.Errorf("GetAllItemsInCollectionName() did not request the proper URI! requested '%s'.", reqURI)
// 		}
// 	}))
// 	defer server.Close()

// 	api := New("mytoken", siteID)
// 	api.BaseURL = server.URL
// 	jsonItems, err := api.GetAllItemsInCollectionName(exampleDogCollection.Name)

// 	if err != nil {
// 		t.Error("API GetAllItemsInCollectionName() is expected to return no error when receiving a properly formatted response.")
// 	}

// 	items := []mockItem{}
// 	for _, jsonItem := range jsonItems {
// 		tempItems := &[]mockItem{}
// 		err2 := json.Unmarshal(jsonItem, tempItems)
// 		if err2 != nil {
// 			t.Errorf("API GetAllItemsInCollectionName() did not return the proper collection items types. Error %+v", err2)
// 		}
// 		items = append(items, *tempItems...)
// 	}

// 	if !reflect.DeepEqual(items, *exampleItemsDogs) {
// 		t.Errorf("API GetAllItemsInCollectionName() is expected to return exampleItemsDogs! Got %+v.", items)
// 	}
// }
