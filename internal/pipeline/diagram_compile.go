package pipeline

// Server-side diagram compilation. The "mermaid" and "excalidraw" diagram
// kinds let the content model author a structured *source* (Mermaid syntax or
// a small Excalidraw element list) instead of freehand SVG. This file turns
// that source into a self-contained SVG using the headless browser we already
// run for vision QA — so the compiled SVG then flows through the exact same
// screenshot → vision-review loop and the same inline-SVG renderer as the
// "svg" kind. Mermaid.js and Rough.js are embedded in the binary so
// compilation is offline and version-pinned (no CDN at runtime).

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

//go:embed assets/mermaid.min.js
var mermaidJS string

//go:embed assets/rough.js
var roughJS string

// DiagramCompiler compiles a structured diagram source to a self-contained
// SVG. RodScreenshotter implements it with headless Chromium; the visuals
// stage type-asserts its Screenshotter to this and degrades to freehand-SVG
// generation when no browser is available.
type DiagramCompiler interface {
	// RenderMermaid renders Mermaid syntax to an SVG string themed with the
	// course's video design tokens.
	RenderMermaid(ctx context.Context, syntax string, theme SceneTheme) ([]byte, error)
	// RenderExcalidraw renders an ExcalidrawScene (marshalled JSON) to an SVG
	// string using Rough.js, matching Excalidraw's hand-drawn look; default
	// ink comes from the theme so strokes read on the dark stage.
	RenderExcalidraw(ctx context.Context, sceneJSON []byte, theme SceneTheme) ([]byte, error)
}

// mermaidThemeVariables maps the course video theme onto Mermaid's `base`
// theme variables (the only customizable Mermaid theme). Every variable is
// set explicitly — Mermaid's derivation chain is otherwise non-deterministic
// across versions — so compiled diagrams sit natively on the dark stage:
// navy surfaces, accent borders, readable muted edges, transparent canvas.
func mermaidThemeVariables(t SceneTheme) map[string]any {
	surface2 := shiftLightness(t.Surface, 0.04)
	surface3 := shiftLightness(t.Surface, -0.04)
	return map[string]any{
		"darkMode":   true,
		"background": "transparent",
		"fontFamily": t.FontBody + ", sans-serif",
		// Diagrams render on a 1920x1080 stage and are upscaled to fill it —
		// a generous base size keeps labels crisp instead of blown-up-thin.
		"fontSize": "22px",

		// Nodes.
		"primaryColor":       t.Surface,
		"primaryTextColor":   t.Text,
		"primaryBorderColor": t.Accent,
		"nodeTextColor":      t.Text,
		"mainBkg":            t.Surface,
		"nodeBorder":         t.Accent,

		// Secondary/tertiary shapes (subgraph internals, alt shapes).
		"secondaryColor":       surface2,
		"secondaryTextColor":   t.Text,
		"secondaryBorderColor": t.SurfaceBorder,
		"tertiaryColor":        surface3,
		"tertiaryTextColor":    t.TextMuted,
		"tertiaryBorderColor":  t.SurfaceBorder,

		// Edges & labels.
		"lineColor":           t.TextMuted,
		"defaultLinkColor":    t.TextMuted,
		"edgeLabelBackground": t.BgTop,
		"textColor":           t.Text,
		"titleColor":          t.Accent,

		// Subgraphs/clusters.
		"clusterBkg":    surface3,
		"clusterBorder": t.SurfaceBorder,

		// Notes (sequence diagrams).
		"noteBkgColor":    surface2,
		"noteTextColor":   t.Text,
		"noteBorderColor": t.Accent,

		// Sequence diagrams.
		"actorBkg":            t.Surface,
		"actorBorder":         t.Accent,
		"actorTextColor":      t.Text,
		"actorLineColor":      t.TextMuted,
		"signalColor":         t.TextMuted,
		"signalTextColor":     t.Text,
		"labelBoxBkgColor":    surface2,
		"labelBoxBorderColor": t.Accent,
		"labelTextColor":      t.Text,
		"loopTextColor":       t.Text,
		"activationBkgColor":  surface2,
		"activationBorderColor": t.Accent,

		// State/class/pie extras that otherwise fall back to light defaults.
		"stateBkg":      t.Surface,
		"stateLabelColor": t.Text,
		"classText":     t.Text,
		"pie1":          t.Accent,
		"pie2":          shiftLightness(t.Accent, -0.12),
		"pie3":          surface2,
		"pie4":          t.SurfaceBorder,
		"pieTitleTextColor":   t.Text,
		"pieSectionTextColor": t.Text,
		"pieLegendTextColor":  t.Text,
		"pieStrokeColor":      t.BgTop,
		"pieOuterStrokeColor": t.SurfaceBorder,
	}
}

// injectScriptJS appends a <script> whose body is a bundled library and runs
// it at global scope (so a UMD bundle attaches its global, e.g. window.mermaid
// / window.rough). Returning a value keeps rod's round-trip simple.
const injectScriptJS = `(src) => {
	const s = document.createElement('script');
	s.textContent = src;
	(document.head || document.documentElement).appendChild(s);
	return true;
}`

// mermaidRenderJS initialises Mermaid with SVG (not HTML) labels — keeping the
// output self-contained and animatable — and renders one diagram, resolving to
// the SVG markup. mermaid.render rejects on invalid syntax; that rejection
// surfaces as the Go error the generation loop feeds back to the model.
// The base theme + injected themeVariables give course-branded dark output;
// useMaxWidth:false plus stripping the root inline style (Mermaid's
// max-width cap) lets the renderer scale the diagram up to fill the stage.
const mermaidRenderJS = `(syntax, varsJSON) => {
	window.mermaid.initialize({
		startOnLoad: false,
		htmlLabels: false,
		theme: 'base',
		themeVariables: JSON.parse(varsJSON),
		flowchart: {
			htmlLabels: false,
			useMaxWidth: false,
			curve: 'basis',
			nodeSpacing: 80,
			rankSpacing: 90,
			diagramPadding: 16,
			padding: 22,
			wrappingWidth: 230,
		},
		sequence: {useMaxWidth: false},
		securityLevel: 'strict',
	});
	return window.mermaid.render('coursesmith-diagram', syntax).then((r) => {
		const doc = new DOMParser().parseFromString(r.svg, 'image/svg+xml');
		const root = doc.documentElement;
		if (root.nodeName === 'parsererror') return r.svg;
		root.removeAttribute('style'); // kills mermaid's max-width cap

		// Polish pass. Mermaid's base theme draws 1px hairlines and sharp
		// corners — fine in a doc, flimsy on a 1080p video frame where the
		// diagram is upscaled to fill the stage. Round the node shapes,
		// thicken every stroke, and weight the labels so the diagram reads
		// as designed rather than defaulted.
		for (const rect of root.querySelectorAll('.node rect, rect.basic')) {
			rect.setAttribute('rx', '12');
			rect.setAttribute('ry', '12');
		}
		for (const shape of root.querySelectorAll('.node rect, .node circle, .node ellipse, .node polygon, .node path')) {
			shape.setAttribute('stroke-width', '2.5');
		}
		for (const edge of root.querySelectorAll('.edgePath path, path.flowchart-link')) {
			edge.setAttribute('stroke-width', '2.5');
		}
		for (const label of root.querySelectorAll('.nodeLabel, .node .label text, .node text')) {
			label.setAttribute('font-weight', '600');
		}

		// Mermaid's SVG-label path (htmlLabels: false) HTML-escapes label text
		// and then inserts it as a text node, so "a < b" reaches the screen as
		// the literal characters "a &lt; b". Decode entity sequences inside
		// text nodes; serialization re-escapes whatever actually needs it.
		const walker = doc.createTreeWalker(root, NodeFilter.SHOW_TEXT);
		const entities = {lt: '<', gt: '>', amp: '&', quot: '"', '#39': "'", nbsp: ' '};
		for (let n = walker.nextNode(); n; n = walker.nextNode()) {
			if (n.nodeValue && n.nodeValue.includes('&')) {
				n.nodeValue = n.nodeValue.replace(/&(lt|gt|amp|quot|#39|nbsp);/g, (_, e) => entities[e]);
			}
		}
		return new XMLSerializer().serializeToString(root);
	});
}`

// excalidrawRenderJS builds a hand-drawn SVG from an ExcalidrawScene using
// Rough.js — the same rendering engine Excalidraw itself uses. Each element is
// wrapped in its own top-level <g> so the renderer's group-stagger reveal
// animates the drawing element by element. Kept in sync with ExcalidrawScene
// in diagram_excalidraw.go.
const excalidrawRenderJS = `(sceneStr, inkColor) => {
	const scene = JSON.parse(sceneStr);
	const NS = 'http://www.w3.org/2000/svg';
	const svg = document.createElementNS(NS, 'svg');
	svg.setAttribute('xmlns', NS);
	svg.setAttribute('viewBox', '0 0 ' + scene.width + ' ' + scene.height);
	svg.setAttribute('width', scene.width);
	svg.setAttribute('height', scene.height);
	document.body.appendChild(svg);
	const rc = window.rough.svg(svg);
	// Default ink follows the (dark) video theme; near-black strokes the
	// model emits out of light-canvas habit are remapped so they stay
	// visible on the dark stage.
	const DARK = inkColor || '#e8ecf4';
	const remap = (c) => {
		if (!c) return DARK;
		const m = /^#([0-9a-f]{6})$/i.exec(String(c).trim());
		if (m) {
			const v = parseInt(m[1], 16);
			const lum = 0.2126 * ((v >> 16) & 255) + 0.7152 * ((v >> 8) & 255) + 0.0722 * (v & 255);
			if (lum < 64) return DARK;
		}
		return c;
	};

	for (const el of (scene.elements || [])) {
		const g = document.createElementNS(NS, 'g');
		const stroke = remap(el.strokeColor);
		const opts = {stroke: stroke, strokeWidth: el.strokeWidth || 2, roughness: (el.roughness == null ? 1 : el.roughness)};
		if (el.backgroundColor) { opts.fill = el.backgroundColor; opts.fillStyle = el.fillStyle || 'hachure'; }
		const pts = (el.points || []).map((p) => [el.x + p[0], el.y + p[1]]);
		let shape = null;
		switch (el.type) {
			case 'rectangle': shape = rc.rectangle(el.x, el.y, el.width, el.height, opts); break;
			case 'ellipse':   shape = rc.ellipse(el.x + el.width / 2, el.y + el.height / 2, el.width, el.height, opts); break;
			case 'diamond':   shape = rc.polygon([[el.x + el.width / 2, el.y], [el.x + el.width, el.y + el.height / 2], [el.x + el.width / 2, el.y + el.height], [el.x, el.y + el.height / 2]], opts); break;
			case 'line':      shape = rc.linearPath(pts, opts); break;
			case 'arrow': {
				shape = rc.linearPath(pts, opts);
				const [x2, y2] = pts[pts.length - 1];
				const [x1, y1] = pts[pts.length - 2];
				const a = Math.atan2(y2 - y1, x2 - x1);
				const L = 16;
				const head = rc.polygon([
					[x2, y2],
					[x2 - L * Math.cos(a - 0.45), y2 - L * Math.sin(a - 0.45)],
					[x2 - L * Math.cos(a + 0.45), y2 - L * Math.sin(a + 0.45)],
				], {stroke: stroke, fill: stroke, fillStyle: 'solid', roughness: opts.roughness});
				g.appendChild(shape);
				shape = head;
				break;
			}
		}
		if (shape) g.appendChild(shape);

		const label = el.type === 'text' ? el.text : el.label;
		if (label) {
			const fs = el.fontSize || (el.type === 'text' ? 20 : 16);
			const lines = String(label).split('\n');
			const isText = el.type === 'text';
			// Anchor the label: standalone text at its x,y; shapes at their
			// centre; lines/arrows at the midpoint of their path (they have no
			// width/height), lifted slightly above the stroke.
			let cx, cy, centred = true;
			if (isText) {
				cx = el.x; cy = el.y + fs; centred = false;
			} else if (pts.length >= 2) {
				const a = pts[0], b = pts[pts.length - 1];
				cx = (a[0] + b[0]) / 2; cy = (a[1] + b[1]) / 2 - 10;
			} else {
				cx = el.x + el.width / 2;
				cy = el.y + el.height / 2 - ((lines.length - 1) * fs * 1.2) / 2;
			}
			if (Number.isFinite(cx) && Number.isFinite(cy)) {
				lines.forEach((ln, i) => {
					const t = document.createElementNS(NS, 'text');
					t.setAttribute('x', cx);
					t.setAttribute('y', cy + i * fs * 1.2);
					t.setAttribute('font-family', "'Comic Sans MS', 'Segoe Print', 'Bradley Hand', cursive");
					t.setAttribute('font-size', fs);
					t.setAttribute('fill', stroke);
					if (centred) { t.setAttribute('text-anchor', 'middle'); t.setAttribute('dominant-baseline', 'central'); }
					t.textContent = ln;
					g.appendChild(t);
				});
			}
		}
		svg.appendChild(g);
	}
	const out = svg.outerHTML;
	svg.remove();
	return out;
}`

// compilePage opens a fresh blank page and injects a bundled library into it.
// The caller renders and then closes the page (returned closer).
func (r *RodScreenshotter) compilePage(ctx context.Context, lib string) (page *rod.Page, closer func(), err error) {
	browser, err := r.getBrowser()
	if err != nil {
		return nil, nil, err
	}
	p, err := browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return nil, nil, fmt.Errorf("opening compile page: %w", err)
	}
	closer = func() { _ = p.Close() }
	p = p.Timeout(45 * time.Second).Context(ctx)
	if err := p.WaitLoad(); err != nil {
		closer()
		return nil, nil, fmt.Errorf("loading compile page: %w", err)
	}
	if _, err := p.Eval(injectScriptJS, lib); err != nil {
		closer()
		return nil, nil, fmt.Errorf("injecting diagram library: %w", err)
	}
	return p, closer, nil
}

// RenderMermaid renders Mermaid syntax to a theme-branded SVG string.
func (r *RodScreenshotter) RenderMermaid(ctx context.Context, syntax string, theme SceneTheme) (svg []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("rendering mermaid: %v", rec)
		}
	}()
	vars, err := json.Marshal(mermaidThemeVariables(theme))
	if err != nil {
		return nil, fmt.Errorf("marshalling mermaid theme: %w", err)
	}
	page, closer, err := r.compilePage(ctx, mermaidJS)
	if err != nil {
		return nil, err
	}
	defer closer()
	obj, err := page.Eval(mermaidRenderJS, syntax, string(vars))
	if err != nil {
		return nil, fmt.Errorf("mermaid render failed: %w", err)
	}
	out := obj.Value.Str()
	if out == "" {
		return nil, fmt.Errorf("mermaid produced no SVG")
	}
	return []byte(out), nil
}

// RenderExcalidraw renders an ExcalidrawScene (marshalled JSON) to an SVG
// string using Rough.js.
func (r *RodScreenshotter) RenderExcalidraw(ctx context.Context, sceneJSON []byte, theme SceneTheme) (svg []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("rendering excalidraw: %v", rec)
		}
	}()
	page, closer, err := r.compilePage(ctx, roughJS)
	if err != nil {
		return nil, err
	}
	defer closer()
	obj, err := page.Eval(excalidrawRenderJS, string(sceneJSON), theme.Text)
	if err != nil {
		return nil, fmt.Errorf("excalidraw render failed: %w", err)
	}
	out := obj.Value.Str()
	if out == "" {
		return nil, fmt.Errorf("excalidraw produced no SVG")
	}
	return []byte(out), nil
}
