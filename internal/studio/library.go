package studio

// Reusable-asset library (Phase 7). Teachers save diagrams and questions they
// like and pull them into other lessons. This is net-new persistence with no
// other consumer, so it stores plain JSON arrays under the studio state dir
// (.coursesmith/library/), guarded by a mutex since the studio is single-user
// and low-traffic.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// libraryMu serializes read-modify-write on the library files.
var libraryMu sync.Mutex

// LibraryDiagram is a saved diagram spec (any kind: svg/d3/mermaid/excalidraw).
type LibraryDiagram struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Source    string `json:"source"`
	CreatedAt string `json:"created_at"`
}

// LibraryQuestion mirrors the pipeline quiz Question fields so a saved question
// can be dropped straight into a lesson quiz.
type LibraryQuestion struct {
	ID          string   `json:"id"`
	Prompt      string   `json:"prompt"`
	Type        string   `json:"type"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
	CreatedAt   string   `json:"created_at"`
}

func (s *Server) libraryFile(name string) string {
	return filepath.Join(s.stateDir, "library", name)
}

// loadLibrary reads a JSON array file, returning an empty slice if absent.
func loadLibrary[T any](path string) ([]T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []T{}, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return []T{}, nil
	}
	var items []T
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}
	return items, nil
}

// saveLibrary writes a JSON array atomically, creating the library dir.
func saveLibrary[T any](path string, items []T) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(path, data, 0o644)
}

// newLibraryID returns a monotonic-ish id from the wall clock (single-user).
func newLibraryID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func (s *Server) handleLibraryDiagramsList(w http.ResponseWriter, r *http.Request) {
	libraryMu.Lock()
	defer libraryMu.Unlock()
	items, err := loadLibrary[LibraryDiagram](s.libraryFile("diagrams.json"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleLibraryDiagramCreate(w http.ResponseWriter, r *http.Request) {
	var in LibraryDiagram
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}
	libraryMu.Lock()
	defer libraryMu.Unlock()
	path := s.libraryFile("diagrams.json")
	items, err := loadLibrary[LibraryDiagram](path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	in.ID = newLibraryID()
	in.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	items = append(items, in)
	if err := saveLibrary(path, items); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

func (s *Server) handleLibraryDiagramDelete(w http.ResponseWriter, r *http.Request) {
	libraryMu.Lock()
	defer libraryMu.Unlock()
	path := s.libraryFile("diagrams.json")
	items, err := loadLibrary[LibraryDiagram](path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id := r.PathValue("id")
	kept := make([]LibraryDiagram, 0, len(items))
	found := false
	for _, it := range items {
		if it.ID == id {
			found = true
			continue
		}
		kept = append(kept, it)
	}
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("no diagram %q", id))
		return
	}
	if err := saveLibrary(path, kept); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLibraryQuestionsList(w http.ResponseWriter, r *http.Request) {
	libraryMu.Lock()
	defer libraryMu.Unlock()
	items, err := loadLibrary[LibraryQuestion](s.libraryFile("questions.json"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleLibraryQuestionCreate(w http.ResponseWriter, r *http.Request) {
	var in LibraryQuestion
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(in.Prompt) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("prompt is required"))
		return
	}
	libraryMu.Lock()
	defer libraryMu.Unlock()
	path := s.libraryFile("questions.json")
	items, err := loadLibrary[LibraryQuestion](path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	in.ID = newLibraryID()
	in.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	items = append(items, in)
	if err := saveLibrary(path, items); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, in)
}

func (s *Server) handleLibraryQuestionDelete(w http.ResponseWriter, r *http.Request) {
	libraryMu.Lock()
	defer libraryMu.Unlock()
	path := s.libraryFile("questions.json")
	items, err := loadLibrary[LibraryQuestion](path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id := r.PathValue("id")
	kept := make([]LibraryQuestion, 0, len(items))
	found := false
	for _, it := range items {
		if it.ID == id {
			found = true
			continue
		}
		kept = append(kept, it)
	}
	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("no question %q", id))
		return
	}
	if err := saveLibrary(path, kept); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
