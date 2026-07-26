package pipeline

// conceptsSVG renders the course concept map: one column per lesson, a box
// per concept the lesson introduces, and an edge from each concept to
// every later lesson that requires it. Violations render in red.

import (
	"fmt"
	"github.com/enfec/coursesmith/internal/project"
	"html"
	"strings"
)

const (
	svgColWidth   = 240
	svgColGap     = 30
	svgBoxHeight  = 30
	svgBoxGap     = 10
	svgHeaderH    = 70
	svgMarginX    = 30
	svgMarginY    = 30
	svgEdgeSpace  = 60 // vertical space under the deepest column for edges
	svgBoxPadding = 8
)

func conceptsSVG(course *project.Course, report *AnalysisReport, perLesson map[string]*LessonConcepts) string {
	primary := course.Config.Branding.Colors.Primary
	if primary == "" {
		primary = "#2563eb"
	}
	accent := course.Config.Branding.Colors.Accent
	if accent == "" {
		accent = "#f59e0b"
	}

	colX := make(map[string]int, len(report.Lessons))
	for i, id := range report.Lessons {
		colX[id] = svgMarginX + i*(svgColWidth+svgColGap)
	}
	violated := map[string]bool{} // "concept|lesson" of violating usage
	unresolvedConcept := map[string]bool{}
	for _, v := range report.Violations {
		violated[v.Concept+"|"+v.UsedIn] = true
		unresolvedConcept[v.Concept] = true
	}

	// Box positions for introduced concepts, per lesson column.
	type box struct {
		concept string
		x, y    int
	}
	var boxes []box
	maxRows := 0
	introBox := map[string]box{}
	for _, id := range report.Lessons {
		lc := perLesson[id]
		if lc == nil {
			continue
		}
		row := 0
		seen := map[string]bool{}
		for _, ref := range lc.Introduced {
			if seen[ref.Name] {
				continue
			}
			seen[ref.Name] = true
			b := box{
				concept: ref.Name,
				x:       colX[id],
				y:       svgMarginY + svgHeaderH + row*(svgBoxHeight+svgBoxGap),
			}
			boxes = append(boxes, b)
			if _, ok := introBox[ref.Name]; !ok {
				introBox[ref.Name] = b
			}
			row++
		}
		if row > maxRows {
			maxRows = row
		}
	}

	width := svgMarginX*2 + len(report.Lessons)*(svgColWidth+svgColGap) - svgColGap
	if width < 400 {
		width = 400
	}
	height := svgMarginY*2 + svgHeaderH + maxRows*(svgBoxHeight+svgBoxGap) + svgEdgeSpace

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" font-family="system-ui, sans-serif">`+"\n", width, height)
	fmt.Fprintf(&b, `<style>
  .lesson { font-size: 14px; font-weight: 600; fill: #111827; }
  .lesson-sub { font-size: 11px; fill: #6b7280; }
  .concept { fill: #ffffff; stroke: %s; stroke-width: 1.5; }
  .concept.violated { stroke: #dc2626; stroke-width: 2.5; }
  .concept-label { font-size: 12px; fill: #111827; }
  .edge { stroke: %s; stroke-width: 1.2; fill: none; opacity: 0.55; }
  .edge.violated { stroke: #dc2626; opacity: 0.9; }
  .legend { font-size: 11px; fill: #6b7280; }
</style>
`, primary, accent)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="#ffffff"/>`+"\n", width, height)
	fmt.Fprintf(&b, `<text class="lesson" x="%d" y="%d" font-size="16">%s — concept map</text>`+"\n",
		svgMarginX, svgMarginY-8, html.EscapeString(course.Name))

	// Edges first (under the boxes): introduction box → each requiring
	// lesson's column header.
	for _, node := range report.Concepts {
		from, ok := introBox[node.Name]
		for _, usedIn := range node.RequiredBy {
			toX, okCol := colX[usedIn]
			if !okCol {
				continue
			}
			cls := "edge"
			if violated[node.Name+"|"+usedIn] {
				cls = "edge violated"
			}
			if !ok {
				// Never introduced: dangling edge from below the column.
				fmt.Fprintf(&b, `<path class="%s" stroke-dasharray="4 3" d="M %d %d L %d %d"/>`+"\n",
					cls, toX+svgColWidth/2, height-svgMarginY, toX+svgColWidth/2, svgMarginY+svgHeaderH-8)
				continue
			}
			fx := from.x + svgColWidth - svgBoxPadding
			fy := from.y + svgBoxHeight/2
			tx := toX + svgColWidth/2
			ty := svgMarginY + svgHeaderH - 10
			fmt.Fprintf(&b, `<path class="%s" d="M %d %d C %d %d, %d %d, %d %d"/>`+"\n",
				cls, fx, fy, fx+40, fy, tx, ty+30, tx, ty)
		}
	}

	// Lesson column headers.
	for _, id := range report.Lessons {
		x := colX[id]
		fmt.Fprintf(&b, `<text class="lesson" x="%d" y="%d">%s</text>`+"\n",
			x, svgMarginY+24, html.EscapeString(id))
		intro := 0
		if lc := perLesson[id]; lc != nil {
			intro = len(lc.Introduced)
		}
		fmt.Fprintf(&b, `<text class="lesson-sub" x="%d" y="%d">introduces %d concept(s)</text>`+"\n",
			x, svgMarginY+42, intro)
	}

	// Concept boxes.
	for _, bx := range boxes {
		cls := "concept"
		if unresolvedConcept[bx.concept] {
			cls = "concept violated"
		}
		fmt.Fprintf(&b, `<rect class="%s" x="%d" y="%d" width="%d" height="%d" rx="6"/>`+"\n",
			cls, bx.x, bx.y, svgColWidth-svgBoxPadding*2, svgBoxHeight)
		fmt.Fprintf(&b, `<text class="concept-label" x="%d" y="%d">%s</text>`+"\n",
			bx.x+svgBoxPadding, bx.y+svgBoxHeight/2+4, html.EscapeString(truncate(bx.concept, 30)))
	}

	fmt.Fprintf(&b, `<text class="legend" x="%d" y="%d">edges: concept → lessons that require it · red = used before taught</text>`+"\n",
		svgMarginX, height-10)
	b.WriteString("</svg>\n")
	return b.String()
}
