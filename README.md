# webflow

## Status

Not ready for public consumption. Please see the todo section below.

## Description

Go (golang) Webflow API client. It attempts retries, with an exponential backoff, when it encounters rate limiting or server errors. _Thanks to pester package for the simplified retry logic._

Currently supports:

* General GET method requests.
* Get all collections.
* Get collection by name.
* Get all items in collection by collection ID.
* Get all items in collection by collection name.

## Examples

Get all items from a collection named "posts":

```go
  import (
    "encoding/json"
    "fmt"

    "github.com/redeemed2011/webflowAPI"
  )

  type myItem struct {
    Name      string `json:"name"`
    Body      string `json:"body"`
    Something string `json:"something"`
    ID        string `json:"_id"`
  }

  func getItems() error {
    api := webflowAPI.New("my token", "my site ID")
    items := []myItem{}
    err := api.GetAllItemsInCollectionByName("posts", 10, func(jsonItems json.RawMessage) error {
      tempItems := &[]myItem{}
      if err2 := json.Unmarshal(jsonItems, tempItems); err2 != nil {
        return fmt.Errorf("API did not return the proper collection items type. Error %+v", err2)
      }
      items = append(items, *tempItems...)
      return nil
    })

    if err != nil {
      fmt.Errorf("Error getting collection items: %+v\n", err)
    }

    fmt.Printf("collection items: %+v\n", items)
  }
```

## Todo

So much. :)

Features are added as needed for a personal project however please know you may open a pull request for features if you like.

The code was under great flux at the time this was published.

Items of interest:

* ~~Allow custom structs for API interactions.~~
* Replace MethodGet() with something more general since the internal code supports all request methods.
