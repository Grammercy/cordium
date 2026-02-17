package handlers

import (
	"html/template"
	"io"
	"path/filepath"
	"testing"
)

func BenchmarkHomeTemplate(b *testing.B) {
	// The path must be relative to where the test runs (internal/handlers)
	// Or absolute.
	path := filepath.Join("..", "..", "web", "templates", "layout.html")

	// Verify file exists first to avoid confusing errors
	if _, err := template.ParseFiles(path); err != nil {
		b.Fatalf("Failed to verify template path %s: %v", path, err)
	}

	b.Run("ParseEveryTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				b.Fatalf("ParseFiles failed: %v", err)
			}
			// Execute with empty data just to complete the cycle
			tmpl.Execute(io.Discard, nil)
		}
	})

	// Pre-parse for the optimized case
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		b.Fatalf("Pre-parse failed: %v", err)
	}

	b.Run("ParseOnce", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := tmpl.Execute(io.Discard, nil); err != nil {
				b.Fatalf("Execute failed: %v", err)
			}
		}
	})
}
