package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/joho/godotenv"

	"catchup-x-post/internal/generator"
	"catchup-x-post/internal/webarticle"
)

const (
	maxKeywordLength   = 100
	maxGenerateTimeout = 5 * time.Minute
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

	var generateMu sync.Mutex

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/articles", articlesListHandler(outputDir))
	mux.HandleFunc("/api/articles/", articlesDetailHandler(outputDir))
	mux.HandleFunc("/api/generate", generateHandler(&generateMu))

	addr := ":" + port
	log.Printf("web: listening on %s (output_dir=%s)", addr, outputDir)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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

type generateRequest struct {
	Keyword string `json:"keyword"`
	Count   int    `json:"count"`
}

type generateArticleResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type generateResponse struct {
	Articles []generateArticleResponse `json:"articles"`
}

// generateHandler は POST /api/generate を処理する。
// 同時実行は generateMu で1件に制限し、count * 1分（上限5分）でタイムアウトする。
func generateHandler(generateMu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		keyword := strings.TrimSpace(req.Keyword)
		if keyword == "" {
			writeError(w, http.StatusBadRequest, "keyword is required")
			return
		}
		if utf8.RuneCountInString(keyword) > maxKeywordLength {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("keyword must be %d characters or fewer", maxKeywordLength))
			return
		}
		if req.Count < 1 || req.Count > generator.MaxArticleCount {
			writeError(w, http.StatusBadRequest, fmt.Sprintf("count must be between 1 and %d", generator.MaxArticleCount))
			return
		}

		if !generateMu.TryLock() {
			writeError(w, http.StatusConflict, "generation already in progress")
			return
		}

		timeout := time.Duration(req.Count) * time.Minute
		if timeout > maxGenerateTimeout {
			timeout = maxGenerateTimeout
		}

		type runOutcome struct {
			result *generator.Result
			err    error
		}
		done := make(chan runOutcome, 1)
		go func() {
			// generateMu は生成が完了するまで保持する。HTTP 側がタイムアウトで
			// 先に応答を返しても、バックグラウンドの生成処理と解除タイミングを
			// 一致させ、同時実行を確実に1件に制限する。
			defer generateMu.Unlock()
			result, err := generator.Run(generator.Options{DryRun: true, Count: req.Count, Topic: keyword})
			done <- runOutcome{result, err}
		}()

		select {
		case outcome := <-done:
			if outcome.result == nil {
				log.Printf("web: generate: %v", outcome.err)
				writeError(w, http.StatusInternalServerError, "failed to generate articles: "+outcome.err.Error())
				return
			}
			if outcome.err != nil {
				log.Printf("web: generate (partial): %v", outcome.err)
			}
			writeJSON(w, http.StatusOK, toGenerateResponse(outcome.result))
		case <-time.After(timeout):
			log.Printf("web: generate: timed out after %s (keyword=%q count=%d)", timeout, keyword, req.Count)
			writeError(w, http.StatusGatewayTimeout, "generation timed out")
		}
	}
}

func toGenerateResponse(result *generator.Result) generateResponse {
	resp := generateResponse{Articles: []generateArticleResponse{}}
	for _, a := range result.Articles {
		resp.Articles = append(resp.Articles, generateArticleResponse{ID: a.ID, Title: a.Title})
	}
	return resp
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
