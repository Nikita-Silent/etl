package main

import (
	"net/http"
	"os"

	scalargo "github.com/bdpiprava/scalar-go"
)

// openAPISpec загружает OpenAPI спецификацию.
var openAPISpec = func() []byte {
	data, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		return []byte{}
	}
	return data
}()

// docsHandler обрабатывает запросы к /api/docs.
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if len(openAPISpec) == 0 {
		s.logger.Error("OpenAPI specification not loaded",
			"event", "openapi_spec_not_loaded",
		)
		http.Error(w, "OpenAPI specification not available", http.StatusInternalServerError)
		return
	}

	html, err := scalargo.NewV2(
		scalargo.WithSpecBytes(openAPISpec),
		scalargo.WithTheme(scalargo.ThemeDefault),
		scalargo.WithSearchHotKey("k"),
		scalargo.WithDefaultHTTPClient("javascript", "fetch"),
	)
	if err != nil {
		s.logger.Error("Failed to generate API documentation",
			"error", err.Error(),
			"event", "docs_generation_error",
		)
		http.Error(w, "Failed to generate documentation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// openAPIHandler обрабатывает запросы к /api/openapi.yaml.
func (s *Server) openAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	yamlContent, err := os.ReadFile("api/openapi.yaml")
	if err != nil {
		s.logger.Error("Failed to read openapi.yaml",
			"error", err.Error(),
			"event", "openapi_file_read_error",
		)
		http.Error(w, "OpenAPI specification not available", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(yamlContent)
}
