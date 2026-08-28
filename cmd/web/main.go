package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"catchup-x-post/internal/webarticle"
)

func main() {
	_ = godotenv.Load()

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "./output"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/articles", articlesListHandler(outputDir))
	mux.HandleFunc("/api/articles/", articlesDetailHandler(outputDir))

	addr := ":" + port
	log.Printf("web: listening on %s (output_dir=%s)", addr, outputDir)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func articlesListHandler(outputDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		articles, err := webarticle.List(outputDir)
		if err != nil {
			log.Printf("web: list articles: %v", err)
			writeError(w, http.StatusInternalServerError, "failed to list articles")
			return
		}
		writeJSON(w, http.StatusOK, articles)
	}
}

func articlesDetailHandler(outputDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/articles/")
		article, err := webarticle.Get(outputDir, id)
		if errors.Is(err, webarticle.ErrNotFound) {
			writeError(w, http.StatusNotFound, "article not found")
			return
		}
		if err != nil {
			log.Printf("web: get article %q: %v", id, err)
			writeError(w, http.StatusInternalServerError, "failed to load article")
			return
		}
		writeJSON(w, http.StatusOK, article)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("web: encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
