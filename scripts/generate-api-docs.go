package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

// OpenAPISpec represents the basic structure of an OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string                 `yaml:"openapi"`
	Info       map[string]interface{} `yaml:"info"`
	Servers    []map[string]string    `yaml:"servers"`
	Paths      map[string]interface{} `yaml:"paths"`
	Components map[string]interface{} `yaml:"components"`
}

// RouteInfo contains information about a route extracted from Go code
type RouteInfo struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Handler     string
}

func main() {
	// Parse handlers directory for route annotations
	routes, err := parseHandlers("internal/handlers")
	if err != nil {
		log.Fatal("Error parsing handlers:", err)
	}

	// Load existing OpenAPI spec
	spec, err := loadOpenAPISpec("api/openapi.yaml")
	if err != nil {
		log.Fatal("Error loading OpenAPI spec:", err)
	}

	// Update paths in the spec
	updateOpenAPISpec(spec, routes)

	// Save updated spec
	err = saveOpenAPISpec("api/openapi.yaml", spec)
	if err != nil {
		log.Fatal("Error saving OpenAPI spec:", err)
	}

	fmt.Println("API documentation updated successfully!")
}

func parseHandlers(dir string) ([]RouteInfo, error) {
	var routes []RouteInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileRoutes, err := parseFile(path)
		if err != nil {
			return err
		}

		routes = append(routes, fileRoutes...)
		return nil
	})

	return routes, err
}

func parseFile(filename string) ([]RouteInfo, error) {
	var routes []RouteInfo

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Doc != nil {
				route := parseDocComments(fn.Doc.Text())
				if route.Method != "" && route.Path != "" {
					route.Handler = fn.Name.Name
					routes = append(routes, route)
				}
			}
		}
	}

	return routes, nil
}

func parseDocComments(comments string) RouteInfo {
	var route RouteInfo

	// Parse swagger-style comments
	routerRegex := regexp.MustCompile(`@Router\s+(\S+)\s+\[(\w+)\]`)
	summaryRegex := regexp.MustCompile(`@Summary\s+(.+)`)
	descriptionRegex := regexp.MustCompile(`@Description\s+(.+)`)
	tagsRegex := regexp.MustCompile(`@Tags\s+(.+)`)

	if match := routerRegex.FindStringSubmatch(comments); len(match) > 2 {
		route.Path = match[1]
		route.Method = strings.ToLower(match[2])
	}

	if match := summaryRegex.FindStringSubmatch(comments); len(match) > 1 {
		route.Summary = match[1]
	}

	if match := descriptionRegex.FindStringSubmatch(comments); len(match) > 1 {
		route.Description = match[1]
	}

	if match := tagsRegex.FindStringSubmatch(comments); len(match) > 1 {
		tags := strings.Split(match[1], ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		route.Tags = tags
	}

	return route
}

func loadOpenAPISpec(filename string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	err = yaml.Unmarshal(data, &spec)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

func updateOpenAPISpec(spec *OpenAPISpec, routes []RouteInfo) {
	if spec.Paths == nil {
		spec.Paths = make(map[string]interface{})
	}

	for _, route := range routes {
		pathItem, exists := spec.Paths[route.Path]
		if !exists {
			pathItem = make(map[string]interface{})
			spec.Paths[route.Path] = pathItem
		}

		pathMap := pathItem.(map[string]interface{})

		operation := map[string]interface{}{
			"summary":     route.Summary,
			"description": route.Description,
			"operationId": route.Handler,
			"tags":        route.Tags,
		}

		pathMap[route.Method] = operation
	}
}

func saveOpenAPISpec(filename string, spec *OpenAPISpec) error {
	data, err := yaml.Marshal(spec)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
