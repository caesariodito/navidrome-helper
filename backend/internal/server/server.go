package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"navidrome-helper/internal/config"
	"navidrome-helper/internal/jobs"
	"navidrome-helper/internal/library"
	"navidrome-helper/internal/store"
	"navidrome-helper/internal/util"
)

// Server wires HTTP handlers to the runner and store.
type Server struct {
	cfg    config.Config
	store  *store.Store
	runner *jobs.Runner
	index  *library.Indexer
}

func New(cfg config.Config, store *store.Store, runner *jobs.Runner, indexer *library.Indexer) *Server {
	return &Server{cfg: cfg, store: store, runner: runner, index: indexer}
}

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/search", s.handleSearch)
	r.Post("/api/import", s.handleImport)
	r.Get("/api/jobs", s.handleListJobs)
	r.Get("/api/jobs/{id}", s.handleGetJob)
	r.Get("/api/library", s.handleLibraryList)
	r.Post("/api/library/refresh", s.handleLibraryRefresh)

	return r
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	results := mockSearchResults(query)
	s.annotateExists(results)
	writeJSON(w, http.StatusOK, map[string]any{"items": results})
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "no items provided", http.StatusBadRequest)
		return
	}

	dedup := map[string]importItem{}
	for _, it := range req.Items {
		if it.Type == "" {
			it.Type = "album"
		}
		if strings.ToLower(it.Type) == "song" && it.AlbumID != "" {
			if _, ok := dedup[it.AlbumID]; ok {
				continue
			}
			dedup[it.AlbumID] = importItem{
				ID:         it.AlbumID,
				Type:       "album",
				Title:      it.AlbumTitle,
				Artist:     it.Artist,
				AlbumID:    it.AlbumID,
				AlbumTitle: it.AlbumTitle,
				CoverURL:   it.CoverURL,
			}
			continue
		}
		if _, ok := dedup[it.ID]; ok {
			continue
		}
		dedup[it.ID] = it
	}

	var items []store.JobItem
	artist := ""
	album := ""
	for _, v := range dedup {
		if artist == "" {
			artist = v.Artist
		}
		if album == "" {
			album = v.Title
		}
		items = append(items, store.JobItem{
			SourceID:   v.ID,
			SourceType: v.Type,
			Title:      v.Title,
			Artist:     v.Artist,
			Album:      v.AlbumTitle,
			CoverURL:   v.CoverURL,
			Status:     jobs.StatusQueued,
			Message:    "queued",
		})
	}

	job := &store.Job{
		ID:       uuid.NewString(),
		Status:   jobs.StatusQueued,
		Phase:    jobs.PhaseQueued,
		Message:  "queued",
		Artist:   artist,
		Album:    album,
		Progress: 0,
		Items:    items,
	}

	if err := s.store.InsertJob(job); err != nil {
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}
	s.runner.Enqueue(job)
	writeJSON(w, http.StatusAccepted, map[string]string{"jobId": job.ID})
}

func (s *Server) handleListJobs(w http.ResponseWriter, r *http.Request) {
	jobsList, err := s.store.ListJobs(50)
	if err != nil {
		http.Error(w, "failed to list jobs", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": jobsList})
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := s.store.GetJob(id)
	if err != nil {
		http.Error(w, "failed to fetch job", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func (s *Server) handleLibraryList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("refresh") == "true" && s.index != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		if _, err := s.index.Refresh(ctx); err != nil {
			http.Error(w, "failed to refresh library", http.StatusInternalServerError)
			return
		}
	}
	entries, err := s.store.ListLibrary()
	if err != nil {
		http.Error(w, "failed to list library", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"library": entries})
}

func (s *Server) handleLibraryRefresh(w http.ResponseWriter, r *http.Request) {
	if s.index == nil {
		http.Error(w, "indexer not available", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	entries, err := s.index.Refresh(ctx)
	if err != nil {
		http.Error(w, "failed to refresh library", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"library": entries})
}

type importRequest struct {
	Items []importItem `json:"items"`
}

type importItem struct {
	ID         string `json:"id"`
	Type       string `json:"type"` // album|song
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	AlbumID    string `json:"albumId"`
	AlbumTitle string `json:"albumTitle"`
	CoverURL   string `json:"coverUrl"`
}

type searchResult struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Artist     string `json:"artist"`
	AlbumID    string `json:"albumId,omitempty"`
	AlbumTitle string `json:"albumTitle,omitempty"`
	CoverURL   string `json:"coverUrl"`
	Tracks     int    `json:"tracks,omitempty"`
	Duration   int    `json:"duration,omitempty"`
	Exists     bool   `json:"exists"`
}

func mockSearchResults(query string) []searchResult {
	q := strings.TrimSpace(strings.ToLower(query))
	if q == "" || strings.Contains("demo", q) || len(q) >= 0 {
		return []searchResult{
			{
				ID:         "alb_demo_1",
				Type:       "album",
				Title:      "Lights & Echoes",
				Artist:     "Demo Ensemble",
				AlbumID:    "alb_demo_1",
				AlbumTitle: "Lights & Echoes",
				CoverURL:   "https://placehold.co/200x200?text=Album",
				Tracks:     10,
				Duration:   2300,
			},
			{
				ID:         "alb_demo_single_parent",
				Type:       "song",
				Title:      "Silent Rivers",
				Artist:     "Demo Ensemble",
				AlbumID:    "alb_demo_single_parent",
				AlbumTitle: "Silent Rivers (Single)",
				CoverURL:   "https://placehold.co/200x200?text=Single",
				Tracks:     1,
				Duration:   210,
			},
			{
				ID:         "alb_electro_2024",
				Type:       "album",
				Title:      "Cities in Motion",
				Artist:     "Pulse Runner",
				AlbumID:    "alb_electro_2024",
				AlbumTitle: "Cities in Motion",
				CoverURL:   "https://placehold.co/200x200?text=Album",
				Tracks:     12,
				Duration:   2600,
			},
		}
	}
	return []searchResult{}
}

func (s *Server) annotateExists(results []searchResult) {
	for idx := range results {
		res := &results[idx]
		artist := util.NormalizeName(res.Artist)
		album := util.NormalizeName(res.Title)
		if res.Type == "song" && res.AlbumTitle != "" {
			album = util.NormalizeName(res.AlbumTitle)
		}
		if artist == "" || album == "" {
			res.Exists = false
			continue
		}
		ok, err := s.store.LibraryExists(artist, album)
		if err != nil {
			res.Exists = false
			continue
		}
		res.Exists = ok
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
