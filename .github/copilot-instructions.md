# GitHub Copilot Configuration for LearnLoop

This file provides guidance to GitHub Copilot for generating code and documentation for the LearnLoop project.

## Go Best Practices

### Code Style

- Follow standard Go formatting using `gofmt`
- Use camelCase for variable and function names
- Use PascalCase for exported functions, variables, constants, and types
- Use snake_case for file names
- Keep lines under 100 characters when possible

### Error Handling

- Always check and handle errors explicitly
- Return errors rather than using panic
- Use custom error types for specific error cases
- Follow the pattern: `if err != nil { return nil, err }`

### Comments and Documentation

- Write godoc-style comments for all exported functions and types
- Begin comments with the name of the thing being documented
- Use complete sentences with proper punctuation
- Example:

  ```go
  // Client represents a single WebSocket connection and handles
  // bidirectional communication between the WebSocket and the Hub.
  type Client struct {
      // fields...
  }
  ```

### Testing

- Write table-driven tests when appropriate
- Use meaningful test names prefixed with "Test"
- Separate test utilities into *_test.go files

### WebSocket Specifics

- Use goroutines carefully, ensuring they terminate properly
- Always close channels and connections when done
- Implement proper error handling for WebSocket connections
- Use thread-safe operations when accessing shared resources

## Markdown Best Practices

### Document Structure

- Use ATX-style headers (`#` for H1, `##` for H2, etc.)
- Maintain a single H1 (`#`) header at the top of the document
- Structure documents with logical heading hierarchy
- Include a table of contents for documents longer than 3 sections

### Code Blocks

- Use triple backticks for code blocks with language specification
- Example:
  