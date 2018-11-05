# webflow

## Status

Not ready for public consumption. Please see the todo section below.

## Description

Go (golang) Webflow API client. It attempts retries, with an exponential backoff, when it encounters rate limiting or server errors. _Thanks to pester package for the simplified retry logic._

Currently supports:

* Get all collections.
* Get collection by name.
* Get all items in collection by collection ID.
* Get all items in collection by collection name.
* General GET method requests.

## Examples

Get all items from a collection named "posts":

```go
  api := webflow.New(webflowAPIToken, webflowSiteID)
  collection, err := api.GetCollectionByName("posts")
  if err != nil {
    // fmt.Println(err)
    return err
  }

  fmt.Printf("collection: %+v\n", collection)
```

## Todo

So much. :)

Features are added as needed for a personal project however please know you may open a pull request for features if you like.

The code was under great flux at the time this was published.

Items of interest:

* Allow custom structs for API interactions.
* Replace MethodGet() with something more general since the internal code supports all request methods.
