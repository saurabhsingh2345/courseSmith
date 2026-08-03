package studio

// Hand-written OpenAPI 3.1 spec, served at /api/openapi.json and written to
// studio/openapi.json so the frontend generates its TS client from it
// (openapi-typescript). Raw pipeline artifacts (script, quiz, reviews) pass
// through loosely typed on purpose: the pipeline JSON files are their own
// source of truth and the UI treats them as data.

import (
	"net/http"
	"os"
)

const openAPISpec = `{
  "openapi": "3.1.0",
  "info": {
    "title": "coursesmith studio API",
    "version": "1.0.0",
    "description": "Local-first backend for the coursesmith Studio. Binds to localhost, no auth."
  },
  "paths": {
    "/api/courses": {
      "get": {
        "operationId": "listCourses",
        "responses": {
          "200": {
            "description": "All parseable courses",
            "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/Course"}}}}
          }
        }
      }
    },
    "/api/courses/{slug}": {
      "get": {
        "operationId": "getCourse",
        "parameters": [{"name": "slug", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {
          "200": {"description": "Course with lesson × stage matrix", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CourseDetail"}}}},
          "404": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/lessons/{course}/{id}": {
      "get": {
        "operationId": "getLesson",
        "parameters": [
          {"name": "course", "in": "path", "required": true, "schema": {"type": "string"}},
          {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {
          "200": {"description": "Full lesson state", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/LessonDetail"}}}},
          "404": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/run": {
      "get": {
        "operationId": "getRunStatus",
        "responses": {"200": {"description": "Current run", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/RunStatus"}}}}}
      },
      "post": {
        "operationId": "startRun",
        "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/RunRequest"}}}},
        "responses": {
          "202": {"description": "Run accepted", "content": {"application/json": {"schema": {"type": "object", "properties": {"run_id": {"type": "string"}}, "required": ["run_id"]}}}},
          "404": {"$ref": "#/components/responses/Error"},
          "409": {"$ref": "#/components/responses/Error"}
        }
      },
      "delete": {
        "operationId": "cancelRun",
        "responses": {
          "200": {"description": "Canceling"},
          "409": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/events": {
      "get": {
        "operationId": "streamEvents",
        "description": "Server-sent events: run-started, stage-started, stage-finished, stage-failed, log, run-finished, run-failed, quota. Supports Last-Event-ID resume.",
        "responses": {"200": {"description": "text/event-stream of Event objects", "content": {"text/event-stream": {"schema": {"$ref": "#/components/schemas/Event"}}}}}
      }
    },
    "/api/feedback": {
      "post": {
        "operationId": "addFeedback",
        "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/FeedbackRequest"}}}},
        "responses": {
          "200": {"description": "Note saved to review-notes.yaml"},
          "400": {"$ref": "#/components/responses/Error"},
          "404": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/regenerate": {
      "post": {
        "operationId": "regenerate",
        "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/RegenerateRequest"}}}},
        "responses": {
          "202": {"description": "Forced regeneration started", "content": {"application/json": {"schema": {"type": "object", "properties": {"run_id": {"type": "string"}, "stages": {"type": "array", "items": {"type": "string"}}}}}}},
          "400": {"$ref": "#/components/responses/Error"},
          "409": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/quiz/{course}/{id}": {
      "get": {
        "operationId": "getQuiz",
        "parameters": [
          {"name": "course", "in": "path", "required": true, "schema": {"type": "string"}},
          {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "responses": {"200": {"description": "Generated quiz + overrides + merged result", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/QuizWithOverrides"}}}}}
      }
    },
    "/api/quiz-overrides/{course}/{id}": {
      "put": {
        "operationId": "putQuizOverrides",
        "parameters": [
          {"name": "course", "in": "path", "required": true, "schema": {"type": "string"}},
          {"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}
        ],
        "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/QuizOverrides"}}}},
        "responses": {"200": {"description": "Saved; returns overrides + merged quiz", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/QuizWithOverrides"}}}}}
      }
    },
    "/api/ledger": {
      "get": {
        "operationId": "getLedger",
        "responses": {"200": {"description": "Token/cost/quota data", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Ledger"}}}}}
      }
    },
    "/api/reels": {
      "get": {"summary": "List reels", "responses": {"200": {"description": "reels", "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/ReelSummary"}}}}}}},
      "post": {"summary": "Create and run a reel", "responses": {"201": {"description": "created", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/ReelSummary"}}}}}}
    },
    "/api/reels/cast": {
      "post": {"summary": "Propose segments from a brief (writes nothing)", "responses": {"200": {"description": "proposal", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CastReelResponse"}}}}}}
    },
    "/api/reels/{id}": {
      "get": {"summary": "One reel with its segments", "responses": {"200": {"description": "reel", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/ReelDetail"}}}}}},
      "delete": {"summary": "Delete a reel", "responses": {"204": {"description": "deleted"}}}
    },
    "/api/reels/{id}/run": {
      "post": {"summary": "Re-run a reel", "responses": {"202": {"description": "queued"}}}
    },
    "/api/reels/{id}/segments/{segment}": {
      "patch": {"summary": "Edit one segment", "responses": {"200": {"description": "segments", "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/ReelSegmentInfo"}}}}}}}
    },
    "/api/snippet-templates": {
      "get": {
        "operationId": "listSnippetTemplates",
        "responses": {"200": {"description": "The visual template catalog", "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/SnippetTemplateInfo"}}}}}}
      }
    },
    "/api/snippets": {
      "get": {
        "operationId": "listSnippets",
        "responses": {"200": {"description": "Every snippet, newest first", "content": {"application/json": {"schema": {"type": "array", "items": {"$ref": "#/components/schemas/SnippetSummary"}}}}}}
      },
      "post": {
        "operationId": "createSnippet",
        "requestBody": {"required": true, "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateSnippetRequest"}}}},
        "responses": {
          "201": {"description": "Created; the pipeline is running", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateSnippetResponse"}}}},
          "202": {"description": "Created, but the pipeline was busy — re-run it later", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateSnippetResponse"}}}},
          "400": {"$ref": "#/components/responses/Error"}
        }
      }
    },
    "/api/snippets/{id}": {
      "get": {
        "operationId": "getSnippet",
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {
          "200": {"description": "The request plus the model's plan", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/SnippetDetail"}}}},
          "404": {"$ref": "#/components/responses/Error"}
        }
      },
      "delete": {
        "operationId": "deleteSnippet",
        "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
        "responses": {"204": {"description": "Deleted"}, "404": {"$ref": "#/components/responses/Error"}}
      }
    }
  },
  "components": {
    "responses": {
      "Error": {"description": "Error", "content": {"application/json": {"schema": {"type": "object", "properties": {"error": {"type": "string"}}, "required": ["error"]}}}}
    },
    "schemas": {
      "Course": {
        "type": "object",
        "properties": {"name": {"type": "string"}, "slug": {"type": "string"}, "description": {"type": "string"}, "lesson_count": {"type": "integer"}},
        "required": ["name", "slug", "description", "lesson_count"]
      },
      "LessonSummary": {
        "type": "object",
        "properties": {"id": {"type": "string"}, "title": {"type": "string"}, "stages": {"type": "object", "additionalProperties": {"type": "string", "enum": ["done", "stale", "pending"]}}},
        "required": ["id", "title", "stages"]
      },
      "CourseDetail": {
        "allOf": [
          {"$ref": "#/components/schemas/Course"},
          {"type": "object", "properties": {"stage_order": {"type": "array", "items": {"type": "string"}}, "lessons": {"type": "array", "items": {"$ref": "#/components/schemas/LessonSummary"}}}, "required": ["stage_order", "lessons"]}
        ]
      },
      "Artifact": {
        "type": "object",
        "properties": {
          "name": {"type": "string"}, "size": {"type": "integer"}, "url": {"type": "string"},
          "download_name": {"type": "string", "description": "What this file should be saved as: <course>-<lesson>[-<part>].<ext>, so a folder of downloads sorts by course and lesson instead of being six copies of final.mp4."}
        },
        "required": ["name", "size", "url", "download_name"]
      },
      "CastReelResponse": {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "segments": {"type": "array", "items": {"type": "object", "properties": {"template": {"type": "string"}, "prompt": {"type": "string"}, "material": {"type": "string", "description": "The concrete facts this template will be filled with. POST it back with the proposal — a segment created without it is planned from the one-line prompt alone, and its writer invents the specifics."}, "target_sec": {"type": "integer"}}, "required": ["template", "prompt"]}}
        },
        "required": ["title", "segments"]
      },
      "ReelSegmentInfo": {
        "type": "object",
        "properties": {
          "id": {"type": "string"}, "template": {"type": "string"}, "prompt": {"type": "string"},
          "material": {"type": "string", "description": "The concrete facts this segment is planned from. The field most worth correcting by hand: a wrong figure here becomes a wrong figure in the finished video."},
          "target_sec": {"type": "integer"}, "skip": {"type": "boolean"},
          "template_title": {"type": "string"}, "template_category": {"type": "string"}
        },
        "required": ["id", "template", "prompt", "template_title", "template_category"]
      },
      "ReelSummary": {
        "type": "object",
        "properties": {
          "id": {"type": "string"}, "title": {"type": "string"}, "brief": {"type": "string"},
          "segments": {"type": "integer"}, "skipped": {"type": "integer"},
          "ready": {"type": "boolean"}, "video_url": {"type": "string"}, "created_at": {"type": "string"},
          "run_id": {"type": "string"}
        },
        "required": ["id", "title", "segments", "skipped", "ready"]
      },
      "ReelDetail": {
        "allOf": [
          {"$ref": "#/components/schemas/ReelSummary"},
          {"type": "object", "properties": {"segment_list": {"type": "array", "items": {"$ref": "#/components/schemas/ReelSegmentInfo"}}, "plan": {}}, "required": ["segment_list"]}
        ]
      },
      "SnippetTemplateInfo": {
        "type": "object",
        "properties": {
          "name": {"type": "string"}, "title": {"type": "string"}, "description": {"type": "string"},
          "example": {"type": "string"}, "shows_code": {"type": "boolean"},
          "min_target_sec": {"type": "integer"}, "default_target_sec": {"type": "integer"},
          "category": {"type": "string"}, "category_title": {"type": "string"},
          "since": {"type": "string"}, "family": {"type": "string"}
        },
        "required": ["name", "title", "description", "example", "shows_code", "category", "category_title"]
      },
      "SnippetSummary": {
        "type": "object",
        "properties": {
          "id": {"type": "string"}, "title": {"type": "string"}, "prompt": {"type": "string"},
          "template": {"type": "string"}, "ready": {"type": "boolean"},
          "video_url": {"type": "string"}, "created_at": {"type": "string"}
        },
        "required": ["id", "title", "prompt", "template", "ready"]
      },
      "SnippetDetail": {
        "allOf": [
          {"$ref": "#/components/schemas/SnippetSummary"},
          {"type": "object", "properties": {"target_sec": {"type": "integer"}, "plan": {}}, "required": ["target_sec"]}
        ]
      },
      "CreateSnippetRequest": {
        "type": "object",
        "properties": {
          "prompt": {"type": "string"}, "template": {"type": "string"}, "title": {"type": "string"},
          "target_sec": {"type": "integer"}, "code_language": {"type": "string"},
          "voice": {"type": "string"}, "plan_only": {"type": "boolean"},
          "captions": {"type": "string", "enum": ["on", "off"]},
          "mode": {"type": "string", "enum": ["dark", "light"]},
          "skin": {"type": "string", "enum": ["default", "broadcast", "minimal"]}
        },
        "required": ["prompt", "template"]
      },
      "CreateSnippetResponse": {
        "allOf": [
          {"$ref": "#/components/schemas/SnippetSummary"},
          {"type": "object", "properties": {"run_id": {"type": "string"}}}
        ]
      },
      "LessonDetail": {
        "type": "object",
        "properties": {
          "course": {"type": "string"}, "id": {"type": "string"}, "title": {"type": "string"},
          "source": {"type": "string"},
          "stages": {"type": "object", "additionalProperties": {"type": "string"}},
          "stage_order": {"type": "array", "items": {"type": "string"}},
          "artifacts": {"type": "array", "items": {"$ref": "#/components/schemas/Artifact"}},
          "script": {"$ref": "#/components/schemas/Script"},
          "quiz": {"$ref": "#/components/schemas/Quiz"},
          "mistakes": {}, "exercises": {}, "chapters": {"type": "array", "items": {"$ref": "#/components/schemas/Chapter"}},
          "alignment": {}, "reviews": {"type": "object", "additionalProperties": {}}
        },
        "required": ["course", "id", "title", "source", "stages", "stage_order", "artifacts"]
      },
      "Script": {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "sections": {"type": "array", "items": {"type": "object", "properties": {
            "id": {"type": "string"}, "narration": {"type": "string"}, "duration_est_sec": {"type": "integer"},
            "cues": {"type": "array", "items": {"type": "object", "properties": {"type": {"type": "string"}, "ref": {"type": "string"}, "at_word": {"type": "integer"}}}}
          }, "required": ["id", "narration"]}}
        },
        "required": ["title", "sections"]
      },
      "Quiz": {
        "type": "object",
        "properties": {
          "title": {"type": "string"},
          "questions": {"type": "array", "items": {"$ref": "#/components/schemas/Question"}}
        },
        "required": ["title", "questions"]
      },
      "Question": {
        "type": "object",
        "properties": {
          "id": {"type": "string"}, "type": {"type": "string", "enum": ["recall", "application", "debugging", "prediction"]},
          "prompt": {"type": "string"}, "review": {"type": "boolean"},
          "options": {"type": "array", "items": {"type": "string"}},
          "answer_index": {"type": "integer"}, "explanation": {"type": "string"}
        },
        "required": ["id", "type", "prompt", "options", "answer_index", "explanation"]
      },
      "QuizOverrides": {
        "type": "object",
        "properties": {"questions": {"type": "array", "items": {"type": "object", "properties": {
          "id": {"type": "string"}, "drop": {"type": "boolean"}, "prompt": {"type": "string"},
          "options": {"type": "array", "items": {"type": "string"}}, "answer_index": {"type": "integer"}, "explanation": {"type": "string"}
        }, "required": ["id"]}}},
        "required": ["questions"]
      },
      "QuizWithOverrides": {
        "type": "object",
        "properties": {
          "generated": {"$ref": "#/components/schemas/Quiz"},
          "overrides": {"$ref": "#/components/schemas/QuizOverrides"},
          "merged": {"$ref": "#/components/schemas/Quiz"}
        }
      },
      "Chapter": {
        "type": "object",
        "properties": {"id": {"type": "string"}, "title": {"type": "string"}, "start_ms": {"type": "integer"}, "end_ms": {"type": "integer"}},
        "required": ["id", "title", "start_ms", "end_ms"]
      },
      "RunRequest": {
        "type": "object",
        "properties": {"course": {"type": "string"}, "lesson": {"type": "string"}, "stage": {"type": "string"}, "force": {"type": "boolean"}},
        "required": ["course", "lesson"]
      },
      "RunStatus": {
        "type": "object",
        "properties": {"running": {"type": "boolean"}, "run_id": {"type": "string"}, "course": {"type": "string"}, "lesson": {"type": "string"}, "stage": {"type": "string"}},
        "required": ["running"]
      },
      "Event": {
        "type": "object",
        "properties": {
          "type": {"type": "string", "enum": ["run-started", "stage-started", "stage-finished", "stage-skipped", "stage-failed", "log", "run-finished", "run-failed", "quota"]},
          "run_id": {"type": "string"}, "course": {"type": "string"}, "lesson": {"type": "string"},
          "stage": {"type": "string"}, "line": {"type": "string"}, "error": {"type": "string"},
          "seq": {"type": "integer"}, "at": {"type": "string", "format": "date-time"}
        },
        "required": ["type", "seq", "at"]
      },
      "FeedbackRequest": {
        "type": "object",
        "properties": {"course": {"type": "string"}, "lesson": {"type": "string"}, "section": {"type": "string"}, "note": {"type": "string"}, "timestamp_ms": {"type": "integer"}},
        "required": ["course", "lesson", "note"]
      },
      "RegenerateRequest": {
        "type": "object",
        "properties": {"course": {"type": "string"}, "lesson": {"type": "string"}, "artifact": {"type": "string", "enum": ["script", "quiz", "visuals", "mistakes", "exercises", "audio"]}, "section": {"type": "string"}, "note": {"type": "string"}},
        "required": ["course", "lesson", "artifact"]
      },
      "LedgerRow": {
        "type": "object",
        "properties": {"day": {"type": "string"}, "provider": {"type": "string"}, "model": {"type": "string"}, "calls": {"type": "integer"}, "prompt_tokens": {"type": "integer"}, "completion_tokens": {"type": "integer"}, "cost_usd": {"type": "number"}, "priced": {"type": "boolean", "description": "False when this model is absent from the pricing table, which makes cost_usd meaningless rather than zero. Render it as unknown, never as $0.00 — the tokens were really spent."}},
        "required": ["day", "provider", "model", "calls", "prompt_tokens", "completion_tokens", "cost_usd", "priced"]
      },
      "QuotaStatus": {
        "type": "object",
        "properties": {"provider": {"type": "string"}, "per_minute": {"type": "integer"}, "per_day": {"type": "integer"}, "day_used": {"type": "integer"}},
        "required": ["provider", "per_minute", "per_day", "day_used"]
      },
      "Ledger": {
        "type": "object",
        "properties": {
          "rows": {"type": "array", "items": {"$ref": "#/components/schemas/LedgerRow"}},
          "total_cost_usd": {"type": "number"}, "total_calls": {"type": "integer"},
          "unpriced_models": {"type": "array", "items": {"type": "string"}, "description": "Models whose spend is missing from total_cost_usd, so the total is a floor rather than the bill."},
          "unpriced_tokens": {"type": "integer"},
          "quotas": {"type": "array", "items": {"$ref": "#/components/schemas/QuotaStatus"}}
        },
        "required": ["rows", "total_cost_usd", "total_calls", "quotas"]
      }
    }
  }
}`

func (s *Server) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(openAPISpec))
}

// WriteOpenAPISpec writes the spec to disk (studio/openapi.json) so the
// frontend build can generate its TypeScript client without a running
// server.
func WriteOpenAPISpec(path string) error {
	return os.WriteFile(path, []byte(openAPISpec), 0o644)
}
