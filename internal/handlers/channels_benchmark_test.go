package handlers

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"testing"
)

func init() {
    // Attempt to change to project root if we are in internal/handlers
    // Check if "web" exists in current dir, if not try ../..
    // This is a heuristic for running tests.
    if _, err := os.Stat("web"); os.IsNotExist(err) {
        _ = os.Chdir("../..")
    }
}

func BenchmarkTemplateParsing(b *testing.B) {
    chatPath := filepath.Join("web", "templates", "chat.html")
    msgPath := filepath.Join("web", "templates", "message.html")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := template.New("chat.html").Funcs(funcMap).ParseFiles(chatPath, msgPath)
		if err != nil {
			b.Fatalf("Template parse error: %v", err)
		}
	}
}

func BenchmarkTemplateCached(b *testing.B) {
    // Initialize once
    if err := InitTemplates(); err != nil {
        b.Fatalf("InitTemplates failed: %v", err)
    }

    data := struct {
		GuildID     string
		ChannelID   string
		ChannelName string
		Messages    interface{}
		Members     interface{}
	}{
		GuildID:     "123",
		ChannelID:   "456",
		ChannelName: "general",
		Messages:    nil,
		Members:     nil,
	}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var buf bytes.Buffer
        if err := ChatTmpl.Execute(&buf, data); err != nil {
            b.Fatalf("Execute error: %v", err)
        }
    }
}
