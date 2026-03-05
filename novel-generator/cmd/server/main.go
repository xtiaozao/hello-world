package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"novel-generator/pkg/automation"
	"novel-generator/pkg/novel"
	"sync"
)

type Server struct {
	Generator    *novel.Generator
	CurrentNovel *novel.Novel
	Mutex        sync.Mutex
}

func main() {
	server := &Server{
		// Use qwen3:8b as default fallback if configured one is missing,
		// but here we explicitly set it because we saw it in `ollama list`.
		Generator:    novel.NewGenerator("qwen3:8b"),
		CurrentNovel: &novel.Novel{},
	}

	// Serve static files
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	http.HandleFunc("/api/outline", server.handleOutline)
	http.HandleFunc("/api/chapters", server.handleChapters)
	http.HandleFunc("/api/chapter/content", server.handleChapterContent)
	http.HandleFunc("/api/download", server.handleDownload)
	http.HandleFunc("/api/upload", server.handleUpload)

	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (s *Server) handleOutline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title string `json:"title"`
		Intro string `json:"intro"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	outline, err := s.Generator.GenerateOutline(req.Title, req.Intro)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.Mutex.Lock()
	s.CurrentNovel.Title = req.Title
	s.CurrentNovel.Intro = req.Intro
	s.CurrentNovel.Outline = outline
	s.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"outline": outline})
}

func (s *Server) handleChapters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Outline string `json:"outline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.Mutex.Lock()
	s.CurrentNovel.Outline = req.Outline
	s.Mutex.Unlock()

	chapters, err := s.Generator.GenerateChapters(req.Outline)
	if err != nil {
		writeJSONError(w, fmt.Sprintf("Failed to generate chapters: %v", err), http.StatusInternalServerError)
		return
	}

	s.Mutex.Lock()
	s.CurrentNovel.Chapters = chapters
	s.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"chapters": chapters})
}

func (s *Server) handleChapterContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Chapter     novel.Chapter `json:"chapter"`
		PrevSummary string        `json:"prev_summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	content, err := s.Generator.GenerateChapterContent(req.Chapter, req.PrevSummary)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the chapter content in memory
	s.Mutex.Lock()
	for i, c := range s.CurrentNovel.Chapters {
		if c.ID == req.Chapter.ID {
			s.CurrentNovel.Chapters[i].Content = content
			break
		}
	}
	s.Mutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": content})
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	s.Mutex.Lock()
	novel := s.CurrentNovel
	s.Mutex.Unlock()

	w.Header().Set("Content-Disposition", "attachment; filename=novel.txt")
	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "Title: %s\n\nIntro: %s\n\nOutline:\n%s\n\n", novel.Title, novel.Intro, novel.Outline)
	for _, ch := range novel.Chapters {
		fmt.Fprintf(w, "Chapter %d: %s\n\n%s\n\n", ch.ID, ch.Title, ch.Content)
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URL     string `json:"url"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For security, limit URL domain or something, but this is a local tool.
	// We'll trust the user input for now.

	err := automation.UploadChapter(req.URL, req.Title, req.Content)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
