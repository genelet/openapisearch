# openapisearch

`openapisearch` is a Go library and CLI for finding, validating, caching, and
importing OpenAPI documents from public catalogs. It also provides a public
OpenAPI-backed core for AI-assisted authoring: callers can combine
natural-language briefs with prompt-safe operation context to draft UWS-aligned
`project.md` and `intent.hcl` artifacts.

It searches APIs.guru first and can fall back to public-apis by probing common
OpenAPI and Swagger paths from catalog landing pages. Imported documents are
treated as untrusted data: the library only downloads, validates, and writes
OpenAPI/Swagger documents. It does not execute workflows or call discovered APIs.

The shared authoring flow is:

```text
brief + OpenAPI docs -> prompt-safe operation context -> project.md / intent.hcl draft -> caller-specific leaf renderer
```

The generated artifacts are drafts. Downstream tools such as Ramen and OpenUdon
must validate, review, and render them through their own domain-specific leaf
logic. `openapisearch` does not execute APIs, inject credentials, bind concrete
runtimes, or perform production side effects.

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

Local project directories can be scanned without network access:

```go
results, err := openapisearch.LocalFiles(context.Background(), openapisearch.LocalOptions{
	Dir:     "./openapi",
	BaseDir: ".",
	Query:   "slack messages",
})
_, _ = results, err
```

`LocalFiles` accepts draft local OpenAPI/Swagger documents with top-level
metadata so authoring tools can find work-in-progress specs. Remote search and
import continue to use stricter validation before returning or writing specs.

## Authoring Core

The authoring core is intended for callers that need OpenAPI-backed assistance
without coupling to a particular runtime or product workflow. It can own common
concepts such as operation inventories, ranked prompt-safe summaries,
transcripts, diagnostics, slots, assumptions, symbolic bindings, readiness
issues, question plans, and artifact sets.

Ramen and OpenUdon embed these shared concepts directly and provide their own
leaf renderers:

- Ramen owns Ramen-specific `project.md`, workflow `intent.hcl`,
  `workflow.hcl`, UWS generation, Symphony review packages, trusted runner
  wrappers, evals, and private udon integration.
- OpenUdon owns concrete IaC intent models, Terraform generation,
  graph/profile/planning, state/drift/handoff bundles, and w8m-facing public
  IaC artifacts.

Ramen and OpenUdon depend on `openapisearch`; they do not depend on each other.
See [docs/authoring.md](docs/authoring.md) for the shared concepts and binding
model.

```go
core := openapisearch.NewAuthoringCore()
opctx, artifacts, err := openapisearch.DraftFromOpenAPI(
	context.Background(),
	core,
	openapisearch.Brief{
		Text:        "Create one support ticket from runtime inputs.",
		ProjectName: "Support Ticket Draft",
	},
	[]openapisearch.OpenAPIDoc{{Path: "openapi/support.yaml"}},
	[]string{"createTicket"},
)
_, _, _ = opctx, artifacts, err

leaf := openapisearch.NewLeafAdapter(artifacts, openapisearch.LeafOptions{
	Name:   "Support Ticket Draft",
	Source: "example",
})
reviewMarkdown := leaf.ReviewMarkdown()
_ = reviewMarkdown
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
