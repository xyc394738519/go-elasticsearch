# go-elasticsearch

The official Go client for [Elasticsearch](https://www.elastic.co/products/elasticsearch).

[![GoDoc](https://godoc.org/github.com/elastic/go-elasticsearch?status.svg)](http://godoc.org/github.com/elastic/go-elasticsearch)
[![Travis-CI](https://travis-ci.org/elastic/go-elasticsearch.svg?branch=master)](https://travis-ci.org/elastic/go-elasticsearch)
[![Go Report Card](https://goreportcard.com/badge/github.com/elastic/go-elasticsearch)](https://goreportcard.com/report/github.com/elastic/go-elasticsearch)
[![codecov.io](https://codecov.io/github/elastic/go-elasticsearch/coverage.svg?branch=master)](https://codecov.io/gh/elastic/go-elasticsearch?branch=master)

## Caveats

We encourage you to try the package in your projects, just keep these caveats in mind, please:

* **This is a work in progress.** Not all the planned features, standard in official Elasticsearch clients — retries on failures, auto-discovering nodes, ... — are implemented yet.
* **There are no guarantees on API stability.** Though the public APIs have been designed very carefully, they can change in a backwards-incompatible way depending on further exploration and user feedback.
* **The `master` branch targets Elasticsearch 7.x.** The [6.x](https://github.com/elastic/go-elasticsearch/tree/6.x) and [5.x](https://github.com/elastic/go-elasticsearch/tree/5.x) branches are compatible with the respective major versions of Elasticsearch; proper compatibility matrix and release tags will be added soon.

<!-- ----------------------------------------------------------------------------------------------- -->

## Installation

Install the package with `go get`:

    go get -u github.com/elastic/go-elasticsearch

Or, add the package to your `go.mod` file:

    require github.com/elastic/go-elasticsearch master

Or, clone the repository:

    git clone https://github.com/elastic/go-elasticsearch.git && cd go-elasticsearch

A complete example:

```bash
mkdir my-elasticsearch-app && cd my-elasticsearch-app

cat > go.mod <<-END
  module my-elasticsearch-app

  require github.com/elastic/go-elasticsearch master
END

cat > main.go <<-END
  package main

  import (
    "log"

    "github.com/elastic/go-elasticsearch"
  )

  func main() {
    es, _ := elasticsearch.NewDefaultClient()
    log.Println(es.Info())
  }
END

go run main.go
```


<!-- ----------------------------------------------------------------------------------------------- -->

## Usage

The `elasticsearch` package ties together two separate packages for calling the Elasticsearch APIs and transferring data over HTTP: `esapi` and `estransport`, respectively.

Use the `elasticsearch.NewDefaultClient()` function to create the client with the default settings.

```golang
es, err := elasticsearch.NewDefaultClient()
if err != nil {
  log.Fatalf("Error creating the client: %s", err)
}

res, err := es.Info()
if err != nil {
  log.Fatalf("Error getting response: %s", err)
}

log.Println(res)

// [200 OK] {
//   "name" : "node-1",
//   "cluster_name" : "go-elasticsearch"
// ...
```

When you export the `ELASTICSEARCH_URL` environment variable,
it will be used to set the cluster endpoint(s). Separate multiple adresses by a comma.

To set the cluster endpoint(s) programatically, pass them in the configuration object
to the `elasticsearch.NewClient()` function.

```golang
cfg := elasticsearch.Config{
  Addresses: []string{
    "http://localhost:9200",
    "http://localhost:9201",
  },
}
es, err := elasticsearch.NewClient(cfg)
// ...
```

To configure the HTTP settings, pass a [`http.Transport`](https://golang.org/pkg/net/http/#Transport)
object in the configuration object (the values are for illustrative purposes only).

```golang
cfg := elasticsearch.Config{
  Transport: &http.Transport{
    MaxIdleConnsPerHost:   10,
    ResponseHeaderTimeout: time.Second,
    DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
    TLSClientConfig: &tls.Config{
      MinVersion: tls.VersionTLS11,
      // ...
    },
  },
}

es, err := elasticsearch.NewClient(cfg)
// ...
```

See the [`_examples/configuration.go`](_examples/configuration.go) and
[`_examples/customization.go`](_examples/customization.go) files for
more examples of configuration and customization of the client.

The following example demonstrates a more complex usage. It fetches the Elasticsearch version from the cluster, indexes a couple of documents concurrently, and prints the search results, using a light wrapper around the response body.

```golang
// $ go run _examples/main.go

package main

import (
  "context"
  "encoding/json"
  "log"
  "strconv"
  "strings"
  "sync"

  "github.com/elastic/go-elasticsearch"
  "github.com/elastic/go-elasticsearch/esapi"
)

func main() {
  log.SetFlags(0)

  var (
    r  map[string]interface{}
    wg sync.WaitGroup
  )

  // Initialize a client with the default settings.
  //
  // An `ELASTICSEARCH_URL` environment variable will be used when exported.
  //
  es, err := elasticsearch.NewDefaultClient()
  if err != nil {
    log.Fatalf("Error creating the client: %s", err)
  }

  // 1. Get cluster info
  //
  res, err := es.Info()
  if err != nil {
    log.Fatalf("Error getting response: %s", err)
  }
  // Check response status
  if res.IsError() {
    log.Fatalf("Error: %s", res.String())
  }
  // Deserialize the response into a map.
  if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
    log.Fatalf("Error parsing the response body: %s", err)
  }
  // Print client and server version numbers.
  log.Printf("Client: %s", elasticsearch.Version)
  log.Printf("Server: %s", r["version"].(map[string]interface{})["number"])
  log.Println(strings.Repeat("~", 37))

  // 2. Index documents concurrently
  //
  for i, title := range []string{"Test One", "Test Two"} {
    wg.Add(1)

    go func(i int, title string) {
      defer wg.Done()

      // Set up the request object directly.
      req := esapi.IndexRequest{
        Index:      "test",
        DocumentID: strconv.Itoa(i + 1),
        Body:       strings.NewReader(`{"title" : "` + title + `"}`),
        Refresh:    "true",
      }

      // Perform the request with the client.
      res, err := req.Do(context.Background(), es)
      if err != nil {
        log.Fatalf("Error getting response: %s", err)
      }
      defer res.Body.Close()

      if res.IsError() {
        log.Printf("[%s] Error indexing document ID=%d", res.Status(), i+1)
      } else {
        // Deserialize the response into a map.
        var r map[string]interface{}
        if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
          log.Printf("Error parsing the response body: %s", err)
        } else {
          // Print the response status and indexed document version.
          log.Printf("[%s] %s; version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
        }
      }
    }(i, title)
  }
  wg.Wait()

  log.Println(strings.Repeat("-", 37))

  // 3. Search for the indexed documents
  //
  // Use the helper methods of the client.
  res, err = es.Search(
    es.Search.WithContext(context.Background()),
    es.Search.WithIndex("test"),
    es.Search.WithBody(strings.NewReader(`{"query" : { "match" : { "title" : "test" } }}`)),
    es.Search.WithTrackTotalHits(true),
    es.Search.WithPretty(),
  )
  if err != nil {
    log.Fatalf("ERROR: %s", err)
  }
  defer res.Body.Close()

  if res.IsError() {
    var e map[string]interface{}
    if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
      log.Fatalf("error parsing the response body: %s", err)
    } else {
      // Print the response status and error information.
      log.Fatalf("[%s] %s: %s",
        res.Status(),
        e["error"].(map[string]interface{})["type"],
        e["error"].(map[string]interface{})["reason"],
      )
    }
  }

  if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
    log.Fatalf("Error parsing the response body: %s", err)
  }
  // Print the response status, number of results, and request duration.
  log.Printf(
    "[%s] %d hits; took: %dms",
    res.Status(),
    int(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
    int(r["took"].(float64)),
  )
  // Print the ID and document source for each hit.
  for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
    log.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
  }

  log.Println(strings.Repeat("=", 37))
}

// Client: 7.0.0-SNAPSHOT
// Server: 7.0.0-SNAPSHOT
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// [201 Created] updated; version=1
// [201 Created] updated; version=1
// -------------------------------------
// [200 OK] 2 hits; took: 5ms
//  * ID=1, map[title:Test One]
//  * ID=2, map[title:Test Two]
// =====================================
```

As you see in the example above, the `esapi` package allows to call the Elasticsearch APIs in two distinct ways: either by creating a struct, such as `IndexRequest`, and calling its `Do()` method by passing it a context and the client, or by calling the `Search()` function on the client directly, using the option functions such as `WithIndex()`. See more information and examples in the
[package documentation](https://godoc.org/github.com/elastic/go-elasticsearch/esapi).

The `estransport` package handles the transfer of data to and from Elasticsearch. At the moment, the implementation is really minimal: it only round-robins across the configured cluster endpoints. In future, more features — retrying failed requests, ignoring certain status codes, auto-discovering nodes in the cluster, and so on — will be added.

<!-- ----------------------------------------------------------------------------------------------- -->

## Examples

The **[`_examples`](./_examples)** folder contains a number of recipes and comprehensive examples to get you started with the client, including configuration and customization of the client, mocking the transport for unit tests, embedding the client in a custom type, building queries, performing requests, and parsing the responses.

<!-- ----------------------------------------------------------------------------------------------- -->

## License

(c) 2019 Elasticsearch. Licensed under the Apache License, Version 2.0.
