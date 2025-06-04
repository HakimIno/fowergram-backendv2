# Fowergram API Documentation

This directory contains the API documentation for the Fowergram backend service.

## Files

- `openapi.yaml` - OpenAPI 3.1 specification defining all REST API endpoints
- `stoplight.html` - Stoplight Elements documentation viewer
- `schema.graphql` - GraphQL schema definition (located at root level)

## Viewing Documentation

### Option 1: Live Server (Recommended)
```bash
# Start the main application
make dev

# Visit the documentation at:
# http://localhost:8000/docs
```

### Option 2: Local Static Server
```bash
# Serve documentation locally
make docs-serve

# Visit: http://localhost:3000/docs/stoplight.html
```

### Option 3: Direct File
Open `stoplight.html` directly in your browser for a static view.

## Updating Documentation

### Automatic Updates
When you add new API endpoints with proper annotations, run:
```bash
make docs-update
```

This will:
1. Parse your handler files for `@Summary`, `@Description`, `@Tags`, `@Router` annotations
2. Update the OpenAPI specification
3. Validate the specification

### Manual Updates
Edit `openapi.yaml` directly for complex changes or to add:
- Request/response schemas
- Authentication requirements
- Examples
- Additional metadata

## Adding New API Endpoints

1. **Create Handler Function** with proper annotations:
```go
// CreatePost creates a new post
// @Summary Create a new post
// @Description Create a new post with title, content, and optional media
// @Tags Posts
// @Accept json
// @Produce json
// @Param request body CreatePostRequest true "Post creation request"
// @Success 201 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/posts [post]
func (h *PostHandler) CreatePost(c *fiber.Ctx) error {
    // Implementation here
}
```

2. **Add Route** in `internal/routes/routes.go`:
```go
posts := api.Group("/posts")
posts.Use(cfg.AuthService.Middleware())
posts.Post("/", cfg.PostHandler.CreatePost)
```

3. **Update Documentation**:
```bash
make docs-update
```

4. **Add Schemas** to `openapi.yaml` if needed:
```yaml
components:
  schemas:
    CreatePostRequest:
      type: object
      required:
        - title
        - content
      properties:
        title:
          type: string
          example: "My First Post"
        content:
          type: string
          example: "This is the content of my first post"
```

## Documentation Standards

### Annotation Format
Use these annotations in your handler functions:

- `@Summary` - Brief description (required)
- `@Description` - Detailed description
- `@Tags` - Grouping tag (e.g., Authentication, Posts, Users)
- `@Accept` - Content types accepted (json, multipart/form-data, etc.)
- `@Produce` - Content types produced (json, xml, etc.)
- `@Param` - Parameters (path, query, body, header)
- `@Success` - Success responses with status codes
- `@Failure` - Error responses with status codes
- `@Security` - Security requirements (BearerAuth, etc.)
- `@Router` - Route path and HTTP method

### Schema Naming
- Use PascalCase for schema names
- Suffix request schemas with "Request"
- Suffix response schemas with "Response"
- Use descriptive names (e.g., `CreatePostRequest`, `PostListResponse`)

### Response Codes
Standard HTTP response codes:
- `200` - Success (GET, PUT, PATCH)
- `201` - Created (POST)
- `204` - No Content (DELETE)
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (authentication required)
- `403` - Forbidden (authorization failed)
- `404` - Not Found
- `409` - Conflict (duplicate resources)
- `422` - Unprocessable Entity (business logic errors)
- `500` - Internal Server Error

## Integration with Stoplight

This documentation is optimized for [Stoplight Elements](https://stoplight.io/open-source/elements), which provides:

- Interactive API documentation
- Try-it-out functionality
- Schema validation
- Code generation
- Mock servers

The documentation automatically updates when you:
1. Add new endpoints with proper annotations
2. Run `make docs-update`
3. Deploy the application

## GraphQL Documentation

GraphQL documentation is available at:
- **Playground**: http://localhost:8000/playground (development only)
- **Schema**: http://localhost:8000/graphql (introspection enabled)

## Validation

Validate your OpenAPI specification:
```bash
make docs-validate
```

This ensures your documentation follows OpenAPI standards and is compatible with various tools.

## Best Practices

1. **Keep Documentation Current**: Update docs with every API change
2. **Use Examples**: Provide realistic examples in requests/responses
3. **Document Errors**: Include all possible error responses
4. **Security First**: Document authentication requirements
5. **Version Control**: Keep API docs in version control
6. **Test Examples**: Ensure all examples work with the actual API

## Troubleshooting

### Documentation Not Loading
- Check that `openapi.yaml` is valid YAML
- Ensure all referenced schemas exist
- Verify file paths in `stoplight.html`

### Missing Endpoints
- Ensure handler functions have proper annotations
- Run `make docs-update` after adding new endpoints
- Check that routes are properly registered

### Validation Errors
- Use `make docs-validate` to check for issues
- Common issues: missing required fields, invalid references, wrong format

For more help, see the [OpenAPI Specification](https://swagger.io/specification/) and [Stoplight Elements Documentation](https://meta.stoplight.io/docs/elements). 