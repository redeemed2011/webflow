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

    itemsJSON, err := api.GetAllItemsInCollectionByName("posts", 10)

    if err != nil {
      return fmt.Errorf("Error getting collection items: %+v\n", err)
    }

    items := []mockItem{}
    for _, itemJSON := range itemsJSON {
      tmpItems := &[]mockItem{}
      if err2 := json.Unmarshal(itemJSON, tmpItems); err2 != nil {
        return fmt.Errorf("API did not return the proper collection items type. Error %+v", err2)
      }

      items = append(items, *tmpItems...)
    }


    fmt.Printf("collection items: %+v\n", items)
  }
```

## Todo

So much. :)

Features are added as needed for a personal project however please know you may open a pull request for features if you like.

The code was under great flux at the time this was published.

Items of interest:

* Evaluate returning lists of items as raw JSON to simplify this pkg's api. At the time of this writing, getting all items in a collection requires the caller to provide a method to decode the JSON. This does not sit well with me and seems to be more complicated than necessary. If we instead return raw JSON the process is simpler and the caller can then decode the JSON however is desired.
* Provide methods to get filtered collection items by ID or name. This may be accomplished by decoding only the `id` & `name` fields for the filter then returning the full raw JSON on match.
* ~~Allow custom structs for API interactions.~~ _Update: may not be worth the investment with the above changes._
* Replace `MethodGet()` with something more general since the internally used HTTP pkgs support all request methods.
* Perhaps make `MethodGet()` private rather than exported.
* Implement Go Modules support for this pkg--specifically versioning of this pkg.
