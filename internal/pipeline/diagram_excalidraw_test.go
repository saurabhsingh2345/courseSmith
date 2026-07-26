package pipeline

import (
	"strings"
	"testing"
)

func validScene() *ExcalidrawScene {
	return &ExcalidrawScene{
		Kind:   "excalidraw",
		Width:  excalidrawCanvasW,
		Height: excalidrawCanvasH,
		Elements: []ExcalidrawElement{
			{Type: "rectangle", X: 80, Y: 80, Width: 200, Height: 90, Label: "list", BackgroundColor: "#dbeafe"},
			{Type: "diamond", X: 80, Y: 260, Width: 200, Height: 120, Label: "empty?"},
			{Type: "arrow", X: 280, Y: 125, Points: [][2]float64{{0, 0}, {120, 0}}, Label: "next"},
			{Type: "line", X: 80, Y: 420, Points: [][2]float64{{0, 0}, {300, 0}}},
			{Type: "text", X: 80, Y: 460, Text: "caption", FontSize: 20},
		},
	}
}

func TestExcalidrawSceneValidateOK(t *testing.T) {
	if err := validScene().Validate(); err != nil {
		t.Fatalf("valid scene rejected: %v", err)
	}
}

func TestExcalidrawSceneValidateErrors(t *testing.T) {
	cases := []struct {
		name  string
		mut   func(s *ExcalidrawScene)
		field string
	}{
		{"no-size", func(s *ExcalidrawScene) { s.Width = 0 }, "width"},
		{"no-elements", func(s *ExcalidrawScene) { s.Elements = nil }, "at least one"},
		{"shape-no-dims", func(s *ExcalidrawScene) { s.Elements[0].Width = 0 }, "width and height"},
		{"line-few-points", func(s *ExcalidrawScene) { s.Elements[2].Points = [][2]float64{{0, 0}} }, "2 points"},
		{"text-empty", func(s *ExcalidrawScene) { s.Elements[4].Text = "" }, "text is required"},
		{"unknown-type", func(s *ExcalidrawScene) { s.Elements[0].Type = "star" }, "type must be one of"},
		{"bad-stroke", func(s *ExcalidrawScene) { s.Elements[0].StrokeColor = "red" }, "hex colour"},
		{"bad-fillstyle", func(s *ExcalidrawScene) { s.Elements[0].FillStyle = "zigzag" }, "fillStyle"},
		{"bad-roughness", func(s *ExcalidrawScene) { s.Elements[0].Roughness = 9 }, "roughness"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validScene()
			tc.mut(s)
			err := s.Validate()
			if err == nil {
				t.Fatalf("expected error mentioning %q, got nil", tc.field)
			}
			if !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("error %q does not mention %q", err.Error(), tc.field)
			}
		})
	}
}
