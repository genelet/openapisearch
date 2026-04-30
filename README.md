# openapisearch

`openapisearch` is a small Go library and CLI for finding, validating, caching,
and importing OpenAPI documents from public catalogs.

It searches APIs.guru first and can fall back to public-apis by probing common
OpenAPI and Swagger paths from catalog landing pages. Imported documents are
treated as untrusted data: the library only downloads, validates, and writes
OpenAPI/Swagger documents. It does not execute workflows or call discovered APIs.

## Install

```bash
go install github.com/genelet/openapisearch/cmd/openapisearch@latest
```

From a local checkout:

```bash
go run ./cmd/openapisearch --help
```

## CLI

```bash
go run ./cmd/openapisearch search --query slack
go run ./cmd/openapisearch search --query slack --json
go run ./cmd/openapisearch search --query slack --cache ~/.cache/openapisearch/cache.sqlite
go run ./cmd/openapisearch search --query slack --cache ~/.cache/openapisearch/cache.sqlite --offline
go run ./cmd/openapisearch import --url https://example.com/openapi.yaml --dir ./openapi --name example
go run ./cmd/openapisearch import --url https://example.com/openapi.yaml --dir ./openapi --cache ~/.cache/openapisearch/cache.sqlite --offline
```

Search sources:

- `auto`: search APIs.guru first, then public-apis fallback.
- `apis-guru`: search the APIs.guru OpenAPI directory only.
- `public-apis`: filter public-apis entries and probe common OpenAPI paths.

## Library

```go
package main

import (
	"context"

	"github.com/genelet/openapisearch"
)

func main() {
	ctx := context.Background()
	client := &openapisearch.Client{}
	report, err := client.Search(ctx, openapisearch.SearchOptions{
		Query:  "slack",
		Source: openapisearch.SourceAuto,
		Limit:  10,
	})
	_, _ = report, err
}
```

## SQLite Cache

Caching is opt-in. Pass a SQLite cache to the client, or use `--cache` in the
CLI. The default cache mode is `read-write` when a cache is configured.

```go
package main

import (
	"context"
	"time"

	"github.com/genelet/openapisearch"
	"github.com/genelet/openapisearch/sqlitecache"
)

func main() error {
	ctx := context.Background()
	cache, err := sqlitecache.Open("cache.sqlite")
	if err != nil {
		return err
	}
	defer cache.Close()

	client := &openapisearch.Client{Cache: cache}
	_, err = client.Search(ctx, openapisearch.SearchOptions{
		Query:       "slack",
		Source:      openapisearch.SourceAuto,
		Limit:       10,
		CacheMode:   openapisearch.CacheModeReadWrite,
		CacheMaxAge: 24 * time.Hour,
	})
	return err
}
```

Cache modes:

- `read-write`: use fresh cached data first, then refresh from the network on miss.
- `refresh`: bypass cache reads and write successful network results.
- `offline`: use only cached search results or OpenAPI documents.
- `bypass`: ignore the cache.

The default client enforces:

- HTTP/HTTPS only
- localhost, private, link-local, multicast, and unspecified address rejection
- redirect checks
- bounded response size
- request timeout
- OpenAPI/Swagger content validation before import

## Development

```bash
go test ./...
go vet ./...
git diff --check
```
