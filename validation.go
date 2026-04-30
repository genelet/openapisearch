package openapisearch

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

func specMetadata(ctx context.Context, content []byte) (SpecMetadata, bool) {
	trimmed := bytes.TrimSpace(content)
	if len(trimmed) == 0 {
		return SpecMetadata{}, false
	}
	var root map[string]any
	if trimmed[0] == '{' {
		if err := json.Unmarshal(trimmed, &root); err != nil {
			return SpecMetadata{}, false
		}
	} else if err := yaml.Unmarshal(trimmed, &root); err != nil {
		return SpecMetadata{}, false
	}
	openapi, _ := root["openapi"].(string)
	swagger, _ := root["swagger"].(string)
	if strings.TrimSpace(openapi) == "" && strings.TrimSpace(swagger) == "" {
		return SpecMetadata{}, false
	}
	if containsExternalRef(root) {
		return SpecMetadata{}, false
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if strings.TrimSpace(openapi) != "" {
		loader := openapi3.NewLoader()
		loader.Context = ctx
		doc, err := loader.LoadFromData(trimmed)
		if err != nil {
			return SpecMetadata{}, false
		}
		if doc.OpenAPIMajorMinor() == "" || doc.Info == nil || doc.Paths == nil {
			return SpecMetadata{}, false
		}
		if err := doc.Validate(loader.Context); err != nil {
			return SpecMetadata{}, false
		}
		return SpecMetadata{
			Title:          strings.TrimSpace(doc.Info.Title),
			Description:    strings.TrimSpace(doc.Info.Description),
			OpenAPI:        strings.TrimSpace(doc.OpenAPI),
			OperationCount: operationCountV3(doc.Paths),
		}, true
	}

	if !validSwagger2Root(root) {
		return SpecMetadata{}, false
	}
	var doc openapi2.T
	if err := yaml.Unmarshal(trimmed, &doc); err != nil {
		return SpecMetadata{}, false
	}
	if strings.TrimSpace(doc.Swagger) != "2.0" || strings.TrimSpace(doc.Info.Title) == "" || strings.TrimSpace(doc.Info.Version) == "" || doc.Paths == nil {
		return SpecMetadata{}, false
	}
	return SpecMetadata{
		Title:          strings.TrimSpace(doc.Info.Title),
		Description:    strings.TrimSpace(doc.Info.Description),
		Swagger:        strings.TrimSpace(doc.Swagger),
		OperationCount: operationCountV2(doc.Paths),
	}, true
}

func containsExternalRef(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "$ref" {
				ref, _ := child.(string)
				if isExternalRef(ref) {
					return true
				}
			}
			if containsExternalRef(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsExternalRef(child) {
				return true
			}
		}
	}
	return false
}

func isExternalRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	return ref != "" && !strings.HasPrefix(ref, "#")
}

func validSwagger2Root(root map[string]any) bool {
	swagger, _ := root["swagger"].(string)
	if strings.TrimSpace(swagger) != "2.0" {
		return false
	}
	info, ok := root["info"].(map[string]any)
	if !ok {
		return false
	}
	title, _ := info["title"].(string)
	version, _ := info["version"].(string)
	if strings.TrimSpace(title) == "" || strings.TrimSpace(version) == "" {
		return false
	}
	paths, ok := root["paths"].(map[string]any)
	if !ok {
		return false
	}
	for path, value := range paths {
		if !strings.HasPrefix(path, "/") {
			return false
		}
		pathItem, ok := value.(map[string]any)
		if !ok {
			return false
		}
		if !validSwagger2PathItem(pathItem) {
			return false
		}
	}
	return true
}

func validSwagger2PathItem(pathItem map[string]any) bool {
	for key, value := range pathItem {
		if strings.HasPrefix(key, "x-") || key == "$ref" {
			continue
		}
		if key == "parameters" {
			if _, ok := value.([]any); !ok {
				return false
			}
			continue
		}
		if !isHTTPMethod(key) {
			return false
		}
		operation, ok := value.(map[string]any)
		if !ok {
			return false
		}
		responses, ok := operation["responses"].(map[string]any)
		if !ok || len(responses) == 0 {
			return false
		}
		for _, responseValue := range responses {
			response, ok := responseValue.(map[string]any)
			if !ok {
				return false
			}
			if ref, _ := response["$ref"].(string); strings.TrimSpace(ref) != "" {
				continue
			}
			description, _ := response["description"].(string)
			if strings.TrimSpace(description) == "" {
				return false
			}
		}
	}
	return true
}

func isHTTPMethod(method string) bool {
	switch method {
	case "delete", "get", "head", "options", "patch", "post", "put":
		return true
	default:
		return false
	}
}

func operationCountV3(paths *openapi3.Paths) int {
	count := 0
	for _, path := range paths.Map() {
		if path == nil {
			continue
		}
		count += len(path.Operations())
	}
	return count
}

func operationCountV2(paths map[string]*openapi2.PathItem) int {
	count := 0
	for _, path := range paths {
		if path == nil {
			continue
		}
		count += len(path.Operations())
	}
	return count
}
