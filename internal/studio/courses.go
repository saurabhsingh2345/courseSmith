package studio

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/enfec/coursesmith/internal/project"
)

// CreateCourseRequest is the POST /api/courses payload.
type CreateCourseRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// handleCourseCreate scaffolds a new course (course.yaml + an example lesson)
// via the same code path as `coursesmith init`, then returns the list-view
// Course so the UI can navigate straight to it.
func (s *Server) handleCourseCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("name is required"))
		return
	}

	res, err := project.ScaffoldCourse(s.coursesDir, req.Name, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, project.ErrCourseExists):
			writeError(w, http.StatusConflict, err)
		case errors.Is(err, project.ErrEmptySlug):
			writeError(w, http.StatusBadRequest, err)
		default:
			writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	c, err := project.LoadCourse(res.CourseDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	lessons, _ := c.Lessons()
	writeJSON(w, http.StatusCreated, Course{
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		LessonCount: len(lessons),
	})
}
