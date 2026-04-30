# AGENTS.md

## Purpose

`openapisearch` is an open-source Go library and CLI for discovering, validating,
caching, and importing OpenAPI or Swagger documents from public catalogs.

The package is intentionally generic. It is shared by private consumers such as
Ramen and udon, but it must not contain Ramen workflow semantics, udon runtime
behavior, product-specific policy, or private infrastructure assumptions.

## Boundaries

- Keep provider discovery, URL safety, OpenAPI/Swagger validation, local cache,
  and CLI behavior in this repository.
- Put workflow authoring, synthesis policy, trusted-runner gates, and example
  artifacts in Ramen.
- Put UWS/OpenAPI execution, lowering, and runtime behavior in udon.
- Put public UWS semantics in `../uws`.
- Do not add LLM-generated spec synthesis here without a deliberate design pass.

Rule of thumb:

- If it helps find or safely import OpenAPI documents, it belongs here.
- If it executes an API or workflow, it does not belong here.
- If it depends on a private repo or product workflow, it does not belong here.

## Commands

```bash
go test ./...
go vet ./...
git diff --check
go run ./cmd/openapisearch search --query slack
go run ./cmd/openapisearch search --query slack --cache /tmp/openapisearch.sqlite
go run ./cmd/openapisearch import --url https://example.com/openapi.yaml --dir /tmp/openapi
```

When changing public APIs, also run dependent checks in local sibling consumers
when available:

```bash
(cd ../ramen && go test ./...)
(cd ../udon && go test ./...)
```

## Go Conventions

- Primary language is Go.
- Keep `cmd/openapisearch` thin; reusable behavior belongs in the root package
  or a focused subpackage such as `sqlitecache`.
- Keep the root package dependency-light. Optional storage integrations should
  live in subpackages.
- Preserve exported API compatibility unless the change is intentionally
  breaking and documented.
- Prefer deterministic tests with `httptest` over live network tests.

## Safety

- Treat all discovered OpenAPI documents as untrusted.
- Never execute operations from a discovered document.
- Enforce HTTP/HTTPS only for remote fetches.
- Reject localhost, private, link-local, multicast, and unspecified addresses
  by default.
- Keep redirect limits, response-size limits, and request timeouts in place.
- Do not cache secrets, credentials, tokens, or workflow execution data.
- Do not add HTML scraping or LLM-synthesized specs without explicit review and
  provenance labeling.
