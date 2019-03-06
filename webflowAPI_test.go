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
		Slug: "dogs1",
	}
	exampleItemDog1 = &mockItem{
		ID:   "1",
		Name: "blue",
	}
	exampleItemDog1JSON, _ = json.Marshal(*exampleItemDog1)
	exampleItemDog2        = &mockItem{
		ID:   "2",
		Name: "green",
	}
	exampleItemDog2JSON, _ = json.Marshal(*exampleItemDog2)
	exampleItemDog3        = &mockItem{
		ID:   "3",
		Name: "red",
	}
	exampleItemDog3JSON, _ = json.Marshal(*exampleItemDog3)
	exampleItemDog4        = &mockItem{
		ID:   "4",
		Name: "brown",
	}
	exampleItemDog4JSON, _               = json.Marshal(*exampleItemDog4)
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
	exampleItemsDogsJSON2   = [][]byte{
		exampleItemDog1JSON,
		exampleItemDog2JSON,
		exampleItemDog3JSON,
		exampleItemDog4JSON,
	}
	apiResponseItemsDogs = &CollectionItems{
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
	exampleCatCollection = &Collection{
		ID:   "2",
		Name: "cats",
		Slug: "cats1",
	}
	exampleCollections = &Collections{
		*exampleDogCollection,
		*exampleCatCollection,
	}
)

func TestMethodGet(t *testing.T) {
	{
		// Start a special, local HTTP server.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`Unexpected`))
		}))
		defer server.Close()

		res := &mockItem{}
		api := New("mytoken", "mysiteid", &http.Client{})
		api.BaseURL = server.URL
		err := api.MethodGet("/", nil, res)
		if err == nil {
			t.Error("Get() is expected to return an error when receiving an unknown response format.")
		}

		if !strings.Contains(err.Error(), "invalid character") {
			t.Error("Get() should error stating the unknown response format has an invalid character.")
		}
	}
	{
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
		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		// Setup a backoff that is always just 1 millisecond.
		api.Client.Backoff = func(retry int) time.Duration {
			return 1 * time.Millisecond
		}
		err := api.MethodGet("/", nil, res)

		if err == nil {
			t.Error("MethodGet() is expected to return an error when an API error response is received.")
		}

		if tries < 2 {
			t.Errorf(
				"MethodGet() is expected to retry when rate limiting errors are encountered! It tried the request %d times.",
				tries,
			)
		}
	}
	{
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
		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		query := map[string]string{
			"offset": "2",
			"limit":  "1",
		}
		err := api.MethodGet("/", query, res)

		if err != nil {
			t.Error("MethodGet() is expected to return no error when no error is encountered.")
		}

		if !reflect.DeepEqual(exampleItemDog1, res) {
			t.Errorf("MethodGet() did not return the expected values! Got %+v.", res)
		}
	}
}

func TestGetAllCollections(t *testing.T) {
	{
		// Start a special, local HTTP server.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusRequestTimeout)
			rw.Write([]byte{})
		}))
		defer server.Close()

		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		_, err := api.GetAllCollections()

		if err == nil {
			t.Error("GetAllCollections() is expected to return a func error when receiving an unknown error.")
		}
	}
	{
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

		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		_, err := api.GetAllCollections()

		if err == nil {
			t.Error("GetAllCollections() is expected to return a func error when receiving an unknown error.")
		}

		if err.Error() != errMsg {
			t.Errorf("GetAllCollections() returned an incorrect error message! Got '%s'; expected '%s'.", err.Error(), errMsg)
		}
	}
	{
		api := New("mytoken", siteID, nil)
		api.methodGet = func(uri string, queryParams map[string]string, decodedResponse interface{}) error {
			tmpJSON, err := json.Marshal(exampleCollections)
			if err != nil {
				return err
			}
			return json.Unmarshal(tmpJSON, decodedResponse)
		}

		res, err := api.GetAllCollections()

		if err != nil {
			t.Error("GetAllCollections() is expected to return no func error when receiving a properly formatted response.")
		}

		if res == nil || len(*res) != len(*exampleCollections) {
			t.Errorf(
				"GetAllCollections() is expected to return %d collections! Got %d.",
				len(*exampleCollections),
				len(*res),
			)
		}
	}
}

func TestGetCollectionByName(t *testing.T) {
	api := New("mytoken", siteID, nil)
	api.methodGet = func(uri string, queryParams map[string]string, decodedResponse interface{}) error {
		tmpJSON, err := json.Marshal(exampleCollections)
		if err != nil {
			return err
		}
		return json.Unmarshal(tmpJSON, decodedResponse)
	}

	// Test searching for a collection that exists.
	{
		collection, err := api.GetCollectionByName(exampleDogCollection.Name)

		if err != nil {
			t.Error("GetCollectionByName() is expected to return no func error when receiving a properly formatted response.")
		}

		if !reflect.DeepEqual(collection, exampleDogCollection) {
			t.Errorf("GetCollectionByName() is expected to return exampleDogCollection! Got %+v.", collection)
		}
	}

	// Test searching for a collection that does not exist.
	{
		collection, _ := api.GetCollectionByName("birds")

		if collection != nil {
			t.Errorf("GetCollectionByName() is expected to return nil whenever the collection is not found! Got %+v.", collection)
		}
	}
}

func TestGetCollectionBySlug(t *testing.T) {
	api := New("mytoken", siteID, nil)
	api.methodGet = func(uri string, queryParams map[string]string, decodedResponse interface{}) error {
		tmpJSON, err := json.Marshal(exampleCollections)
		if err != nil {
			return err
		}
		return json.Unmarshal(tmpJSON, decodedResponse)
	}

	// Test searching for a collection that exists.
	{
		collection, err := api.GetCollectionBySlug(exampleDogCollection.Slug)

		if err != nil {
			t.Error("GetCollectionBySlug() is expected to return no func error when receiving a properly formatted response.")
		}

		if !reflect.DeepEqual(collection, exampleDogCollection) {
			t.Errorf("GetCollectionBySlug() is expected to return exampleDogCollection! Got %+v.", collection)
		}
	}

	// Test searching for a collection that does not exist.
	{
		collection, _ := api.GetCollectionBySlug("birds")

		if collection != nil {
			t.Errorf("GetCollectionBySlug() is expected to return nil whenever the collection is not found! Got %+v.", collection)
		}
	}
}

func TestGetAllItemsInCollectionByID(t *testing.T) {
	// Test the GetAllItemsInCollectionByID func for one collection type (dogs).
	{
		// Start a special, local HTTP server.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			reqURI := req.URL.String()
			queryParams := req.URL.Query()
			expectedURI := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)
			if !strings.HasPrefix(reqURI, expectedURI) {
				t.Errorf(
					"GetAllItemsInCollectionByID() did not request the proper URI! requested '%s'; expected '%s'.",
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

		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		itemsJSON, err := api.GetAllItemsInCollectionByID(exampleDogCollection.ID, 10)

		if err != nil {
			t.Errorf(
				"GetAllItemsInCollectionByID() is expected to return no error when receiving a properly formatted response. got: %+v",
				err,
			)
		}

		items := []mockItem{}
		for _, itemJSON := range itemsJSON {
			tmpItem := &mockItem{}
			if err2 := json.Unmarshal(itemJSON, tmpItem); err2 != nil {
				t.Errorf(
					"GetAllItemsInCollectionByID() did not return the proper collection items type. Error %+v",
					err2,
				)
			}

			items = append(items, *tmpItem)
		}

		if !reflect.DeepEqual(items, *exampleItemsDogs) {
			t.Errorf("GetAllItemsInCollectionByID() is expected to return exampleItemsDogs! Got %+v.", items)
		}
	}

	// Test the GetAllItemsInCollectionByID func for another collection type (cats).
	{
		// Start a special, local HTTP server.
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			reqURI := req.URL.String()
			queryParams := req.URL.Query()
			expectedURI := fmt.Sprintf(listCollectionItemsURL, exampleCatCollection.ID)
			if !strings.HasPrefix(reqURI, expectedURI) {
				t.Errorf(
					"GetAllItemsInCollectionByID() did not request the proper URI! requested '%s'; expected '%s'.",
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

		api := New("mytoken", siteID, nil)
		api.BaseURL = server.URL
		itemsJSON, err := api.GetAllItemsInCollectionByID(exampleCatCollection.ID, 10)

		if err != nil {
			t.Errorf(
				"GetAllItemsInCollectionByID() is expected to return no error when receiving a properly formatted response. got: %+v",
				err,
			)
		}

		items := []mockItem{}
		for _, itemJSON := range itemsJSON {
			tmpItem := &mockItem{}
			if err2 := json.Unmarshal(itemJSON, tmpItem); err2 != nil {
				t.Errorf(
					"GetAllItemsInCollectionByID() did not return the proper collection items type. Error %+v",
					err2,
				)
			}

			items = append(items, *tmpItem)
		}

		if !reflect.DeepEqual(items, *exampleItemsCats) {
			t.Errorf("GetAllItemsInCollectionByID() is expected to return exampleItemsCats! Got %+v.", items)
		}
	}
}

// Test the GetAllItemsInCollectionByName func for one collection type (dogs).
func TestGetAllItemsInCollectionByName(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		collectionURL := fmt.Sprintf(listCollectionsURL, siteID)
		itemsURL := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)
		var data []byte
		switch true {
		case strings.HasPrefix(reqURI, collectionURL):
			data, _ = json.Marshal(exampleCollections)
		case strings.HasPrefix(reqURI, itemsURL):
			data, _ = json.Marshal(apiResponseItemsDogs)
		default:
			t.Errorf(
				"GetAllItemsInCollectionByID() did not request the proper URI! requested '%s'; expected '%s' or '%s'.",
				reqURI,
				collectionURL,
				itemsURL,
			)
		}

		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID, nil)
	api.BaseURL = server.URL
	itemsJSON, err := api.GetAllItemsInCollectionByName(exampleDogCollection.Name, 10)

	if err != nil {
		t.Errorf(
			"GetAllItemsInCollectionByName() is expected to return no error when receiving a properly formatted response. got: %+v",
			err,
		)
	}

	items := []mockItem{}
	for _, itemJSON := range itemsJSON {
		tmpItem := &mockItem{}
		if err2 := json.Unmarshal(itemJSON, tmpItem); err2 != nil {
			t.Errorf(
				"GetAllItemsInCollectionByID() did not return the proper collection items type. Error %+v",
				err2,
			)
		}

		items = append(items, *tmpItem)
	}

	if !reflect.DeepEqual(items, *exampleItemsDogs) {
		t.Errorf("GetAllItemsInCollectionByName() is expected to return exampleItemsDogs! Got %+v.", items)
	}
}

// Test the GetAllItemsInCollectionBySlug func for one collection type (dogs).
func TestGetAllItemsInCollectionBySlug(t *testing.T) {
	// Start a special, local HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		reqURI := req.URL.String()
		collectionURL := fmt.Sprintf(listCollectionsURL, siteID)
		itemsURL := fmt.Sprintf(listCollectionItemsURL, exampleDogCollection.ID)
		var data []byte
		switch true {
		case strings.HasPrefix(reqURI, collectionURL):
			data, _ = json.Marshal(exampleCollections)
		case strings.HasPrefix(reqURI, itemsURL):
			data, _ = json.Marshal(apiResponseItemsDogs)
		default:
			t.Errorf(
				"GetAllItemsInCollectionByID() did not request the proper URI! requested '%s'; expected '%s' or '%s'.",
				reqURI,
				collectionURL,
				itemsURL,
			)
		}

		rw.Write(data)
	}))
	defer server.Close()

	api := New("mytoken", siteID, nil)
	api.BaseURL = server.URL
	itemsJSON, err := api.GetAllItemsInCollectionBySlug(exampleDogCollection.Slug, 10)

	if err != nil {
		t.Errorf(
			"GetAllItemsInCollectionBySlug() is expected to return no error when receiving a properly formatted response. got: %+v",
			err,
		)
	}

	items := []mockItem{}
	for _, itemJSON := range itemsJSON {
		tmpItem := &mockItem{}
		if err2 := json.Unmarshal(itemJSON, tmpItem); err2 != nil {
			t.Errorf(
				"GetAllItemsInCollectionByID() did not return the proper collection items type. Error %+v",
				err2,
			)
		}

		items = append(items, *tmpItem)
	}

	if !reflect.DeepEqual(items, *exampleItemsDogs) {
		t.Errorf("GetAllItemsInCollectionBySlug() is expected to return exampleItemsDogs! Got %+v.", items)
	}
}

func TestGetItem(t *testing.T) {
	api := New("mytoken", siteID, nil)
	api.getAllItemsInCollectionByName = func(name string, maxPages int) ([][]byte, error) {
		return exampleItemsDogsJSON2, nil
	}
	api.getAllItemsInCollectionBySlug = func(slug string, maxPages int) ([][]byte, error) {
		return exampleItemsDogsJSON2, nil
	}
	api.getAllItemsInCollectionByID = func(id string, maxPages int) ([][]byte, error) {
		return exampleItemsDogsJSON2, nil
	}

	{
		item, err := api.GetItem("", "", "", "", "")
		if err != nil {
			t.Errorf("GetItem() is expected to not error when neither an item name nor an ID is given: %+v", err)
		}
		if item != nil {
			t.Errorf(
				"GetItem() is expected to return nothing whenever neither an item name nor an ID is given. Got: %+v",
				item,
			)
		}
	}
	{
		item, err := api.GetItem(exampleDogCollection.Name, "", "", "z", "")
		if err != nil {
			t.Errorf("GetItem() is expected to not error when an item name that is not found is given: %+v", err)
		}
		if item != nil {
			t.Errorf("GetItem() is expected to return nothing if the item is not found. Got: %+v", item)
		}
	}
	{
		item, err := api.GetItem("", exampleDogCollection.Slug, "", "z", "")
		if err != nil {
			t.Errorf("GetItem() is expected to not error when an item slug that is not found is given: %+v", err)
		}
		if item != nil {
			t.Errorf("GetItem() is expected to return nothing if the item is not found by slug. Got: %+v", item)
		}
	}
	{
		item, err := api.GetItem("", "", exampleDogCollection.ID, exampleItemDog1.Name, "")
		if err != nil {
			t.Errorf("GetItem() is expected to not error when an item name that is found is given: %+v", err)
		}
		if !reflect.DeepEqual(item, exampleItemDog1JSON) {
			t.Errorf(
				"GetItem() is expected to return the appropriate item by name.\nExpected: %T %+v\nGot: %T %+v",
				exampleItemDog1JSON,
				string(exampleItemDog1JSON),
				item,
				string(item),
			)
		}
	}
	{
		item, err := api.GetItem("", exampleDogCollection.Slug, "", "", exampleItemDog3.ID)
		if err != nil {
			t.Errorf("GetItem() is expected to not error when a collection slug & item ID is given: %+v", err)
		}
		if !reflect.DeepEqual(item, exampleItemDog3JSON) {
			t.Errorf(
				"GetItem() is expected to return the appropriate item by a collection slug & item ID.\nExpected: %T %+v\nGot: %T %+v",
				exampleItemDog3JSON,
				string(exampleItemDog3JSON),
				item,
				string(item),
			)
		}
	}
	{
		item, err := api.GetItem(exampleDogCollection.Name, "", "", "", exampleItemDog3.ID)
		if err != nil {
			t.Errorf("GetItem() is expected to not error when an ID is given: %+v", err)
		}
		if !reflect.DeepEqual(item, exampleItemDog3JSON) {
			t.Errorf(
				"GetItem() is expected to return the appropriate item by ID.\nExpected: %T %+v\nGot: %T %+v",
				exampleItemDog3JSON,
				string(exampleItemDog3JSON),
				item,
				string(item),
			)
		}
	}
	{
		item, err := api.GetItem("", "", exampleDogCollection.ID, exampleItemDog4.Name, exampleItemDog4.ID)
		if err != nil {
			t.Errorf("GetItem() is expected to not error when a name and ID are given: %+v", err)
		}
		if !reflect.DeepEqual(item, exampleItemDog4JSON) {
			t.Errorf(
				"GetItem() is expected to return the appropriate item by name and ID.\nExpected: %T %+v\nGot: %T %+v",
				exampleItemDog4JSON,
				string(exampleItemDog4JSON),
				item,
				string(item),
			)
		}
	}
}
